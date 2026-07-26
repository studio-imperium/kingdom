#include "kingdoms_sim/combat.hpp"

#include <cmath>

namespace kingdoms {

Projectile MakeAimedProjectile(const AssetDatabase& assets, const ProjectileSpawnData& spawn,
                                Vec2 origin_pos, Vec2 target_pos, bool evil, float damage) {
  Projectile p;
  p.type_id = spawn.id;
  p.evil = evil;
  p.damage = damage;
  p.pos = {origin_pos.x + spawn.x, origin_pos.y + spawn.y};
  p.origin = p.pos;

  const float raw = static_cast<float>(spawn.angle) + AngleDegrees(target_pos, origin_pos) - 90.0f;
  p.angle_degrees = std::fmod(std::fmod(raw, 360.0f) + 360.0f, 360.0f);
  return p;
}

Bomb MakeAimedBomb(const AssetDatabase& assets, const BombSpawnData& spawn, Vec2 target_pos,
                    bool evil, float damage) {
  Bomb b;
  b.type_id = spawn.id;
  b.evil = evil;
  b.damage = damage;
  b.pos = {target_pos.x + spawn.x, target_pos.y + spawn.y};
  b.origin = b.pos;
  b.timer = assets.Bomb(spawn.id).airtime;
  return b;
}

Projectile MakeDirectedProjectile(const AssetDatabase& assets, const ProjectileSpawnData& spawn,
                                   Vec2 origin_pos, Vec2 aim_direction, bool evil, float damage) {
  Projectile p;
  p.type_id = spawn.id;
  p.evil = evil;
  p.damage = damage;
  p.pos = {origin_pos.x + spawn.x, origin_pos.y + spawn.y};
  p.origin = p.pos;

  // Projectile::Tick applies its own -90 when turning angle_degrees into a
  // travel vector, so the +90 here cancels that out: with spawn.angle == 0,
  // the projectile ends up traveling exactly along aim_direction.
  const float theta = std::atan2(aim_direction.y, aim_direction.x) * 180.0f / kPi;
  const float raw = static_cast<float>(spawn.angle) + theta + 90.0f;
  p.angle_degrees = std::fmod(std::fmod(raw, 360.0f) + 360.0f, 360.0f);
  return p;
}

Bomb MakeDirectedBomb(const AssetDatabase& assets, const BombSpawnData& spawn, Vec2 thrown_from_pos,
                       Vec2 aim_direction, bool evil, float damage) {
  Bomb b;
  b.type_id = spawn.id;
  b.evil = evil;
  b.damage = damage;
  b.pos = {thrown_from_pos.x + aim_direction.x * kPlayerBombThrowDistance,
            thrown_from_pos.y + aim_direction.y * kPlayerBombThrowDistance};
  b.origin = b.pos;
  b.timer = assets.Bomb(spawn.id).airtime;
  return b;
}

}  // namespace kingdoms
