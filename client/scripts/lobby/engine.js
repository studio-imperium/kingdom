async function init_lobby() {
  await app.init({ resizeTo: window, background: "#1f1f1f", resolution: 1, antialias: true })
  document.body.appendChild(app.canvas)
  app.canvas.style.imageRendering = "pixelated"
  app.stage.scale.set(40)
  await load_textures()
  const map = new DataView(await (await fetch("/assets/lobby.map")).arrayBuffer())
  let offset = 2
  for (let i = 0; i < map.getUint16(0, true); i++) offset += 6 + map.getUint8(offset + 5) * 2
  const map_size = map.getUint16(offset, true), center = Math.floor(map_size / 2)
  offset += 2
  await init_character_previews()
  function draw() {
    app.stage.position.set(app.screen.width / 2, app.screen.height / 2)
    const dx = Math.ceil(app.screen.width / (2 * app.stage.scale.x)) + 2, dy = Math.ceil(app.screen.height / (2 * app.stage.scale.y)) + 2
    for (let y = -dy; y <= dy; y++) {
      for (let x = -dx; x <= dx; x++) {
        if (!tile_map[`${x},${y}`]) add_tile(x, y, map.getUint8(offset + ((y + center) * map_size + x + center) * 3))
      }
    }
    render_tiles()
  }
  app.renderer.on("resize", draw)
  draw()
  await init_enemy_preview()
}
