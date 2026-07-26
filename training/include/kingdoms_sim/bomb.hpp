#pragma once

#include <cstdint>

#include "kingdoms_sim/types.hpp"

// Mirrors server/engine/bomb.go.

namespace kingdoms {

struct Bomb {
  uint32_t entity_id = 0;
  uint8_t type_id = 0;  // indexes AssetDatabase::Bomb()
  bool evil = false;
  float damage = 0.0f;
  Vec2 pos;
  Vec2 origin;
  float timer = 0.0f;  // airtime remaining; explodes at <= 0
  bool dead = false;

  void Tick(float dt_seconds) { timer -= dt_seconds; }
};

}  // namespace kingdoms
