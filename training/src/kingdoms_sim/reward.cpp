#include "kingdoms_sim/reward.hpp"

namespace kingdoms {

namespace {
float ItemScore(const ItemData& item) {
  const Stats& s = item.stats;
  return (s.health - 1.0f) + (s.regen - 1.0f) + (s.speed - 1.0f) + (s.damage - 1.0f) + (s.reload - 1.0f);
}
}  // namespace

float GearScore(const AssetDatabase& assets, uint8_t head_item, uint8_t body_item) {
  return ItemScore(assets.Item(head_item)) + ItemScore(assets.Item(body_item));
}

float ComputeStepReward(const TickEvents& events, const AssetDatabase& assets,
                         const RewardWeights& weights) {
  float reward = 0.0f;
  reward += weights.damage_dealt * events.damage_dealt;
  reward += weights.damage_taken * events.damage_taken;
  reward += weights.kill * static_cast<float>(events.kills);
  reward += weights.loot_pickup * static_cast<float>(events.loot_picked_up.size());
  // gear_score_delta is only ever set (in World::ResolvePlayerAction) when an
  // equip reaches a new episode-best GearScore, so it's always >= 0 here.
  reward += weights.gear_upgrade * events.gear_score_delta;
  reward += weights.attack_miss * static_cast<float>(events.attacks_missed);
  if (events.player_died) reward += weights.death;
  return reward;
}

}  // namespace kingdoms
