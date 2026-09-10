function character_preview(character) {
  const preview = new PIXI.Container(), player = new PIXI.Container()
  const weapon = character.weapon
  const label = new PIXI.Text({ text: caption, style: { fontFamily: "myriad-pro", fontSize: 16, fill: "white", stroke: { color: "#1f1f1f", width: 3 } } })

  label.anchor.set(0.5, 0)
  label.position.set(112, 8)
  label.scale.set(Math.min(1, 208 / label.width))
  player.position.set(112, 80)
  player.scale.set(40 / 64)
  for (const part of [item_data[weapon]?.hand, item_data[character.body || 1].equipped, item_data[character.head || 0].equipped]) {
    if (part) player.addChild(build_object(part, lobby_preview.renderer, lobby_texture_cache))
  }
  preview.addChild(player, label)
  const canvas = lobby_preview.renderer.extract.canvas({ target: preview, frame: new PIXI.Rectangle(0, 0, 224, 128) })
  canvas.setAttribute("role", "img")
  canvas.setAttribute("aria-label", caption)
  preview.destroy({ children: true })
  return canvas
}
