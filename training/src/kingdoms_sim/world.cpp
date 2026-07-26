#include "kingdoms_sim/world.hpp"

#include <algorithm>
#include <cmath>

#include "kingdoms_sim/combat.hpp"
#include "kingdoms_sim/movement.hpp"
#include "kingdoms_sim/reward.hpp"

namespace kingdoms {

World::World(const AssetDatabase& assets, uint32_t seed) : assets_(assets), rng_(seed), arena_(assets) {
  Reset(seed);
}

void World::Reset(uint32_t seed) {
  rng_.seed(seed);
  player_ = MakeDefaultCharacter(assets_);
  npcs_.clear();
  projectiles_.clear();
  bombs_.clear();
  loot_.clear();
  next_entity_id_ = 1;
  best_gear_score_ = GearScore(assets_, player_.head, player_.body);

  for (const auto& [type_id, pos] : arena_.InitialSpawns(rng_)) {
    SpawnNpcAt(type_id, pos);
  }
}

void World::SpawnNpcAt(uint8_t type_id, Vec2 pos) {
  Npc npc = MakeNpc(assets_, type_id, pos);
  npc.entity_id = NextEntityId();
  npcs_[npc.entity_id] = npc;
}

void World::SpawnLoot(uint8_t item_id, Vec2 pos) {
  std::uniform_real_distribution<float> jitter(-0.5f, 0.5f);
  Loot l;
  l.entity_id = NextEntityId();
  l.item_id = item_id;
  l.pos = {pos.x + jitter(rng_), pos.y + jitter(rng_)};
  l.timer = 50.0f;
  loot_[l.entity_id] = l;
}

void World::ResolvePlayerAction(const Action& action, float dt_seconds, TickEvents& events) {
  const float mag = std::sqrt(action.move.x * action.move.x + action.move.y * action.move.y);
  if (mag > 1e-6f) {
    const float clamped = std::min(mag, 1.0f);
    const float nx = action.move.x / mag * clamped;
    const float ny = action.move.y / mag * clamped;
    player_.pos.x += nx * player_.speed * dt_seconds;
    player_.pos.y += ny * player_.speed * dt_seconds;
  }

  if (action.do_equip) {
    player_.ChangeInventory(assets_, action.equip_to, action.equip_from);
    // Only pays out for a NEW high-water mark this episode, not just any
    // change relative to whatever was worn a moment ago -- otherwise
    // equipping, unequipping, and re-equipping the same good item would pay
    // out every time.
    const float current = GearScore(assets_, player_.head, player_.body);
    if (current > best_gear_score_) {
      events.gear_score_delta += (current - best_gear_score_);
      best_gear_score_ = current;
    }
  }

  if (action.do_switch && action.switch_to_slot < kBackpackSlots) {
    player_.hand = action.switch_to_slot;
  }

  if (action.attack) {
    FirePlayerAttack(action.attack_direction, events);
  }
}

void World::FirePlayerAttack(Vec2 attack_direction, TickEvents& events) {
  uint8_t held_item_id = 0;
  if (auto it = player_.inventory.find(player_.hand); it != player_.inventory.end()) {
    held_item_id = it->second;
  }
  const ItemData& held = assets_.Item(held_item_id);

  // A held consumable is "used" by the same attack trigger, dispatched by
  // item type -- matches characterAttack in clients.go, which returns
  // immediately after UseItem without ever touching the projectile path.
  if (!held.on_use.empty()) {
    if (held.on_use == "potion_heal") {
      player_.health += 5.0f;
      player_.RemoveInventory(assets_, player_.hand);
    }
    return;
  }

  if (held.attacks.empty()) return;  // not a weapon and not a consumable: no-op, not a miss

  if (player_.attack_cooldown < -0.1f) player_.attack_counter = 0;
  if (player_.attack_cooldown > 0.1f) return;

  const AttackData& attack = held.attacks[player_.attack_counter % held.attacks.size()];
  player_.attack_cooldown = attack.reload / player_.reload;
  player_.attack_counter = static_cast<uint8_t>((player_.attack_counter + 1) % held.attacks.size());

  // Direction is policy-controlled (not auto-aimed) so it can choose which
  // enemy to prioritize. A near-zero direction means "no aim" -- treated the
  // same as a miss rather than firing somewhere arbitrary.
  const float mag = std::sqrt(attack_direction.x * attack_direction.x +
                               attack_direction.y * attack_direction.y);
  const bool has_direction = mag > 1e-6f;
  const Vec2 dir = has_direction ? Vec2{attack_direction.x / mag, attack_direction.y / mag}
                                  : Vec2{0.0f, 0.0f};

  for (const auto& spawn : attack.projectiles) {
    if (!has_direction) {
      events.attacks_missed++;
      continue;
    }
    const float damage = assets_.Projectile(spawn.id).damage * player_.power;
    Projectile p = MakeDirectedProjectile(assets_, spawn, player_.pos, dir, false, damage);
    p.entity_id = NextEntityId();
    projectiles_[p.entity_id] = p;
  }

  for (const auto& spawn : attack.bombs) {
    if (!has_direction) {
      events.attacks_missed++;
      continue;
    }
    const float damage = assets_.Bomb(spawn.id).damage;
    Bomb b = MakeDirectedBomb(assets_, spawn, player_.pos, dir, false, damage);
    b.entity_id = NextEntityId();
    bombs_[b.entity_id] = b;
  }
}

void World::TickNpcs(float dt_seconds) {
  std::vector<std::pair<uint8_t, Vec2>> pending_summons;

  for (auto& [id, npc] : npcs_) {
    if (npc.dead) continue;

    const float range = assets_.Npc(npc.type_id).range;
    const bool in_range = !player_.dead && Distance(npc.pos, player_.pos) < range;
    if (in_range) {
      npc.targeting_player = true;
    } else if (npc.targeting_player) {
      npc.targeting_player = false;
    }

    npc.mode_timer -= dt_seconds;
    npc.attack_timer -= dt_seconds;

    if (npc.InCombat()) {
      if (!npc.ValidMode(assets_, npc.mode) || npc.mode_timer <= 0.0f) {
        npc.used_modes[npc.mode] = true;
        npc.NewMode(assets_, rng_);
      }

      const bool hover_gated =
          npc.movement == MovementMode::Hover && !IsHovering(npc, assets_, player_.pos);
      if (!hover_gated && npc.attack_timer < 0.0f) {
        FireNpcAttack(npc, pending_summons);
      }
    } else {
      npc.mode_timer = 0.0f;
    }

    // Always runs, matching engine.go's Run() loop calling Move() after Tick()
    // regardless of the hover-gate (which only suppresses firing, not moving).
    MoveNpc(npc, assets_, dt_seconds, player_.pos, rng_);
  }

  // Deferred: inserting into npcs_ while range-iterating it above would risk
  // invalidating the iterator on a rehash.
  for (const auto& [type_id, pos] : pending_summons) {
    SpawnNpcAt(type_id, pos);
  }
}

void World::FireNpcAttack(Npc& npc, std::vector<std::pair<uint8_t, Vec2>>& pending_summons) {
  if (!npc.CanAttack(assets_, player_.pos)) return;
  const AttackData* attack = npc.CurrentAttack(assets_);
  if (!attack) return;

  for (const auto& spawn : attack->projectiles) {
    const float damage = assets_.Projectile(spawn.id).damage;
    Projectile p = MakeAimedProjectile(assets_, spawn, npc.pos, player_.pos, true, damage);
    p.entity_id = NextEntityId();
    projectiles_[p.entity_id] = p;
  }
  for (const auto& spawn : attack->bombs) {
    const float damage = assets_.Bomb(spawn.id).damage;
    Bomb b = MakeAimedBomb(assets_, spawn, player_.pos, true, damage);
    b.entity_id = NextEntityId();
    bombs_[b.entity_id] = b;
  }
  for (const auto& summon : attack->summons) {
    pending_summons.emplace_back(summon.id, Vec2{npc.pos.x + summon.x, npc.pos.y + summon.y});
  }

  npc.attack += 1;
  npc.attack_timer = attack->reload + attack->wait;
}

void World::ApplyDamageToNpc(Npc& npc, float amount, bool from_player, TickEvents& events) {
  npc.health -= amount;
  if (from_player) {
    events.damage_dealt += amount;
    npc.damage_from_player += amount;
  }
  if (npc.health <= 0.0f && !npc.dead) {
    HandleNpcDeath(npc);
    if (from_player) events.kills += 1;
  }
}

void World::HandleNpcDeath(Npc& npc) {
  npc.dead = true;

  const NpcData& data = assets_.Npc(npc.type_id);
  const auto& loot_pool = assets_.LootTable(data.loot);
  const float sb_threshold = std::min(200.0f, data.health / 10.0f);
  std::uniform_real_distribution<float> chance_dist(0.0f, 1.0f);

  for (const auto& entry : loot_pool) {
    if (chance_dist(rng_) > entry.chance) continue;
    if (entry.soulbound) {
      if (npc.damage_from_player >= sb_threshold) SpawnLoot(entry.loot, npc.pos);
    } else {
      SpawnLoot(entry.loot, npc.pos);
    }
  }
}

void World::TickProjectiles(float dt_seconds, TickEvents& events) {
  for (auto& [id, proj] : projectiles_) {
    if (proj.dead) continue;

    const ProjectileData& data = assets_.Projectile(proj.type_id);
    proj.Tick(dt_seconds, data.speed);

    // Out-of-range is checked before hit-testing, matching simulation.go: a
    // projectile that crosses its range this tick expires without getting a
    // final hit-test at its new position.
    if (Distance(proj.pos, proj.origin) > data.range) {
      if (!proj.evil && proj.hitlist.empty()) events.attacks_missed++;
      proj.dead = true;
      continue;
    }

    if (proj.evil) {
      if (!player_.dead && !proj.hitlist.count(kPlayerEntityId) &&
          HitboxesIntersect(proj.pos, data.hitbox, player_.pos, kPlayerHitbox)) {
        proj.hitlist.insert(kPlayerEntityId);
        player_.Damage(proj.damage);
        events.damage_taken += proj.damage;
        if (!data.piercing) proj.dead = true;
      }
    } else {
      // Checks every NPC (not just the first hit), matching simulation.go's
      // deferred hit-processing: a non-piercing projectile can still land on
      // more than one simultaneously-overlapping target in the same tick.
      bool hit_something = false;
      for (auto& [npc_id, npc] : npcs_) {
        if (npc.dead || proj.hitlist.count(npc_id)) continue;
        if (HitboxesIntersect(proj.pos, data.hitbox, npc.pos, assets_.Npc(npc.type_id).hitbox)) {
          proj.hitlist.insert(npc_id);
          ApplyDamageToNpc(npc, proj.damage, true, events);
          hit_something = true;
        }
      }
      if (hit_something && !data.piercing) proj.dead = true;
    }
  }
}

void World::TickBombs(float dt_seconds, TickEvents& events) {
  for (auto& [id, bomb] : bombs_) {
    if (bomb.dead) continue;
    bomb.Tick(dt_seconds);
    if (bomb.timer > 0.0f) continue;

    bomb.dead = true;
    const float radius = assets_.Bomb(bomb.type_id).radius;
    bool hit_something = false;

    if (bomb.evil) {
      if (!player_.dead && WithinBombRadius(bomb.pos, radius, player_.pos, kPlayerHitbox)) {
        player_.Damage(bomb.damage);
        events.damage_taken += bomb.damage;
        hit_something = true;
      }
    } else {
      for (auto& [npc_id, npc] : npcs_) {
        if (npc.dead) continue;
        if (WithinBombRadius(bomb.pos, radius, npc.pos, assets_.Npc(npc.type_id).hitbox)) {
          ApplyDamageToNpc(npc, bomb.damage, true, events);
          hit_something = true;
        }
      }
      if (!hit_something) events.attacks_missed++;
    }
  }
}

void World::TickLoot(float dt_seconds, TickEvents& events) {
  for (auto& [id, l] : loot_) {
    if (l.dead) continue;
    l.Tick(dt_seconds);
    if (l.timer <= 0.0f) continue;  // erased in Step()'s cleanup pass

    if (!player_.dead && Distance(player_.pos, l.pos) < 1.0f && player_.AddItemOrBust(l.item_id)) {
      l.dead = true;
      player_.ApplyGear(assets_);
      events.loot_picked_up.push_back(l.item_id);
    }
  }
}

TickEvents World::Step(const Action& action) {
  TickEvents events;
  constexpr float dt = kTickSeconds;

  player_.Tick(dt);
  ResolvePlayerAction(action, dt, events);
  TickNpcs(dt);
  TickProjectiles(dt, events);
  TickBombs(dt, events);
  TickLoot(dt, events);

  std::vector<Vec2> alive_positions;
  alive_positions.reserve(npcs_.size());
  for (const auto& [id, npc] : npcs_) {
    if (!npc.dead) alive_positions.push_back(npc.pos);
  }
  for (const auto& [type_id, pos] : arena_.MaybeRespawn(alive_positions, dt, rng_)) {
    SpawnNpcAt(type_id, pos);
  }

  std::erase_if(npcs_, [](const auto& kv) { return kv.second.dead; });
  std::erase_if(projectiles_, [](const auto& kv) { return kv.second.dead; });
  std::erase_if(bombs_, [](const auto& kv) { return kv.second.dead; });
  std::erase_if(loot_, [](const auto& kv) { return kv.second.dead || kv.second.timer <= 0.0f; });

  if (player_.health < 1.0f && !player_.dead) {
    player_.dead = true;
    events.player_died = true;
  }

  return events;
}

}  // namespace kingdoms
