#include "kingdoms_sim/npc.hpp"

#include <algorithm>

namespace kingdoms {

bool Npc::ValidMode(const AssetDatabase& assets, uint8_t idx) const {
  const NpcModeData& mode = assets.Npc(type_id).modes[idx];
  if (mode.single_use && used_modes[idx]) return false;
  if (mode.max_health < health) return false;
  if (mode.min_health >= health) return false;
  return true;
}

void Npc::NewMode(const AssetDatabase& assets, std::mt19937& rng) {
  const NpcData& data = assets.Npc(type_id);
  std::vector<uint8_t> pool;

  for (uint8_t idx = 0; idx < data.modes.size(); ++idx) {
    if (ValidMode(assets, idx)) {
      pool.push_back(idx);
      if (data.modes[idx].priority) {
        pool = {idx};
        break;
      }
    }
  }

  if (!pool.empty()) {
    std::uniform_int_distribution<size_t> dist(0, pool.size() - 1);
    uint8_t chosen = pool[dist(rng)];
    mode = chosen;
    mode_timer = data.modes[chosen].duration;
    movement = data.modes[chosen].movement;
    attack_timer = 0.0f;
  }
}

const AttackData* Npc::CurrentAttack(const AssetDatabase& assets) const {
  const NpcModeData& m = assets.Npc(type_id).modes[mode];
  if (m.attacks.empty()) return nullptr;
  return &m.attacks[attack % m.attacks.size()];
}

bool Npc::CanAttack(const AssetDatabase& assets, Vec2 target_pos) const {
  if (!InCombat()) return false;
  const AttackData* atk = CurrentAttack(assets);
  if (!atk) return false;

  float attack_range = 0.0f;
  for (const auto& p : atk->projectiles) {
    attack_range = std::max(attack_range, assets.Projectile(p.id).range);
  }
  if (!atk->bombs.empty()) attack_range = 32.0f;
  if (!atk->summons.empty()) attack_range = std::max(attack_range, assets.Npc(type_id).range);

  return Distance(pos, target_pos) < attack_range;
}

Npc MakeNpc(const AssetDatabase& assets, uint8_t type_id, Vec2 pos) {
  Npc npc;
  npc.type_id = type_id;
  npc.pos = pos;
  npc.origin = pos;
  npc.movement_point = pos;  // forces Wander to pick a fresh roam point on its first tick
  npc.health = assets.Npc(type_id).health;
  npc.used_modes.assign(assets.Npc(type_id).modes.size(), false);
  return npc;
}

}  // namespace kingdoms
