#include <algorithm>
#include <cstdint>
#include <iostream>
#include <random>

#include "kingdoms_sim/assets.hpp"
#include "kingdoms_sim/environment.hpp"

#ifndef KINGDOMS_ASSETS_DIR
#error "KINGDOMS_ASSETS_DIR should be defined by CMake"
#endif

// Smoke test: runs a handful of random-action episodes end to end and prints
// per-episode stats. Doesn't train anything -- just proves the sim itself
// runs, produces observations/rewards, and terminates episodes on death.
int main() {
  kingdoms::AssetDatabase assets = kingdoms::AssetDatabase::LoadFromDirectory(KINGDOMS_ASSETS_DIR);
  kingdoms::Environment env(assets, /*seed=*/1);

  std::mt19937 action_rng(42);
  std::uniform_real_distribution<float> move_dist(-1.0f, 1.0f);
  std::uniform_real_distribution<float> chance(0.0f, 1.0f);

  constexpr int kEpisodes = 5;
  constexpr int kMaxStepsPerEpisode = 20000;  // 20000 * 50ms = ~1000s of sim time

  for (int ep = 0; ep < kEpisodes; ++ep) {
    env.Reset(static_cast<uint32_t>(ep));

    float total_reward = 0.0f;
    float total_damage_dealt = 0.0f;
    float total_damage_taken = 0.0f;
    int total_kills = 0;
    int total_loot = 0;
    int steps = 0;
    uint16_t max_truncated = 0;

    for (; steps < kMaxStepsPerEpisode; ++steps) {
      kingdoms::Action action;
      action.move = {move_dist(action_rng), move_dist(action_rng)};
      action.attack = chance(action_rng) < 0.3f;
      action.attack_direction = {move_dist(action_rng), move_dist(action_rng)};

      const kingdoms::StepResult result = env.Step(action);
      total_reward += result.reward;
      total_damage_dealt += result.events.damage_dealt;
      total_damage_taken += result.events.damage_taken;
      total_kills += result.events.kills;
      total_loot += static_cast<int>(result.events.loot_picked_up.size());
      max_truncated = std::max(max_truncated, result.observation.truncated_count);

      if (result.done) break;
    }

    std::cout << "episode " << ep << ": steps=" << steps << " reward=" << total_reward
              << " damage_dealt=" << total_damage_dealt << " damage_taken=" << total_damage_taken
              << " kills=" << total_kills << " loot=" << total_loot
              << " final_gear_score=" << env.FinalGearScore() << " max_truncated=" << max_truncated << "\n";
  }

  return 0;
}
