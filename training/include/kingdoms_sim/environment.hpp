#pragma once

#include <cstdint>

#include "kingdoms_sim/action.hpp"
#include "kingdoms_sim/assets.hpp"
#include "kingdoms_sim/observation.hpp"
#include "kingdoms_sim/reward.hpp"
#include "kingdoms_sim/world.hpp"

// The RL-facing API. Deliberately not wired to Python/pybind11 or any
// specific training framework yet -- see the README for why that's a
// deferred decision rather than a blocking one.

namespace kingdoms {

struct StepResult {
  Observation observation;
  float reward = 0.0f;
  bool done = false;
  TickEvents events;  // raw per-tick counters, for logging/debugging outside the reward scalar
};

class Environment {
 public:
  static constexpr int kDefaultMaxTicks = 20000;

  explicit Environment(const AssetDatabase& assets, uint32_t seed,
                        RewardWeights weights = RewardWeights{},
                        int max_ticks = kDefaultMaxTicks);

  Observation Reset(uint32_t seed);
  StepResult Step(const Action& action);

  // GearScore of currently-equipped head+body. Meaningful at any time, but
  // named for its role as the terminal "gear collected" objective once
  // `done` is true.
  float FinalGearScore() const;

 private:
  World world_;
  RewardWeights weights_;
  int max_ticks_ = kDefaultMaxTicks;
  int tick_count_ = 0;
};

}  // namespace kingdoms
