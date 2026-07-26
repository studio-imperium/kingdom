#include "kingdoms_sim/projectile.hpp"

#include <cmath>

namespace kingdoms {

// Mirrors Projectile.Tick in projectile.go exactly, including the somewhat
// odd /16 and 60fps-normalization constants baked into the original formula.
void Projectile::Tick(float dt_seconds, float speed) {
  const float delta_ms = dt_seconds * 1000.0f;
  const float delta_time = delta_ms * (60.0f / 1000.0f);
  const float rad = (angle_degrees - 90.0f) * (3.14159265358979323846f / 180.0f);
  const float dx = std::cos(rad);
  const float dy = std::sin(rad);

  pos.x += dx * speed * delta_time / 16.0f;
  pos.y += dy * speed * delta_time / 16.0f;
}

}  // namespace kingdoms
