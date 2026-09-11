const preview_canvas = document.getElementById("inventory_preview")
const preview = new PIXI.Application()
const preview_ready = init_preview()
let preview_character = null
const preview_texture_cache = {}

function angle_preview(x, y) {
  if (!preview_character) return
  center_x =
    preview_canvas.getBoundingClientRect().left + preview_canvas.width / 2
  center_y =
    preview_canvas.getBoundingClientRect().top + preview_canvas.height / 2

  const dx = x - center_x
  const dy = y - center_y
  preview_character.angle = Math.atan2(dy, dx) * (180 / Math.PI) + 90
}

document.addEventListener("dragover", (e) => {
  mouse_x = e.clientX
  mouse_y = e.clientY
  angle_preview(mouse_x, mouse_y)
})

document.addEventListener("mousemove", (e) => {
  mouse_x = e.clientX
  mouse_y = e.clientY
  angle_preview(mouse_x, mouse_y)
})

window.__PIXI_DEVTOOLS__ = {
  app: preview,
}

async function init_preview() {
  await preview.init({
    canvas: preview_canvas,
    background: "#1f1f1f",
    resizeTo: document.querySelector(".inventory_top"),
    width: 256,
    height: 256,
    useContextAlpha: false,
    antialias: true,
    autoDensity: true,
    resolution: 1,
  })
  preview.canvas.style.imageRendering = "pixelated"
  preview.renderer.on("resize", () => {
    if (preview_character) preview_character.position.set(preview.screen.width / 2, preview.screen.height / 2)
  })
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

async function update_preview() {
  if (!character) return
  await preview_ready

  if (preview_character) {
    preview_character.destroy({ children: true })
  }

  preview_character = build_preview_character(
    character.hand,
    character.head,
    character.body,
  )
  preview_character.x = preview.canvas.width / 2
  preview_character.y = preview.canvas.height / 2

  preview.stage.addChild(preview_character)
  angle_preview(mouse_x, mouse_y)
}
