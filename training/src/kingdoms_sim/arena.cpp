#include "kingdoms_sim/arena.hpp"

#include <algorithm>
#include <cmath>

namespace kingdoms {

namespace {
constexpr float kTierBounds[4] = {0.0f, 40.0f, 100.0f, 200.0f};
constexpr int kTargetPerTier = 4;
constexpr float kRespawnCheckIntervalSeconds = 5.0f;
}  // namespace

ArenaSpawner::ArenaSpawner(const AssetDatabase& assets) {
  std::vector<NpcData> candidates;
  for (const auto& npc : assets.AllNpcs()) {
    // NPCs with no modes can't move/attack/pick a speed -- excluded so the
    // arena never spawns something that would immediately index out of
    // bounds when the sim looks up its current mode.
    if (!npc.modes.empty()) candidates.push_back(npc);
  }
  std::sort(candidates.begin(), candidates.end(), [](const NpcData& a, const NpcData& b) {
    return a.difficulty_score < b.difficulty_score;
  });

  tiers_.resize(3);
  for (int t = 0; t < 3; ++t) {
    tiers_[t].min_distance = kTierBounds[t];
    tiers_[t].max_distance = kTierBounds[t + 1];
    tiers_[t].target_population = kTargetPerTier;
  }

  const size_t n = candidates.size();
  for (size_t i = 0; i < n; ++i) {
    int tier = static_cast<int>(i * 3 / std::max<size_t>(n, 1));
    tier = std::min(tier, 2);
    tiers_[tier].npc_type_ids.push_back(candidates[i].id);
  }
}

int ArenaSpawner::TierIndexOf(Vec2 pos) const {
  const float dist = Distance(pos, Vec2{0.0f, 0.0f});
  for (size_t i = 0; i < tiers_.size(); ++i) {
    if (dist >= tiers_[i].min_distance && dist < tiers_[i].max_distance) return static_cast<int>(i);
  }
  return -1;
}

Vec2 ArenaSpawner::RandomPosInTier(const Tier& tier, std::mt19937& rng) const {
  std::uniform_real_distribution<float> angle_dist(0.0f, 6.28318530717958647692f);
  std::uniform_real_distribution<float> radius_dist(tier.min_distance, tier.max_distance);
  const float angle = angle_dist(rng);
  const float radius = radius_dist(rng);
  return Vec2{radius * std::cos(angle), radius * std::sin(angle)};
}

std::vector<std::pair<uint8_t, Vec2>> ArenaSpawner::InitialSpawns(std::mt19937& rng) {
  std::vector<std::pair<uint8_t, Vec2>> out;
  for (const auto& tier : tiers_) {
    if (tier.npc_type_ids.empty()) continue;
    std::uniform_int_distribution<size_t> pick(0, tier.npc_type_ids.size() - 1);
    for (int i = 0; i < tier.target_population; ++i) {
      out.emplace_back(tier.npc_type_ids[pick(rng)], RandomPosInTier(tier, rng));
    }
  }
  respawn_check_timer_ = 0.0f;
  return out;
}

std::vector<std::pair<uint8_t, Vec2>> ArenaSpawner::MaybeRespawn(
    const std::vector<Vec2>& alive_npc_positions, float dt_seconds, std::mt19937& rng) {
  respawn_check_timer_ += dt_seconds;
  if (respawn_check_timer_ < kRespawnCheckIntervalSeconds) return {};
  respawn_check_timer_ = 0.0f;

  std::vector<int> counts(tiers_.size(), 0);
  for (Vec2 pos : alive_npc_positions) {
    const int idx = TierIndexOf(pos);
    if (idx >= 0) counts[idx]++;
  }

  std::vector<std::pair<uint8_t, Vec2>> out;
  for (size_t t = 0; t < tiers_.size(); ++t) {
    const Tier& tier = tiers_[t];
    if (tier.npc_type_ids.empty()) continue;
    std::uniform_int_distribution<size_t> pick(0, tier.npc_type_ids.size() - 1);
    for (int i = counts[t]; i < tier.target_population; ++i) {
      out.emplace_back(tier.npc_type_ids[pick(rng)], RandomPosInTier(tier, rng));
    }
  }
  return out;
}

}  // namespace kingdoms
