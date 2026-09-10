const addr = sessionStorage.getItem("kingdom.server") || "server.kingdomcrushers.io"
const prefixs = ["wss", "https"]
// const addr = "localhost:8082"
// const prefixs = ["ws", "http"]
let CONNECTED = false
let socket
let token

const [
  HANDSHAKE,
  CHARACTER_POSITION,
  CHARACTER_ATTACK,
  RECIEVE_ATTACK,
  WORLD_STATE,
  DAMAGED,
  TILES,
  SELECT_SLOT,
  CHAT_MESSAGE,
  CHANGE_INVENTORY,
  SET_HEALTH,
  CHARACTER_DEAD,
  LOOT_LOOTED,
  DROP_ITEM,
] = [0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13]

function get_character(id) {
  if (characters[id]) {
    return characters[id]
  } else if (token == id) {
    return character
  } else {
    return null
  }
}

function handshake() {
  const name = new TextEncoder().encode(sessionStorage.getItem("kingdom.player_name") || "Guest")
  const buffer = new ArrayBuffer(41 + name.length)
  const data = new DataView(buffer)

  data.setUint8(0, HANDSHAKE)
  new Uint8Array(buffer, 1, 32).set(account.token.match(/../g).map(byte => parseInt(byte, 16)))
  data.setBigInt64(33, BigInt(account.guest ? 0 : account.data.characters[0].id), true)
  new Uint8Array(buffer, 41).set(name)

  socket.send(data)
}

let initialized_character = false
function set_character(data) {
  const [x, y, angle, health, max_health, _reload, _speed, hand, head, body] = [
    data.getFloat32(1, true),
    data.getFloat32(5, true),
    data.getUint16(9, true),
    data.getUint16(11, true),
    data.getUint16(13, true),
    data.getFloat32(15, true),
    data.getFloat32(19, true),
    data.getUint8(23),
    data.getUint8(24),
    data.getUint8(25),
  ]

  speed = 4 * _speed
  reload = _reload

  const inventory = {}
  const slots = data.getUint8(26)
  token = data.getUint32(27 + slots * 2, true)
  for (let i = 0; i < slots; i++) {
    let offset = i * 2
    let slot = data.getUint8(27 + offset)
    let item = data.getUint8(28 + offset)
    inventory[slot] = item
  }

  if (!initialized_character) {
    init_character(x, y, angle, health, inventory[hand], head, body, inventory)
    start_engine()
    initialized_character = true
  } else {
    character.update(
      x,
      y,
      angle,
      health,
      inventory[hand],
      head,
      body,
      inventory,
    )
  }
  refresh_inventory(inventory, hand, head, body)
  update_preview()
  update_healthbar(health, max_health)
}

function set_world(data) {
  let offset = 1
  const character_count = data.getUint16(offset, true)
  offset += 2

  for (let i = 0; i < character_count; i++) {
    let id = data.getUint32(offset, true)
    let x = data.getFloat32(4 + offset, true)
    let y = data.getFloat32(8 + offset, true)
    let angle = data.getUint16(12 + offset, true)
    let health = data.getUint16(14 + offset, true)
    let hand = data.getUint8(16 + offset)
    let head = data.getUint8(17 + offset)
    let body = data.getUint8(18 + offset)

    if (id == token) {
    } else if (characters[id]) {
      characters[id].update(x, y, angle, health, hand, head, body)
    } else {
      characters[id] = new Character(x, y, angle, health, hand, head, body, character_names[id] || "")
    }
    offset += 19
  }

  const npc_count = data.getUint16(offset, true)
  offset += 2

  for (let i = 0; i < npc_count; i++) {
    let id = data.getUint32(offset, true)
    let which = data.getUint8(4 + offset)
    let x = data.getFloat32(5 + offset, true)
    let y = data.getFloat32(9 + offset, true)
    let health = data.getFloat32(13 + offset, true)
    let target = data.getUint8(17 + offset)
    let target_character = null

    offset += 18
    if (target) {
      let cid = data.getUint32(offset, true)
      target_character = get_character(cid)
      offset += 4
    }

    if (npcs[id]) {
      let npc = npcs[id]
      if (target_character) {
        npc.interpolator.look_at(target_character)
      } else {
        npc.interpolator.look_away(character)
      }
      npc.update(x, y, health)
    } else {
      const npc = new Npc(which, x, y, health)
      if (target_character) {
        npc.interpolator.look_at(target_character)
      }
      npcs[id] = npc
    }
  }

  const loot_count = data.getUint16(offset, true)
  offset += 2

  for (let i = 0; i < loot_count; i++) {
    let id = data.getUint32(offset, true)
    let which = data.getUint8(4 + offset)
    let x = data.getFloat32(5 + offset, true)
    let y = data.getFloat32(9 + offset, true)
    offset += 13

    if (loots[id]) {
      loots[id].update()
    } else {
      loots[id] = new Loot(which, x, y)
    }
  }

  if (offset < data.byteLength) {
    const count = data.getUint16(offset, true)
    offset += 2
    for (let i = 0; i < count; i++) {
      const id = data.getUint32(offset, true), length = data.getUint8(offset + 4)
      character_names[id] = new TextDecoder().decode(new Uint8Array(data.buffer, data.byteOffset + offset + 5, length))
      if (characters[id]) characters[id].nameplate.text = character_names[id]
      offset += 5 + length
    }
  }
}

function set_attack(data) {
  let offset = 1
  const id = data.getUint32(offset, true)
  offset += 4
  const animation = data.getUint8(offset)
  offset += 1
  const duration = data.getUint16(offset, true)
  offset += 2
  const projectile_count = data.getUint16(offset, true)
  offset += 2

  if (characters[id]) {
    characters[id].animator.animate(animation, duration / 1000)
  }
  if (npcs[id]) {
    npcs[id].animator.animate(animation, duration / 1000)
  }

  for (let i = 0; i < projectile_count; i++) {
    const projectile_id = data.getUint32(offset, true)
    offset += 4
    const which = data.getUint8(offset)
    offset += 1
    const x = data.getFloat32(offset, true)
    offset += 4
    const y = data.getFloat32(offset, true)
    offset += 4
    const angle = data.getUint16(offset, true)
    offset += 2

    const projectile = new Projectile(which, x, y, angle, false, id)
    projectiles[projectile_id] = projectile
  }

  const bomb_count = data.getUint16(offset, true)
  offset += 2

  for (let i = 0; i < bomb_count; i++) {
    const bomb_id = data.getUint32(offset, true)
    offset += 4
    const which = data.getUint8(offset)
    offset += 1
    const x = data.getFloat32(offset, true)
    offset += 4
    const y = data.getFloat32(offset, true)
    offset += 4
    const origin_x = data.getFloat32(offset, true)
    offset += 4
    const origin_y = data.getFloat32(offset, true)
    offset += 4

    const bomb = new Bomb(which, origin_x, origin_y, x, y)
    bombs[bomb_id] = bomb
  }
}

function set_tiles(data) {
  let offset = 1
  const origin_x = data.getInt32(offset, true)
  const origin_y = data.getInt32(4 + offset, true)
  const tile_count = data.getUint16(8 + offset, true)
  offset += 10

  for (let i = 0; i < tile_count; i++) {
    let x = data.getInt32(offset, true)
    let y = data.getInt32(4 + offset, true)
    let tile_id = data.getUint8(8 + offset)

    if (in_map_bounds(x, y)) {
      tiles[tile_offset(x, y)] = tile_id
    }
    offset += 9
  }
}

function set_health(data) {
  const health = data.getUint16(1, true)
  const max_health = data.getUint16(3, true)

  update_healthbar(health, max_health)
}

function loot_loot(data) {
  const id = data.getUint32(1, true)

  if (loots[id]) {
    loots[id].kill(id)
  }
}

function damaged(data) {
  let offset = 1
  const id = data.getUint32(offset, true)

  if (characters[id]) {
    characters[id].damage()
  }
  if (npcs[id]) {
    npcs[id].damage()
  }
  if (id == token) {
    character.damage()
  }
}

function set_message(data) {
  let offset = 1
  const id = data.getUint32(offset, true)
  offset += 4

  const n = data.getUint8(offset)
  offset += 1

  const sender = new TextDecoder().decode(new Uint8Array(data.buffer, data.byteOffset + offset, n))
  offset += n

  const m = data.getUint8(offset)
  offset += 1

  const msg = new TextDecoder().decode(new Uint8Array(data.buffer, data.byteOffset + offset, m))

  add_message(id, msg, sender)
}

function send_position(x, y, angle) {
  const buffer = new ArrayBuffer(11)
  const data = new DataView(buffer)

  data.setUint8(0, CHARACTER_POSITION)
  data.setFloat32(1, x, true)
  data.setFloat32(5, y, true)
  data.setUint16(9, angle, true)

  socket.send(data)
}

function send_attack(x, y, angle) {
  const buffer = new ArrayBuffer(19)
  const data = new DataView(buffer)

  let [target_x, target_y] = get_mouse_target()

  data.setUint8(0, CHARACTER_ATTACK)
  data.setFloat32(1, x, true)
  data.setFloat32(5, y, true)
  data.setFloat32(9, target_x, true)
  data.setFloat32(13, target_y, true)
  data.setUint16(17, angle, true)

  socket.send(data)
}

function select_slot(idx) {
  const buffer = new ArrayBuffer(2)
  const data = new DataView(buffer)

  data.setUint8(0, SELECT_SLOT)
  data.setUint8(1, idx)

  socket.send(data)
}

function send_message(msg) {
  const message = new TextEncoder().encode(msg)
  const n = message.length
  if (n > 255) return
  const buffer = new ArrayBuffer(2 + n)
  const data = new DataView(buffer)

  data.setUint8(0, CHAT_MESSAGE)
  data.setUint8(1, n)
  new Uint8Array(buffer, 2).set(message)

  socket.send(data)
}

function change_inventory(to, from) {
  const buffer = new ArrayBuffer(3)
  const data = new DataView(buffer)

  data.setUint8(0, CHANGE_INVENTORY)
  data.setUint8(1, to)
  data.setUint8(2, from)

  console.log(to, from)

  socket.send(data)
}

function drop_item(slot) {
  const buffer = new ArrayBuffer(2)
  const data = new DataView(buffer)
  data.setUint8(0, DROP_ITEM)
  data.setUint8(1, slot)
  socket.send(data)
}

async function connect() {
  let leaving = false
  window.addEventListener("pagehide", () => {
    leaving = true
    app.stop()
    socket?.close(1000, "Leaving game")
  })
  window.addEventListener("pageshow", event => { if (event.persisted) location.reload() })
  try {
    await restore_session()
  } catch (error) {
    closed(error); return
  }

  if (leaving) return
  if (!account?.valid) {
    closed()
    return
  }
  socket = new WebSocket(prefixs[0] + "://" + addr + "/connect")
  socket.binaryType = "arraybuffer"

  function open() {
    CONNECTED = true
    handshake()
  }

  function closed(e) {
    console.log(e)
    CONNECTED = false
    if (!leaving) {
      const status = document.getElementById("connecting_status")
      if (status) status.innerHTML = 'Connection closed. <a href="/" style="color:inherit">Return to lobby</a>'
    }
  }

  function handle_packet(e) {
    const data = new DataView(e.data)
    const packet_type = data.getUint8(0)

    switch (packet_type) {
      case HANDSHAKE:
        set_character(data)
        break
      case WORLD_STATE:
        set_world(data)
        document.getElementById("connecting_screen")?.remove()
        break
      case RECIEVE_ATTACK:
        set_attack(data)
        break
      case DAMAGED:
        damaged(data)
        break
      case TILES:
        set_tiles(data)
        break
      case CHAT_MESSAGE:
        set_message(data)
        break
      case SET_HEALTH:
        set_health(data)
        break
      case LOOT_LOOTED:
        loot_loot(data)
        break
      case CHARACTER_DEAD:
        location.href = "/"
        break
      default:
        console.log("Bad packet recieved: ", packet_type)
    }
  }

  socket.addEventListener("open", open)
  socket.addEventListener("close", closed)
  socket.addEventListener("error", closed)
  socket.addEventListener("message", handle_packet)
}
