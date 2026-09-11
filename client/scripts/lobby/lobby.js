const OBJECT_SIZE = 8
const TILE_SIZE = 8
const app = new PIXI.Application()

async function init_lobby() {
  await app.init({
    background: "#1f1f1f",
    resizeTo: window,
    width: window.innerWidth,
    height: window.innerHeight,
    useContextAlpha: false,
    antialias: true,
    autoDensity: true,
    resolution: 1,
  })
  document.body.appendChild(app.canvas)

  app.stage.scale = 48
  app.canvas.style.imageRendering = "pixelated"
  app.canvas.style.imageRendering = "crisp-edges"

  await PIXI.Assets.load("/assets/myriad-pro.ttf")
  await load_textures()
  await load_background_map()
  await init_preview()
}

async function load_background_map() {
  const map_file = await fetch("/assets/sandy.map")
  const map_bytes = await map_file.arrayBuffer()
  const map = new DataView(map_bytes)

  let offset = 2
  for (let i = 0; i < map.getUint16(0, true); i++) {
    offset += 6 + map.getUint8(offset + 5) * 2
  }

  const map_size = map.getUint16(offset, true)
  const center = Math.floor(map_size / 2)
  offset += 2


  function resize_background_map() {
    const width = app.screen.width
    const height = app.screen.height
    const scale = app.stage.scale

    app.stage.position.set(width / 2, height / 2)

    const dx = Math.ceil(width / (2 * scale.x)) + 2
    const dy = Math.ceil(height / (2 * scale.y)) + 2

    for (let y = -dy; y <= dy; y++) {
      for (let x = -dx; x <= dx; x++) {
        if (!tile_map[`${x},${y}`]) add_tile(x, y, map.getUint8(offset + ((y + center) * map_size + x + center) * 3))
      }
    }
    render_tiles()
  }
  resize_background_map()
  app.renderer.on("resize", resize_background_map)
}
