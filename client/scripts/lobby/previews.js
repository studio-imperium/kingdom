const preview_canvas = document.getElementById("character_preview")
const preview = new PIXI.Application()
let preview_character = null
const preview_texture_cache = {}

function angle_preview(x, y) {
  center_x =
    preview_canvas.getBoundingClientRect().left + preview_canvas.width / 2
  center_y =
    preview_canvas.getBoundingClientRect().top + preview_canvas.height / 2

  const dx = x - center_x
  const dy = y - center_y
  preview_character.angle = Math.atan2(dy, dx) * (180 / Math.PI) + 90
}

document.addEventListener("mousemove", (e) => {
  mouse_x = e.clientX
  mouse_y = e.clientY
  angle_preview(mouse_x, mouse_y)
})

async function init_preview() {
  await preview.init({
    canvas: preview_canvas,
    resizeTo: document.querySelector(".inventory_top"),
    width: 256,
    height: 256,
    useContextAlpha: false,
    backgroundAlpha: 0,
    antialias: true,
    autoDensity: true,
    resolution: 1,
  })
  preview.canvas.style.imageRendering = "pixelated"
}

function build_preview_character(hand, head, body) {
  const obj = new PIXI.Container()
  obj.sortableChildren = true

  if (hand && item_data[hand].hand) {
    obj.addChild(
      build_object(
        item_data[hand].hand,
        preview.renderer,
        preview_texture_cache,
      ),
    )
  }

  obj.addChild(
    build_object(
      body ? item_data[body].equipped : item_data[1].equipped,
      preview.renderer,
      preview_texture_cache,
    ),
  )
  obj.addChild(
    build_object(
      head ? item_data[head].equipped : item_data[0].equipped,
      preview.renderer,
      preview_texture_cache,
    ),
  )

  obj.angle = 98
  obj.scale.set(0.7)
  return obj
}

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

async function update_preview(custom_hand, custom_head, custom_body) {
  const character = account.data.characters[0]

  if (preview_character) {
    preview_character.destroy({ children: true })
  }

  preview_character = build_preview_character(
    custom_hand ? custom_hand : character.hand ? character.hand : character.inventory[0],
    custom_head ? custom_head : character.head,
    custom_body ? custom_body : character.body,
  )
  preview_character.x = preview.canvas.width / 2
  preview_character.y = preview.canvas.height / 2

  preview.stage.addChild(preview_character)
  angle_preview(mouse_x, mouse_y)
}
