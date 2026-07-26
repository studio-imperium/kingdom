#pragma once

#include <cstdint>
#include <random>
#include <utility>
#include <vector>

#include "kingdoms_sim/assets.hpp"
#include "kingdoms_sim/types.hpp"

// Replaces the real game's procedural biome/cell spawn system (map.go,
// spawns.json) with a simple arena: NPCs are bucketed into difficulty tiers
// by their own stats (see AssetDatabase::difficulty_score), and spawned in
// concentric distance bands around the player's origin -- roughly
// approximating "tougher enemies further from spawn" without needing the
// real biome/tile system, which the training sim doesn't model at all.
//
// All distances/populations below are starting placeholders, not balanced
// numbers -- tune freely once training results say what actually matters.

namespace kingdoms {

class ArenaSpawner {
 public:
  explicit ArenaSpawner(const AssetDatabase& assets);

  std::vector<std::pair<uint8_t, Vec2>> InitialSpawns(std::mt19937& rng);

  // Call once per Step. Internally throttles the actual check to once every
  // few seconds of sim time; returns newly-spawned (type_id, pos) pairs, or
  // empty most calls.
  std::vector<std::pair<uint8_t, Vec2>> MaybeRespawn(const std::vector<Vec2>& alive_npc_positions,
                                                      float dt_seconds, std::mt19937& rng);

 private:
  struct Tier {
    float min_distance;
    float max_distance;
    int target_population;
    std::vector<uint8_t> npc_type_ids;
  };

  int TierIndexOf(Vec2 pos) const;
  Vec2 RandomPosInTier(const Tier& tier, std::mt19937& rng) const;

  std::vector<Tier> tiers_;
  float respawn_check_timer_ = 0.0f;
};

}  // namespace kingdoms
