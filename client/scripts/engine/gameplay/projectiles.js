const projectiles = {}

class Projectile {
  constructor(id, x, y, angle, mine = false, owner = null) {
    this.object = build_projectile(id)
    this.which = id
    this.origin_x = x
    this.origin_y = y
    this.object.x = x
    this.object.y = y
    this.object.angle = angle
    this.object.scale.set(0)
    this.mine = mine
    this.owner = mine ? token : owner

    add_object(this.object)
  }

  kill(id) {
    this.object.destroy()
    delete projectiles[id]
  }
}

function build_projectile(projectile_id) {
  let projectile = new PIXI.Container()

  for (let part of projectile_data[projectile_id].object) {
    let obj = build_object(part)
    body_layer.attach(obj)
    projectile.addChild(obj)
  }
  projectile.alpha = 0
  return projectile
}

function projectile_tick(deltaMS) {
  for (let id of Object.keys(projectiles)) {
    const { object, which, origin_x, origin_y, mine, owner } = projectiles[id]
    const data = projectile_data[which]
    const speed = data.speed
    const rad = (object.angle - 90) * (Math.PI / 180)
    const dx = Math.cos(rad)
    const dy = Math.sin(rad)
    object.x += (dx * speed * deltaMS * 6) / 1600
    object.y += (dy * speed * deltaMS * 6) / 1600

    let [proj_x, proj_y] = [object.x, object.y]
    let travelled_distance = distance(origin_x, origin_y, proj_x, proj_y)
    let out_of_range = travelled_distance > data.range
    let hit_enemy = false

    if (object.scale.x * 64 < 1) {
      object.scale.set(object.scale.x + 1 / 500)
    } else {
      object.scale.set(1 / 64)
    }

    if (travelled_distance < data.range / 2) {
      object.alpha += deltaMS / 500
    } else {
      object.alpha = data.range - travelled_distance
    }

    if (!data.piercing) {
      for (const [target_id, target] of Object.entries({ ...characters, [token]: character })) {
        if (
          target && Number(target_id) !== owner &&
          distance(target.object.x, target.object.y, proj_x, proj_y) <= data.hitbox + 0.5
        ) {
          hit_enemy = true
          break
        }
      }
    }

    if (!data.piercing && mine && !hit_enemy) {
      for (let npc of Object.values(npcs)) {
        if (
          distance(npc.object.x, npc.object.y, proj_x, proj_y) <
          data.hitbox + npc_data[npc.which].hitbox
        ) {
          hit_enemy = true
          npc.damage()
          break
        }
      }
    }

    if (out_of_range || hit_enemy) {
      projectiles[id].kill(id)
    }
  }
}
