#pragma once

#include "kingdoms_sim/assets.hpp"
#include "kingdoms_sim/bomb.hpp"
#include "kingdoms_sim/projectile.hpp"
#include "kingdoms_sim/types.hpp"

// Mirrors the hit-test math in server/engine/combat.go. In the real game
// these hit tests ran inside the per-client simulation.go loop, not the main
// engine tick -- see world.hpp for where that logic actually lives here.

namespace kingdoms {

inline bool HitboxesIntersect(Vec2 proj_pos, float proj_hitbox, Vec2 entity_pos, float entity_hitbox) {
  const float radius = proj_hitbox + entity_hitbox;
  return DistanceSq(proj_pos, entity_pos) <= radius * radius;
}

inline bool WithinBombRadius(Vec2 bomb_pos, float bomb_radius, Vec2 entity_pos, float entity_hitbox) {
  const float radius = bomb_radius + entity_hitbox;
  return DistanceSq(bomb_pos, entity_pos) <= radius * radius;
}

// Spawns a projectile at `origin_pos + spawn`'s relative offset, auto-aimed
// at `target_pos`. Angle formula ported verbatim from npc.Projectile() in
// combat.go (including the -90 offset) since that's the one convention with
// hard evidence it lands hits in the live game. NPC attacks only -- see
// MakeDirectedProjectile for player attacks, which are policy-aimed.
Projectile MakeAimedProjectile(const AssetDatabase& assets, const ProjectileSpawnData& spawn,
                                Vec2 origin_pos, Vec2 target_pos, bool evil, float damage);

// Spawns a bomb landing at `target_pos + spawn`'s relative offset (bombs
// don't travel, so no angle is needed). Mirrors npc.Bomb() in combat.go. NPC
// attacks only -- see MakeDirectedBomb for player attacks.
Bomb MakeAimedBomb(const AssetDatabase& assets, const BombSpawnData& spawn, Vec2 target_pos,
                    bool evil, float damage);

// Spawns a projectile fired from `origin_pos` in `aim_direction` (must
// already be a unit vector), offset/angled by the weapon's own spawn data.
// This is the real formula from characterAttack in clients.go --
// `(projectile.Angle + angle) % 360` -- now that player attacks have an
// explicit direction again instead of an auto-aim substitute.
Projectile MakeDirectedProjectile(const AssetDatabase& assets, const ProjectileSpawnData& spawn,
                                   Vec2 origin_pos, Vec2 aim_direction, bool evil, float damage);

// Spawns a bomb landing `kPlayerBombThrowDistance` away from `thrown_from_pos`
// in `aim_direction` (unit vector). Lands at that exact point with no further
// spawn-offset applied, matching characterAttack in clients.go -- unlike NPC
// bombs, player bomb throws in the real game ignored the attack's bomb
// x/y offset entirely and used the raw click position.
Bomb MakeDirectedBomb(const AssetDatabase& assets, const BombSpawnData& spawn, Vec2 thrown_from_pos,
                       Vec2 aim_direction, bool evil, float damage);

}  // namespace kingdoms
