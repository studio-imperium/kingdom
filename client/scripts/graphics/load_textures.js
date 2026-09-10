let textures
const DATA_ASSET_BASE = "https://gateway.kingdomcrushers.io/assets"
const spritesheet_data = {
  frames: {},
  meta: {
    image: "assets.png",
    format: "RGBA8888",
    size: { w: 512, h: 512 },
    scale: 1,
  },
}

function load_json_asset(name) {
  return fetch(`${DATA_ASSET_BASE}/${name}`).then((response) => response.json())
}

async function load_textures() {
  await load_tiles()
  await load_items()
  await load_npcs()
  await load_projectiles()
  await load_bombs()
  await load_animations()

  const assets_texture = await PIXI.Assets.load("assets/assets.png")
  const sheet = new PIXI.Spritesheet(assets_texture, spritesheet_data)
  await sheet.parse()
  sheet.textureSource.source.scaleMode = "nearest"
  textures = sheet.textures

  console.log("Loaded " + Object.keys(textures).length + " textures")
}

async function load_tiles() {
  const tiles_json = await load_json_asset("tiles.json")

  for (let { id, size, x, y } of tiles_json) {
    tile_data[id] = { size, x, y }
    spritesheet_data.frames[id] = {
      frame: { x: x + 8, y: y + 8, w: size, h: size },
      sourceSize: { w: size, h: size },
      spriteSourceSize: { x: 0, y: 0, w: size, h: size },
    }
  }
  tile_data = tiles_json
}

function load_recursive(obj) {
  if (obj.type == "container") {
    for (let child of obj.children) {
      load_recursive(child)
    }
  } else {
    const { w, h } = obj.sprite
    spritesheet_data.frames[obj.label] = {
      frame: obj.sprite,
      sourceSize: { w, h },
      spriteSourceSize: { x: 0, y: 0, w, h },
    }
  }
}

async function load_items() {
  const items_json = await load_json_asset("items.json")

  for (let { id, sprite, hand, equipped } of items_json) {
    spritesheet_data.frames[id] = {
      frame: sprite,
      sourceSize: { w: sprite.w, h: sprite.h },
      spriteSourceSize: { x: 0, y: 0, w: sprite.w, h: sprite.h },
    }

    if (equipped && Object.keys(equipped).length) {
      load_recursive(equipped)
    }
    if (hand && Object.keys(hand).length) {
      load_recursive(hand)
    }
  }
  item_data = items_json
}

async function load_npcs() {
  const npc_json = await load_json_asset("npcs.json")

  for (let { body } of npc_json) {
    for (let bodypart of body) {
      load_recursive(bodypart)
    }
  }
  npc_data = npc_json
}

async function load_projectiles() {
  const projectile_json = await load_json_asset("projectiles.json")

  for (let { object } of projectile_json) {
    for (let part of object) {
      load_recursive(part)
    }
  }
  projectile_data = projectile_json
}

async function load_bombs() {
  const bomb_json = await load_json_asset("bombs.json")

  for (let { object, trail } of bomb_json) {
    for (let part of object) {
      load_recursive(part)
    }
    console.log(trail)
    load_recursive(trail)
  }
  bomb_data = bomb_json
}

async function load_animations() {
  const animation_json = await load_json_asset("animations.json")
  animation_data = animation_json
}
