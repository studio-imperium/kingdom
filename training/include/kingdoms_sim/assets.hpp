#pragma once

#include <cstdint>
#include <string>
#include <unordered_map>
#include <vector>

#include "kingdoms_sim/types.hpp"

// Mirrors server/engine/assets.go. Field names/shapes match the Go structs
// 1:1 so the JSON in server/engine/assets/*.json (the live game's balance
// data) can be loaded without transformation. `animation` fields are dropped
// throughout: they're a client rendering hint, never read by gameplay logic.

namespace kingdoms {

struct Stats {
  float health = 1.0f;
  float regen = 1.0f;
  float speed = 1.0f;
  float damage = 1.0f;
  float reload = 1.0f;
};

struct ProjectileSpawnData {
  uint8_t id = 0;
  uint16_t angle = 0;  // relative offset angle, degrees
  float x = 0.0f;      // relative spawn offset
  float y = 0.0f;
};

struct BombSpawnData {
  uint8_t id = 0;
  float x = 0.0f;
  float y = 0.0f;
};

struct SummonData {
  uint8_t id = 0;
  float x = 0.0f;
  float y = 0.0f;
};

struct AttackData {
  float reload = 0.0f;
  float wait = 0.0f;
  std::vector<ProjectileSpawnData> projectiles;
  std::vector<BombSpawnData> bombs;
  std::vector<SummonData> summons;
};

struct NpcModeData {
  float duration = 0.0f;
  float max_health = 0.0f;
  float min_health = 0.0f;
  bool single_use = false;
  bool priority = false;
  MovementMode movement = MovementMode::Wander;
  float speed = 0.0f;
  std::vector<AttackData> attacks;
};

struct NpcData {
  uint8_t id = 0;
  std::string name;
  float health = 0.0f;
  uint16_t loot = 0;
  float range = 0.0f;
  float hitbox = 0.0f;
  std::vector<NpcModeData> modes;

  // Not in the Go source: a derived score used only by the training arena's
  // spawner to bucket NPCs into difficulty tiers (see arena.hpp). Computed
  // once at load time from health/damage/range so it stays in sync if
  // npcs.json changes, instead of hand-maintaining a tier list.
  float difficulty_score = 0.0f;
};

struct ItemData {
  uint8_t id = 0;
  ItemSlot slot = ItemSlot::Hand;
  Stats stats;
  std::string on_use;  // empty = not a consumable
  std::vector<AttackData> attacks;
};

struct ProjectileData {
  uint8_t id = 0;
  float speed = 0.0f;
  float range = 0.0f;
  float damage = 0.0f;
  bool piercing = false;
  float hitbox = 0.0f;
};

struct BombData {
  uint8_t id = 0;
  float airtime = 0.0f;
  float damage = 0.0f;
  float radius = 0.0f;
};

struct LootData {
  uint8_t loot = 0;
  float chance = 0.0f;
  bool soulbound = false;
};

// Loaded from the server's assets/*.json. Owns every table needed to
// replicate combat/inventory/loot; does NOT load spawns.json/tiles.json/
// maps/ since the training arena uses its own distance-tiered spawner
// instead of the real game's procedural biome/cell system (see arena.hpp).
class AssetDatabase {
 public:
  // Reads items.json, npcs.json, projectiles.json, bombs.json, loot.json
  // from `assets_dir` (see KINGDOMS_ASSETS_DIR, set by CMake to point at
  // server/engine/assets so balance data can't drift from the live game).
  static AssetDatabase LoadFromDirectory(const std::string& assets_dir);

  const ItemData& Item(uint8_t id) const { return items_.at(id); }
  const NpcData& Npc(uint8_t id) const { return npcs_.at(id); }
  const ProjectileData& Projectile(uint8_t id) const { return projectiles_.at(id); }
  const BombData& Bomb(uint8_t id) const { return bombs_.at(id); }
  const std::vector<LootData>& LootTable(uint16_t id) const { return loot_tables_.at(id); }

  const std::vector<NpcData>& AllNpcs() const { return npc_list_; }

 private:
  std::unordered_map<uint8_t, ItemData> items_;
  std::unordered_map<uint8_t, NpcData> npcs_;
  std::vector<NpcData> npc_list_;  // same objects as npcs_, kept in load order for the arena's tiering
  std::unordered_map<uint8_t, ProjectileData> projectiles_;
  std::unordered_map<uint8_t, BombData> bombs_;
  std::unordered_map<uint16_t, std::vector<LootData>> loot_tables_;
};

}  // namespace kingdoms
