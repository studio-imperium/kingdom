#pragma once

#include <cstdint>

#include "kingdoms_sim/types.hpp"

// Mirrors server/engine/loot.go.

namespace kingdoms {

struct Loot {
  uint32_t entity_id = 0;
  uint8_t item_id = 0;
  Vec2 pos;
  float timer = 50.0f;  // despawn timer in seconds, matches CreateLoot in loot.go
  bool dead = false;

  void Tick(float dt_seconds) { timer -= dt_seconds; }
};

}  // namespace kingdoms
