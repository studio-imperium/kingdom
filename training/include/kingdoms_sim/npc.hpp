#pragma once

#include <cstdint>
#include <random>
#include <vector>

#include "kingdoms_sim/assets.hpp"
#include "kingdoms_sim/types.hpp"

// Mirrors server/engine/npc.go, minus the nearby/EnterView/ExitView view-
// culling machinery (that existed only to support the real game's per-client
// render-distance simulation; see the training README for why the arena
// just ticks every NPC unconditionally instead).

namespace kingdoms {

struct Npc {
  uint32_t entity_id = 0;
  uint8_t type_id = 0;  // indexes AssetDatabase::Npc()

  Vec2 pos;
  Vec2 origin;  // spawn point; wander/chase roam relative to this
  float health = 0.0f;

  // Go's `target` is an Object interface (either the Character being fought,
  // or a wander/roam Point). Single-agent training only ever has one possible
  // Character target, so that collapses to a bool + the roam point itself.
  bool targeting_player = false;
  Vec2 movement_point;

  uint8_t mode = 0;
  std::vector<bool> used_modes;
  float mode_timer = 0.0f;
  MovementMode movement = MovementMode::Wander;

  uint8_t attack = 0;
  float attack_timer = 0.0f;

  // Damage the player has personally dealt to this NPC, for the soulbound
  // loot threshold check on death (Death() in npc.go). Collapses Go's
  // map[characterID]float32 to a scalar since there's only one player.
  float damage_from_player = 0.0f;

  bool dead = false;

  bool InCombat() const { return targeting_player; }
  bool ValidMode(const AssetDatabase& assets, uint8_t idx) const;
  void NewMode(const AssetDatabase& assets, std::mt19937& rng);
  const AttackData* CurrentAttack(const AssetDatabase& assets) const;
  bool CanAttack(const AssetDatabase& assets, Vec2 target_pos) const;
};

Npc MakeNpc(const AssetDatabase& assets, uint8_t type_id, Vec2 pos);

}  // namespace kingdoms
