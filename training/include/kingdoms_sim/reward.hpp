#pragma once

#include <cstdint>

#include "kingdoms_sim/assets.hpp"
#include "kingdoms_sim/world.hpp"

namespace kingdoms {

// Sum of (stat - 1.0) across the 5 multiplier fields for the equipped head +
// body items. All 5 gear stats are "higher is better" for the player (speed/
// power/reload/health/regen all directly help), so this is a simple additive
// placeholder for "how good is this loadout" -- used both for the optional
// per-tick equip-upgrade shaping term and the terminal "gear collected"
// score. Tune per-stat weights once real training runs show what matters.
float GearScore(const AssetDatabase& assets, uint8_t head_item, uint8_t body_item);

// All magnitudes below are starting placeholders, not balanced numbers.
struct RewardWeights {
  float damage_dealt = 1.0f;       // per point of damage dealt to any NPC
  float damage_taken = -1.0f;      // per point of damage taken
  float kill = 5.0f;                // per NPC killed
  float loot_pickup = 2.0f;         // per item picked up
  float gear_upgrade = 3.0f;        // per point of GearScore improvement from an equip
  float attack_miss = 0.0f;         // per attack that resolved without hitting anything.
                                    // Disabled: at -0.5 a random policy accrues ~-700/episode
                                    // from misses, making early death outscore survival.
                                    // Revisit as shaping once a policy trains at all.
  float death = -100.0f;            // terminal, on top of the above -- dominates everything else
};

float ComputeStepReward(const TickEvents& events, const AssetDatabase& assets,
                         const RewardWeights& weights = RewardWeights{});

}  // namespace kingdoms
