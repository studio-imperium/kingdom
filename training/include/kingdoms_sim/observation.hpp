#pragma once

#include <array>
#include <cstdint>
#include <type_traits>

#include "kingdoms_sim/types.hpp"
#include "kingdoms_sim/world.hpp"

// The policy-facing world state for one tick. No cosmetic facing angle (see
// action.hpp) -- but see ObservedEntity below, which gives the policy
// everything it needs to choose an attack direction deliberately (e.g. an
// NPC's remaining health, to prioritize finishing off a weak target).
//
// Fixed-size and trivially copyable throughout (enforced by the
// static_asserts at the bottom of this file), so the whole struct can be
// written straight into a PufferLib shared-memory observation buffer with a
// memcpy -- no variable-length containers anywhere. That's the whole reason
// this isn't a std::vector<ObservedEntity>: PufferLib's native PufferEnv
// convention wants one fixed-size block per environment instance, not a
// ragged one.

namespace kingdoms {

// Per-type slot caps, nearest-to-player-first. Generous placeholders sized
// off the arena's steady-state population (12 NPCs across 3 tiers, plus
// whatever they're actively firing) -- not measured against real training
// data yet. If Observation::truncated_count is ever consistently nonzero,
// these are too small for how this plays out in practice and should grow.
constexpr int kMaxNpcSlots = 16;
constexpr int kMaxProjectileSlots = 24;
constexpr int kMaxBombSlots = 12;
constexpr int kMaxLootSlots = 12;
constexpr int kMaxObservedEntities = kMaxNpcSlots + kMaxProjectileSlots + kMaxBombSlots + kMaxLootSlots;

struct ObservedEntity {
  ObservedEntityType type = ObservedEntityType::None;  // None = padding; the slot isn't a real entity
  Vec2 pos;

  float npc_health = 0.0f;
  uint8_t npc_type_id = 0;

  float projectile_damage = 0.0f;
  bool is_enemy_projectile = false;

  float bomb_damage = 0.0f;
  float bomb_radius = 0.0f;
  float bomb_timer = 0.0f;

  uint8_t loot_item_id = 0;
};

struct Observation {
  Vec2 player_pos;
  float health = 0.0f;
  float max_health = 0.0f;
  float attack_cooldown = 0.0f;  // raw value, can be slightly negative -- see character.hpp Tick()
  float speed = 0.0f;
  float power = 0.0f;
  float reload = 0.0f;
  float regen = 0.0f;

  uint8_t hand = 0;
  uint8_t head = 0;
  uint8_t body = 0;
  std::array<int16_t, kBackpackSlots> inventory{};  // -1 = empty slot, else item id

  // Slots [0, kMaxNpcSlots) are NPCs, then kMaxProjectileSlots worth of
  // projectiles, then bombs, then loot -- always in that fixed order, so a
  // Python-side reader can slice this array by constant offsets instead of
  // scanning for type tags. Unused slots at the end of each type's range are
  // zero-initialized ObservedEntity{} (type == None).
  std::array<ObservedEntity, kMaxObservedEntities> entities{};

  // How many live entities (summed across all 4 types) didn't fit this tick
  // and were dropped -- always the farthest ones, per type. Should be 0 in
  // normal play; a nonzero value here that shows up often means the caps
  // above need raising, not that anything is broken.
  uint16_t truncated_count = 0;
};

static_assert(std::is_trivially_copyable_v<ObservedEntity>);
static_assert(std::is_trivially_copyable_v<Observation>);

Observation BuildObservation(const World& world);

}  // namespace kingdoms
