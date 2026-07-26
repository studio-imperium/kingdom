#pragma once

#include <random>

#include "kingdoms_sim/assets.hpp"
#include "kingdoms_sim/npc.hpp"

// Mirrors server/engine/movement.go's six movement modes.

namespace kingdoms {

bool IsHovering(const Npc& npc, const AssetDatabase& assets, Vec2 target_pos);

// Dispatches on npc.movement (forced to Wander when not InCombat(), matching
// Npc.Move() in npc.go). `player_pos` is unused by Wander/Turret but always
// passed for simplicity -- there's only ever one possible target.
void MoveNpc(Npc& npc, const AssetDatabase& assets, float dt_seconds, Vec2 player_pos,
             std::mt19937& rng);

}  // namespace kingdoms
