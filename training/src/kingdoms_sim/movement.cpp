#include "kingdoms_sim/movement.hpp"

namespace kingdoms {

namespace {

void ChaseTarget(Npc& npc, float speed, float dt, Vec2 target_pos, std::mt19937& rng) {
  const Vec2 aim = NearbyPoint(target_pos, Distance(npc.pos, target_pos), rng);
  const float dx = Sign(aim.x - npc.pos.x);
  const float dy = Sign(aim.y - npc.pos.y);
  npc.pos.x += dx * speed * dt;
  npc.pos.y += dy * speed * dt;
}

void RunFromTarget(Npc& npc, float speed, float dt, Vec2 target_pos, std::mt19937& rng) {
  const Vec2 aim = NearbyPoint(target_pos, Distance(npc.pos, target_pos), rng);
  const float dx = Sign(npc.pos.x - aim.x);
  const float dy = Sign(npc.pos.y - aim.y);
  npc.pos.x += dx * speed * dt;
  npc.pos.y += dy * speed * dt;
}

void OvershootTarget(Npc& npc, float speed, float dt, Vec2 target_pos, std::mt19937& rng) {
  const Vec2 aim = NearbyPoint(target_pos, 1.0f, rng);
  const float dx = Sign(aim.x - npc.pos.x);
  const float dy = Sign(aim.y - npc.pos.y);
  npc.pos.x += dx * speed * dt;
  npc.pos.y += dy * speed * dt;
}

void WanderStep(Npc& npc, float speed, float dt, std::mt19937& rng) {
  if (Distance(npc.movement_point, npc.pos) < 1.0f) {
    npc.movement_point = NearbyPoint(npc.origin, 4.0f, rng);
  }
  const float dx = Sign(npc.movement_point.x - npc.pos.x);
  const float dy = Sign(npc.movement_point.y - npc.pos.y);
  npc.pos.x += dx * speed * dt;
  npc.pos.y += dy * speed * dt;
}

}  // namespace

bool IsHovering(const Npc& npc, const AssetDatabase& assets, Vec2 target_pos) {
  const float dist = Distance(npc.pos, target_pos);
  const float range = assets.Npc(npc.type_id).range;
  const float too_close = range / 2.0f - 1.0f;
  const float too_far = range / 2.0f + 1.0f;
  return dist < too_far && dist > too_close;
}

void MoveNpc(Npc& npc, const AssetDatabase& assets, float dt_seconds, Vec2 player_pos,
             std::mt19937& rng) {
  if (!npc.InCombat()) npc.movement = MovementMode::Wander;

  const float speed = assets.Npc(npc.type_id).modes[npc.mode].speed;

  switch (npc.movement) {
    case MovementMode::Wander:
      WanderStep(npc, speed, dt_seconds, rng);
      break;
    case MovementMode::Chase:
      ChaseTarget(npc, speed, dt_seconds, player_pos, rng);
      break;
    case MovementMode::Run:
      RunFromTarget(npc, speed, dt_seconds, player_pos, rng);
      break;
    case MovementMode::Overshoot:
      OvershootTarget(npc, speed, dt_seconds, player_pos, rng);
      break;
    case MovementMode::Hover: {
      const float dist = Distance(npc.pos, player_pos);
      const float range = assets.Npc(npc.type_id).range;
      if (dist > range / 2.0f + 1.0f) {
        ChaseTarget(npc, speed, dt_seconds, player_pos, rng);
      } else if (dist < range / 2.0f - 1.0f) {
        RunFromTarget(npc, speed, dt_seconds, player_pos, rng);
      }
      break;
    }
    case MovementMode::Turret:
      break;  // stationary; Look()-only in the original, no gameplay effect
  }
}

}  // namespace kingdoms
