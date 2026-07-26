#include "kingdoms_sim/assets.hpp"

#include <fstream>

#include <nlohmann/json.hpp>

namespace kingdoms {

namespace {

using json = nlohmann::json;

json LoadJson(const std::string& path) {
  std::ifstream in(path);
  if (!in) {
    throw std::runtime_error("kingdoms_sim: could not open asset file: " + path);
  }
  json j;
  in >> j;
  return j;
}

ItemSlot ParseItemSlot(const std::string& s) {
  if (s == "head") return ItemSlot::Head;
  if (s == "body") return ItemSlot::Body;
  return ItemSlot::Hand;
}

MovementMode ParseMovementMode(const std::string& s) {
  if (s == "chase") return MovementMode::Chase;
  if (s == "run") return MovementMode::Run;
  if (s == "overshoot") return MovementMode::Overshoot;
  if (s == "hover") return MovementMode::Hover;
  if (s == "turret") return MovementMode::Turret;
  return MovementMode::Wander;
}

// Go's Stats accessors run every value through NonZero(), treating an
// explicit 0 the same as an absent field (1.0 = no modifier). Applied here
// at load time instead of on every access.
float NonZero(float v) { return v != 0.0f ? v : 1.0f; }

Stats ParseStats(const json& j) {
  Stats s;
  s.health = NonZero(j.value("health", 1.0f));
  s.regen = NonZero(j.value("regen", 1.0f));
  s.speed = NonZero(j.value("speed", 1.0f));
  s.damage = NonZero(j.value("damage", 1.0f));
  s.reload = NonZero(j.value("reload", 1.0f));
  return s;
}

std::vector<ProjectileSpawnData> ParseProjectileSpawns(const json& j) {
  std::vector<ProjectileSpawnData> out;
  for (const auto& p : j) {
    out.push_back(ProjectileSpawnData{
        p.value<uint8_t>("id", 0), p.value<uint16_t>("angle", 0), p.value("x", 0.0f), p.value("y", 0.0f)});
  }
  return out;
}

std::vector<BombSpawnData> ParseBombSpawns(const json& j) {
  std::vector<BombSpawnData> out;
  for (const auto& b : j) {
    out.push_back(BombSpawnData{b.value<uint8_t>("id", 0), b.value("x", 0.0f), b.value("y", 0.0f)});
  }
  return out;
}

std::vector<SummonData> ParseSummons(const json& j) {
  std::vector<SummonData> out;
  for (const auto& s : j) {
    out.push_back(SummonData{s.value<uint8_t>("id", 0), s.value("x", 0.0f), s.value("y", 0.0f)});
  }
  return out;
}

std::vector<AttackData> ParseAttacks(const json& j) {
  std::vector<AttackData> out;
  for (const auto& a : j) {
    AttackData attack;
    attack.reload = a.value("reload", 0.0f);
    attack.wait = a.value("wait", 0.0f);
    if (a.contains("projectiles")) attack.projectiles = ParseProjectileSpawns(a.at("projectiles"));
    if (a.contains("bombs")) attack.bombs = ParseBombSpawns(a.at("bombs"));
    if (a.contains("summons")) attack.summons = ParseSummons(a.at("summons"));
    out.push_back(std::move(attack));
  }
  return out;
}

}  // namespace

AssetDatabase AssetDatabase::LoadFromDirectory(const std::string& assets_dir) {
  AssetDatabase db;

  for (const auto& j : LoadJson(assets_dir + "/items.json")) {
    ItemData item;
    item.id = j.value("id", 0);
    item.slot = ParseItemSlot(j.value("type", std::string("hand")));
    if (j.contains("stats")) item.stats = ParseStats(j.at("stats"));
    item.on_use = j.value("on_use", std::string());
    if (j.contains("attacks")) item.attacks = ParseAttacks(j.at("attacks"));
    db.items_[item.id] = std::move(item);
  }

  for (const auto& j : LoadJson(assets_dir + "/projectiles.json")) {
    ProjectileData proj;
    proj.id = j.value("id", 0);
    proj.speed = j.value("speed", 0.0f);
    proj.range = j.value("range", 0.0f);
    proj.damage = j.value("damage", 0.0f);
    proj.piercing = j.value("piercing", false);
    proj.hitbox = j.value("hitbox", 0.0f);
    db.projectiles_[proj.id] = proj;
  }

  for (const auto& j : LoadJson(assets_dir + "/bombs.json")) {
    BombData bomb;
    bomb.id = j.value("id", 0);
    bomb.airtime = j.value("airtime", 0.0f);
    bomb.damage = j.value("damage", 0.0f);
    bomb.radius = j.value("radius", 0.0f);
    db.bombs_[bomb.id] = bomb;
  }

  for (const auto& j : LoadJson(assets_dir + "/loot.json")) {
    std::vector<LootData> entries;
    for (const auto& e : j.at("entries")) {
      entries.push_back(LootData{
          e.value<uint8_t>("loot", 0), e.value("chance", 0.0f), e.value("soulbound", false)});
    }
    db.loot_tables_[static_cast<uint16_t>(db.loot_tables_.size())] = std::move(entries);
  }

  for (const auto& j : LoadJson(assets_dir + "/npcs.json")) {
    NpcData npc;
    npc.id = j.value("id", 0);
    npc.name = j.value("display", std::string());
    npc.health = j.value("health", 0.0f);
    npc.loot = j.value("loot", 0);
    npc.range = j.value("range", 0.0f);
    npc.hitbox = j.value("hitbox", 0.0f);

    if (j.contains("modes")) {
      for (const auto& m : j.at("modes")) {
        NpcModeData mode;
        mode.duration = m.value("duration", 0.0f);
        mode.max_health = m.value("max_health", 0.0f);
        mode.min_health = m.value("min_health", 0.0f);
        mode.single_use = m.value("single_use", false);
        mode.priority = m.value("priority", false);
        mode.movement = ParseMovementMode(m.value("movement", std::string("wander")));
        mode.speed = m.value("speed", 0.0f);
        if (m.contains("attacks")) mode.attacks = ParseAttacks(m.at("attacks"));
        npc.modes.push_back(std::move(mode));
      }
    }

    // Placeholder difficulty heuristic for the arena spawner (arena.hpp):
    // health + aggro range + 20x the hardest single hit this NPC can land.
    // Tune freely once real training data shows what actually correlates
    // with how dangerous an NPC is.
    float max_hit = 0.0f;
    for (const auto& mode : npc.modes) {
      for (const auto& attack : mode.attacks) {
        for (const auto& p : attack.projectiles) {
          auto it = db.projectiles_.find(p.id);
          if (it != db.projectiles_.end()) max_hit = std::max(max_hit, it->second.damage);
        }
        for (const auto& b : attack.bombs) {
          auto it = db.bombs_.find(b.id);
          if (it != db.bombs_.end()) max_hit = std::max(max_hit, it->second.damage);
        }
      }
    }
    npc.difficulty_score = npc.health + npc.range + max_hit * 20.0f;

    db.npcs_[npc.id] = npc;
    db.npc_list_.push_back(std::move(npc));
  }

  return db;
}

}  // namespace kingdoms
