const render_dist = 32
const size = 2500
const tiles = new Uint8Array(size * size)
const added = new Uint8Array(size * size)
tiles.fill(0)

function in_map_bounds(x, y) {
  return x >= 0 && x < size && y >= 0 && y < size
}

function tile_offset(x, y) {
  return y * size + x
}

function outside_range(x, y) {
  let player_x = character.object.x
  let player_y = character.object.y
  let halfdist = render_dist / 2

  return (
    x < player_x - halfdist ||
    x >= player_x + halfdist ||
    y < player_y - halfdist ||
    y >= player_y + halfdist
  )
}
function distance(x1, y1, x2, y2) {
  return Math.sqrt(Math.pow(x1 - x2, 2) + Math.pow(y1 - y2, 2))
}

let last_x
let last_y
function diff(x, y) {
  let difference = Math.abs(last_x - x) + Math.abs(last_y - y)
  last_x = x
  last_y = y
  return difference
}

let timer = 0
let elapsed = 0
function start_engine() {
  app.ticker.add(({ deltaMS }) => {
    timer += deltaMS
    for (let id of Object.keys(characters)) {
      let { interpolator, animator, colorAnimator, object } = characters[id]
      interpolator.tick(deltaMS)
      animator.tick(deltaMS)
      if (colorAnimator) {
        colorAnimator.tick(deltaMS)
      }

      if (Date.now() > interpolator.last_frame + 400) {
        characters[id].kill(id)
      }
    }
    for (let id of Object.keys(npcs)) {
      let { interpolator, animator, colorAnimator, object } = npcs[id]
      interpolator.tick(deltaMS)
      animator.tick(deltaMS)
      if (colorAnimator) {
        colorAnimator.tick(deltaMS)
      }

      if (Date.now() > interpolator.last_frame + 400) {
        npcs[id].kill(id)
      }
    }
    for (let id of Object.keys(loots)) {
      let { last_update, dead, object } = loots[id]

      if (dead) {
        let new_scale = object.scale.x - deltaMS / 10000
        if (new_scale < 0) {
          object.scale.set(0)
        } else {
          object.scale.set(new_scale)
        }
      } else if (Date.now() > last_update + 400) {
        loots[id].kill(id)
      }
    }

    character_tick(deltaMS)
    projectile_tick(deltaMS)
    bomb_tick(deltaMS)

    if (timer >= 500) {
      timer = 0
      tile_tick()
    }

    elapsed += deltaMS / 1000
    tile_animations()
  })
}

class Cell {
  constructor(origin_x, origin_y) {
    this.visible = false
    this.tiles = []
    this.origin = [origin_x, origin_y]
    cells.push(this)
  }

  add_tile(x, y, tile_idx) {
    this.tiles.push([x, y, tile_idx])
  }

  hide() {
    if (this.visible) {
      this.visible = false
      for (let tile of this.tiles) {
        let x = tile[0]
        let y = tile[1]
        let data = tile_map[`${x},${y}`]

        if (data) {
          data.mesh.destroy()
          delete tile_map[`${x},${y}`]
        }
      }
    }
  }
  show() {
    if (!this.visible) {
      this.visible = true
      let player_x = character.object.x
      let player_y = character.object.y

      for (let tile of this.tiles) {
        let x = tile[0]
        let y = tile[1]
        let tile_id = tile[2]

        if (distance(x, y, player_x, player_y) < 16) {
        }
        add_tile(x, y, tile_id)
      }
    }
  }
}

class Interpolator {
  constructor(object) {
    this.set_object(object)
  }

  set_object(object) {
    this.object = object
    this.frames = []
    this.last_frame = Date.now()
  }

  lerp_angle_degrees(a, b, t) {
    d1 * frame1[2] + d2 * frame2[2]
    const delta = ((b - a + 540) % 360) - 180
    return a + delta * t
  }

  tick(delta_ms) {
    const time = Date.now()

    while (this.frames.length > 2) {
      this.frames.shift()
    }
    if (this.frames.length > 1) {
      let frame1 = this.frames[0]
      let frame2 = this.frames[1]
      let diff = frame2[0] - frame1[0]

      const d1 = (frame2[0] - time) / diff
      const d2 = (time - frame1[0]) / diff

      // npcs
      if (frame1.length == 3 || frame2.length == 3) {
        const dx = frame2[1] - frame1[1]
        const dy = frame2[2] - frame1[2]

        if (frame1[3] === undefined) {
          frame1[3] = this.object.angle
        }

        if (dx !== 0 || dy !== 0) {
          frame2[3] = Math.atan2(dy, dx) * (180 / Math.PI) + 90
        } else if (frame2[3] === undefined) {
          frame2[3] = frame1[3]
        }
      }

      if (frame1[1] != frame2[1]) {
        this.object.x = d1 * frame1[1] + d2 * frame2[1]
      }
      if (frame1[2] != frame2[2]) {
        this.object.y = d1 * frame1[2] + d2 * frame2[2]
      }
      if (frame1[3] != frame2[3]) {
        if (frame1[3] - frame2[3] > 180) {
          this.object.angle = d1 * frame1[3] + d2 * (frame2[3] + 360)
        } else if (frame2[3] - frame1[3] > 180) {
          this.object.angle = d1 * (frame1[3] + 360) + d2 * frame2[3]
        } else {
          this.object.angle = d1 * (frame1[3] + 360) + d2 * (frame2[3] + 360)
        }
      }
    }
  }

  look_at(object) {
    this.target = object
  }
  look_away() {
    this.target = null
  }

  add_char_frame(x, y, angle) {
    this.last_frame = Date.now()
    this.frames.push([this.last_frame + 200, x, y, angle % 360])
  }
  add_npc_frame(x, y) {
    this.last_frame = Date.now()
    if (!this.target) {
      this.frames.push([this.last_frame + 200, x, y])
    } else {
      const dx = this.target.object.x - this.object.x
      const dy = this.target.object.y - this.object.y
      const angle = Math.atan2(dy, dx) * (180 / Math.PI) + 90

      this.frames.push([this.last_frame + 200, x, y, angle])
    }
  }
}

class Animator {
  constructor(object) {
    this.set_object(object)
  }

  set_object(object) {
    this.object = object
    this.head = object.getChildByName("head")
    this.body = object.getChildByName("body")
    this.hand = object.getChildByName("hand")
    this.animation = null
    this.timestamp = 0
  }

  tick(delta_ms) {
    if (this.animation != null) {
      this.timestamp += delta_ms
      const frames = this.animation.frames || []

      if (frames.length === 0) {
        this.animation = null
        this.timestamp = 0
        return
      }

      const progress = this.timestamp / (this.duration * 1000)
      let frame1 = frames[0]
      let frame2 = null

      for (let i = 0; i < frames.length; i++) {
        const frame = frames[i]
        if (frame.time <= progress) {
          frame1 = frame
        } else {
          frame2 = frame
          break
        }
      }

      const applyFrame = (frame) => {
        if (this.head && frame.head_angle !== undefined) {
          this.head.angle = frame.head_angle
        }
        if (this.body && frame.body_angle !== undefined) {
          this.body.angle = frame.body_angle
        }
        if (this.hand && frame.hand_angle !== undefined) {
          this.hand.angle = frame.hand_angle
        }
        if (this.hand && frame.hand_y !== undefined) {
          this.hand.y = frame.hand_y
        }
        if (this.hand && frame.hand_x !== undefined) {
          this.hand.x = frame.hand_x
        }
        if (this.hand && frame.hand_scale !== undefined) {
          if (this.hand.scale && typeof this.hand.scale.set === "function") {
            this.hand.scale.set(frame.hand_scale)
          } else {
            this.hand.scale = frame.hand_scale
          }
        }
        if (this.object && frame.object_scale !== undefined) {
          if (
            this.object.scale &&
            typeof this.object.scale.set === "function"
          ) {
            this.object.scale.set(frame.object_scale / 64)
          } else {
            this.object.scale = frame.object_scale / 64
          }
        }
      }

      if (!frame2) {
        applyFrame(frame1)
        this.animation = null
        this.timestamp = 0
        return
      }

      const diff = frame2.time - frame1.time
      const t = diff > 0 ? (progress - frame1.time) / diff : 1
      const lerp = (a, b) => a + (b - a) * t

      applyFrame({
        object_scale:
          frame1.object_scale !== undefined && frame2.object_scale !== undefined
            ? lerp(frame1.object_scale, frame2.object_scale)
            : frame1.object_scale,
        head_angle:
          frame1.head_angle !== undefined && frame2.head_angle !== undefined
            ? lerp(frame1.head_angle, frame2.head_angle)
            : frame1.head_angle,
        body_angle:
          frame1.body_angle !== undefined && frame2.body_angle !== undefined
            ? lerp(frame1.body_angle, frame2.body_angle)
            : frame1.body_angle,
        hand_angle:
          frame1.hand_angle !== undefined && frame2.hand_angle !== undefined
            ? lerp(frame1.hand_angle, frame2.hand_angle)
            : frame1.hand_angle,
        hand_scale:
          frame1.hand_scale !== undefined && frame2.hand_scale !== undefined
            ? lerp(frame1.hand_scale, frame2.hand_scale)
            : frame1.hand_scale,
        hand_y:
          frame1.hand_y !== undefined && frame2.hand_y !== undefined
            ? lerp(frame1.hand_y, frame2.hand_y)
            : frame1.hand_y,
        hand_x:
          frame1.hand_x !== undefined && frame2.hand_x !== undefined
            ? lerp(frame1.hand_x, frame2.hand_x)
            : frame1.hand_x,
      })
    }
  }

  animate(which, duration = 0) {
    this.animation = animation_data[which]
    this.timestamp = 0
    this.duration = duration
  }
}

class ColorAnimator {
  constructor(object) {
    this.object = object
    this.active = false
    this.duration = 0
    this.elapsed = 0
    this.startR = 255
    this.startG = 255
    this.startB = 255
  }

  tick(delta_ms) {
    if (!this.active || !this.object) {
      return
    }

    this.elapsed += delta_ms
    const duration = this.duration
    const t = duration > 0 ? Math.min(this.elapsed / duration, 1) : 1

    const r = Math.round(this.startR + (255 - this.startR) * t)
    const g = Math.round(this.startG + (255 - this.startG) * t)
    const b = Math.round(this.startB + (255 - this.startB) * t)

    this.object.tint = (r << 16) | (g << 8) | b

    if (t >= 1) {
      this.object.tint = 0xffffff
      this.active = false
    }
  }

  animate(color, duration_ms) {
    const value = color >>> 0
    this.startR = (value >> 16) & 0xff
    this.startG = (value >> 8) & 0xff
    this.startB = value & 0xff
    this.duration = Math.max(0, duration_ms || 0)
    this.elapsed = 0
    this.active = true

    if (this.object) {
      this.object.tint = value
    }

    if (this.duration === 0) {
      if (this.object) {
        this.object.tint = 0xffffff
      }
      this.active = false
    }
  }
}
