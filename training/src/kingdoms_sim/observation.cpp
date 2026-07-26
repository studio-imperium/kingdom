#include "kingdoms_sim/observation.hpp"

#include <algorithm>
#include <unordered_map>
#include <vector>

namespace kingdoms {

namespace {

// Collects the alive entries of `source`, sorts nearest-to-`from` first,
// writes up to `max_slots` of them into `entities` starting at `slot`
// (advancing it), and adds however many didn't fit to `truncated_count`.
// `write` maps one source entry onto its ObservedEntity fields; the type tag
// and distance-sort/cap mechanics are identical across NPCs/projectiles/
// bombs/loot, only the field mapping differs per type.
template <typename T, typename WriteFn>
void FillNearest(const std::unordered_map<uint32_t, T>& source, Vec2 from, int max_slots,
                  std::array<ObservedEntity, kMaxObservedEntities>& entities, int& slot,
                  uint16_t& truncated_count, WriteFn write) {
  std::vector<const T*> alive;
  alive.reserve(source.size());
  for (const auto& [id, item] : source) {
    if (!item.dead) alive.push_back(&item);
  }
  std::sort(alive.begin(), alive.end(), [from](const T* a, const T* b) {
    return DistanceSq(from, a->pos) < DistanceSq(from, b->pos);
  });

  const int fit = std::min<int>(static_cast<int>(alive.size()), max_slots);
  truncated_count += static_cast<uint16_t>(static_cast<int>(alive.size()) - fit);
  for (int i = 0; i < fit; ++i) {
    write(entities[slot++], *alive[i]);
  }
}

}  // namespace

Observation BuildObservation(const World& world) {
  Observation obs;
  const Character& p = world.player();

  obs.player_pos = p.pos;
  obs.health = p.health;
  obs.max_health = p.max_health;
  obs.attack_cooldown = p.attack_cooldown;
  obs.speed = p.speed;
  obs.power = p.power;
  obs.reload = p.reload;
  obs.regen = p.regen;
  obs.hand = p.hand;
  obs.head = p.head;
  obs.body = p.body;

  obs.inventory.fill(-1);
  for (uint8_t i = 0; i < kBackpackSlots; ++i) {
    if (auto it = p.inventory.find(i); it != p.inventory.end()) {
      obs.inventory[i] = static_cast<int16_t>(it->second);
    }
  }

  int slot = 0;

  FillNearest(world.npcs(), p.pos, kMaxNpcSlots, obs.entities, slot, obs.truncated_count,
              [](ObservedEntity& e, const Npc& npc) {
                e.type = ObservedEntityType::Npc;
                e.pos = npc.pos;
                e.npc_health = npc.health;
                e.npc_type_id = npc.type_id;
              });

  FillNearest(world.projectiles(), p.pos, kMaxProjectileSlots, obs.entities, slot, obs.truncated_count,
              [](ObservedEntity& e, const Projectile& proj) {
                e.type = ObservedEntityType::Projectile;
                e.pos = proj.pos;
                e.projectile_damage = proj.damage;
                e.is_enemy_projectile = proj.evil;
              });

  FillNearest(world.bombs(), p.pos, kMaxBombSlots, obs.entities, slot, obs.truncated_count,
              [&world](ObservedEntity& e, const Bomb& bomb) {
                e.type = ObservedEntityType::Bomb;
                e.pos = bomb.pos;
                e.bomb_damage = bomb.damage;
                e.bomb_radius = world.assets().Bomb(bomb.type_id).radius;
                e.bomb_timer = bomb.timer;
              });

  FillNearest(world.loot(), p.pos, kMaxLootSlots, obs.entities, slot, obs.truncated_count,
              [](ObservedEntity& e, const Loot& l) {
                e.type = ObservedEntityType::Loot;
                e.pos = l.pos;
                e.loot_item_id = l.item_id;
              });

  return obs;
}

}  // namespace kingdoms
