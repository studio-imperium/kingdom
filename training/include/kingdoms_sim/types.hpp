#pragma once

#include <cmath>
#include <cstdint>
#include <random>

namespace kingdoms {

constexpr float kPi = 3.14159265358979323846f;

struct Vec2 {
  float x = 0.0f;
  float y = 0.0f;
};

inline float DistanceSq(Vec2 a, Vec2 b) {
  const float dx = a.x - b.x;
  const float dy = a.y - b.y;
  return dx * dx + dy * dy;
}

inline float Distance(Vec2 a, Vec2 b) {
  return std::sqrt(DistanceSq(a, b));
}

// atan2 angle in degrees, from `from` looking at `to`, normalized to [0, 360).
// Mirrors engine/utils.go Angle(): used only for NPC attacks, which always
// auto-aim at the player (true in the original game too -- NPC attacks were
// never client/AI-direction-driven). Player attacks use a policy-supplied
// direction instead; see combat.hpp's MakeDirectedProjectile/MakeDirectedBomb.
inline float AngleDegrees(Vec2 from, Vec2 to) {
  float angle = std::atan2(to.y - from.y, to.x - from.x) * 180.0f / kPi;
  if (angle < 0.0f) angle += 360.0f;
  return angle;
}

inline float Sign(float n) { return n > 0.0f ? 1.0f : -1.0f; }

// Uniform random point in the square [-within, within] around `origin`, matching
// engine/utils.go NearbyPoint (square jitter, not a circular/polar distribution).
inline Vec2 NearbyPoint(Vec2 origin, float within, std::mt19937& rng) {
  std::uniform_real_distribution<float> dist(-within, within);
  return {origin.x + dist(rng), origin.y + dist(rng)};
}

enum class ItemSlot { Head, Body, Hand };

enum class MovementMode { Wander, Chase, Run, Overshoot, Hover, Turret };

// None = 0 so a zero-initialized (padding) ObservedEntity slot reads as "empty" by default.
enum class ObservedEntityType : uint8_t { None = 0, Npc = 1, Projectile = 2, Bomb = 3, Loot = 4 };

// Backpack is slots [0, kBackpackSlots); head/body are addressed as the two
// slots immediately after, matching character.go's headSlot=24/bodySlot=25
// convention for ChangeInventory.
constexpr uint8_t kBackpackSlots = 24;
constexpr uint8_t kHeadSlot = 24;
constexpr uint8_t kBodySlot = 25;
constexpr uint8_t kDefaultHeadItem = 0;
constexpr uint8_t kDefaultBodyItem = 1;

constexpr float kTickSeconds = 0.05f;  // 50ms / 20Hz, matches engine.go's Run() ticker

// Character.GetHitbox() in character.go returns this hardcoded constant.
constexpr float kPlayerHitbox = 0.5f;

// Reserved id for the player in projectile hitlists (real entity ids from
// World::NextEntityId() start at 1, so 0 never collides with an NPC).
constexpr uint32_t kPlayerEntityId = 0;

// Placeholder: how far a player-thrown bomb lands in the aimed direction.
// The real game let the client click anywhere on the ground; there's no
// data-driven "throw range" to port, since that was a UI constraint, not an
// engine one. Tune freely.
constexpr float kPlayerBombThrowDistance = 8.0f;

}  // namespace kingdoms
