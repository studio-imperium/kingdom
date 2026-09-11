function render_character(canvas, size, hand, head, body) {
  const capture = new PIXI.Container()
  const character = build_preview_character(hand, head, body)

  character.position.set(size / 2, size / 2)
  character.scale.set(40 / 64)
  character.angle = 225
  capture.addChild(character)

  const rendered = preview.renderer.extract.canvas({
    target: capture,
    frame: new PIXI.Rectangle(0, 0, size, size)
  })

  canvas.width = canvas.height = size
  canvas.getContext("2d").drawImage(rendered, 0, 0)
  capture.destroy({ children: true })

  return canvas
}

