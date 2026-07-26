#pragma once

#include <cstdint>
#include <random>
#include <unordered_map>
#include <vector>

#include "kingdoms_sim/action.hpp"
#include "kingdoms_sim/arena.hpp"
#include "kingdoms_sim/assets.hpp"
#include "kingdoms_sim/bomb.hpp"
#include "kingdoms_sim/character.hpp"
#include "kingdoms_sim/loot.hpp"
#include "kingdoms_sim/npc.hpp"
#include "kingdoms_sim/projectile.hpp"
#include "kingdoms_sim/types.hpp"

// Single-agent authoritative world. Merges engine.go's main tick (NPC AI,
// projectile/bomb/loot updates) with the hit-resolution logic that used to
// live in the per-client simulation.go -- see the module comment on
// combat.hpp. No mutexes/channels/goroutines: single-threaded, single
// player, and stepped explicitly rather than ticked against a wall clock.

namespace kingdoms {

// Reward-relevant events gathered during one World::Step call. Kept separate
// from reward weighting (reward.hpp) so tuning reward magnitudes never
// requires touching simulation code.
struct TickEvents {
  float damage_dealt = 0.0f;   // to any NPC, this tick
  float damage_taken = 0.0f;   // by the player, this tick
  int kills = 0;
  std::vector<uint8_t> loot_picked_up;  // item ids picked up this tick
  float gear_score_delta = 0.0f;        // change in equipped head+body gear score from an equip this tick
  int attacks_missed = 0;               // player projectiles/bombs that resolved without hitting anything
  bool player_died = false;
};

class World {
 public:
  explicit World(const AssetDatabase& assets, uint32_t seed);

  void Reset(uint32_t seed);
  TickEvents Step(const Action& action);

  const Character& player() const { return player_; }
  const std::unordered_map<uint32_t, Npc>& npcs() const { return npcs_; }
  const std::unordered_map<uint32_t, Projectile>& projectiles() const { return projectiles_; }
  const std::unordered_map<uint32_t, Bomb>& bombs() const { return bombs_; }
  const std::unordered_map<uint32_t, Loot>& loot() const { return loot_; }
  const AssetDatabase& assets() const { return assets_; }

  // Highest GearScore(head, body) reached so far this episode. Deliberately
  // NOT "whatever's equipped right now" -- see ResolvePlayerAction, this is
  // what the gear_upgrade reward term is measured against.
  float best_gear_score() const { return best_gear_score_; }

 private:
  uint32_t NextEntityId() { return next_entity_id_++; }
  void SpawnNpcAt(uint8_t type_id, Vec2 pos);
  void SpawnLoot(uint8_t item_id, Vec2 pos);

  void ResolvePlayerAction(const Action& action, float dt_seconds, TickEvents& events);
  void FirePlayerAttack(Vec2 attack_direction, TickEvents& events);

  void TickNpcs(float dt_seconds);
  void FireNpcAttack(Npc& npc, std::vector<std::pair<uint8_t, Vec2>>& pending_summons);

  void TickProjectiles(float dt_seconds, TickEvents& events);
  void TickBombs(float dt_seconds, TickEvents& events);
  void TickLoot(float dt_seconds, TickEvents& events);

  void ApplyDamageToNpc(Npc& npc, float amount, bool from_player, TickEvents& events);
  void HandleNpcDeath(Npc& npc);

  const AssetDatabase& assets_;
  std::mt19937 rng_;
  ArenaSpawner arena_;

  Character player_;
  std::unordered_map<uint32_t, Npc> npcs_;
  std::unordered_map<uint32_t, Projectile> projectiles_;
  std::unordered_map<uint32_t, Bomb> bombs_;
  std::unordered_map<uint32_t, Loot> loot_;
  uint32_t next_entity_id_ = 1;
  float best_gear_score_ = 0.0f;
};

}  // namespace kingdoms
