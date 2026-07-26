#include "kingdoms_sim/environment.hpp"

namespace kingdoms {

Environment::Environment(const AssetDatabase& assets, uint32_t seed, RewardWeights weights,
                         int max_ticks)
    : world_(assets, seed), weights_(weights), max_ticks_(max_ticks) {}

Observation Environment::Reset(uint32_t seed) {
  world_.Reset(seed);
  tick_count_ = 0;
  return BuildObservation(world_);
}

StepResult Environment::Step(const Action& action) {
  StepResult result;
  result.events = world_.Step(action);
  result.reward = ComputeStepReward(result.events, world_.assets(), weights_);
  ++tick_count_;
  // Time-limit truncation, not just death: without it an agent that learns to
  // hide never ends an episode. No death penalty is applied on timeout.
  result.done = result.events.player_died || tick_count_ >= max_ticks_;
  result.observation = BuildObservation(world_);
  return result;
}

float Environment::FinalGearScore() const {
  // The episode high-water mark, not "whatever's equipped right now" --
  // consistent with the gear_upgrade reward term, and doesn't understate
  // progress if the agent happened to unequip something good just before
  // dying.
  return world_.best_gear_score();
}

}  // namespace kingdoms
