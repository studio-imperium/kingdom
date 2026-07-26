#pragma once

#include <cstdint>

#include "kingdoms_sim/types.hpp"

// The policy's action for one tick. No cosmetic facing angle (the real
// game's `angle` byte was purely visual) -- but attacks DO get an explicit
// direction, so the policy can choose which enemy to prioritize rather than
// always hitting whatever's nearest. See combat.hpp's
// MakeDirectedProjectile/MakeDirectedBomb.

namespace kingdoms {

struct Action {
  Vec2 move;             // desired movement direction; internally clamped to unit length,
                          // then scaled by character.speed * dt (same formula as NPC movement)
  bool attack = false;    // attempt to attack/use the held item, subject to cooldown
  Vec2 attack_direction;  // direction to fire in, only read when attack=true; magnitude is
                          // ignored beyond a near-zero check (near-zero = no aim = counted as a
                          // miss). Independent of `move` -- the player can strafe one way while
                          // firing another.

  bool do_equip = false;
  uint8_t equip_from = 0;  // ChangeInventory(to=equip_to, from=equip_from)
  uint8_t equip_to = 0;    // slots 0-23 = backpack, 24 = head, 25 = body

  bool do_switch = false;
  uint8_t switch_to_slot = 0;  // SelectSlot(new=switch_to_slot); must be a backpack slot [0,24)
};

}  // namespace kingdoms
