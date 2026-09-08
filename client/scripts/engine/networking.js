const addr = "server.kingdomcrushers.io"
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
  token = (Math.random() * 0x100000000) >>> 0
  const buffer = new ArrayBuffer(5)
  const data = new DataView(buffer)

  data.setUint8(0, HANDSHAKE)
  data.setUint32(1, token, true)

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
      characters[id] = new Character(x, y, angle, health, hand, head, body)
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

    const projectile = new Projectile(which, x, y, angle)
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

  let sender = ""
  for (let i = 0; i < n; i++) {
    sender += String.fromCharCode(data.getUint8(offset))
    offset++
  }

  const m = data.getUint8(offset)
  offset += 1

  let msg = ""
  for (let i = 0; i < m; i++) {
    msg += String.fromCharCode(data.getUint8(offset))
    offset++
  }

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
  const n = msg.length
  const buffer = new ArrayBuffer(2 + n)
  const data = new DataView(buffer)

  data.setUint8(0, CHAT_MESSAGE)
  data.setUint8(1, n)
  for (let i = 0; i < n; i++) {
    data.setUint8(i + 2, msg.charCodeAt(i))
  }

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

function connect() {
  socket = new WebSocket(prefixs[0] + "://" + addr + "/connect")
  socket.binaryType = "arraybuffer"

  function open() {
    CONNECTED = true
    handshake()
  }

  function closed(e) {
    console.log(e)
    CONNECTED = false
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
      default:
        console.log("Bad packet recieved: ", packet_type)
    }
  }

  socket.addEventListener("open", open)
  socket.addEventListener("close", closed)
  socket.addEventListener("error", closed)
  socket.addEventListener("message", handle_packet)
}
