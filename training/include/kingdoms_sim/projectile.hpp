#pragma once

#include <cstdint>
#include <unordered_set>

#include "kingdoms_sim/types.hpp"

// Mirrors server/engine/projectile.go.

namespace kingdoms {

struct Projectile {
  uint32_t entity_id = 0;
  uint8_t type_id = 0;  // indexes AssetDatabase::Projectile()
  bool evil = false;    // true = NPC-fired (threatens the player), false = player-fired
  float damage = 0.0f;
  Vec2 pos;
  Vec2 origin;
  float angle_degrees = 0.0f;  // internal only: travel direction, never observed/chosen by the policy

  // Entities already hit, so piercing projectiles don't hit the same target twice.
  std::unordered_set<uint32_t> hitlist;
  bool dead = false;

  void Tick(float dt_seconds, float speed);
};

}  // namespace kingdoms
