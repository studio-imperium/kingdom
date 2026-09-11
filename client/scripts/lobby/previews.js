const preview_canvas = document.getElementById("character_preview")
const preview = new PIXI.Application()
let preview_character = null
const preview_texture_cache = {}
let mouse_x = 0, mouse_y = 0

function angle_preview(x, y) {
  if (!preview_character) return
  const { left, top, width, height } = preview_canvas.getBoundingClientRect()
  const center_x = left + width / 2, center_y = top + height / 2

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

function update_preview(custom_hand, custom_head, custom_body) {
  if (!preview.renderer || !account?.data) return
  const character = account.data.characters[0]

  if (preview_character) {
    preview_character.destroy({ children: true })
  }

  preview_character = build_preview_character(
    custom_hand ?? character.inventory[character.hand],
    custom_head ?? character.head,
    custom_body ?? character.body,
  )
  preview_character.x = preview.canvas.width / 2
  preview_character.y = preview.canvas.height / 2

  preview.stage.addChild(preview_character)
  angle_preview(mouse_x, mouse_y)
}
