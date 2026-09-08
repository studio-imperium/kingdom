const lobby_preview = new PIXI.Application()
const lobby_player = new PIXI.Container()
const lobby_texture_cache = {}

function update_lobby_character() {
  const character = account_session?.guest ? { hand: 0, inventory: { 0: 8 }, body: 1, head: 0 } : account_session?.data.characters[0]
  document.getElementById("lobby_preview").classList.toggle("hidden", !character)
  if (!lobby_preview.renderer) return
  lobby_player.removeChildren().forEach(part => part.destroy({ children: true }))
  if (!character) return
  for (const part of [item_data[character.inventory[character.hand]]?.hand, item_data[character.body || 1].equipped, item_data[character.head || 0].equipped]) {
    if (part) lobby_player.addChild(build_object(part, lobby_preview.renderer, lobby_texture_cache))
  }
}

async function init_character_previews() {
  const canvas = document.getElementById("lobby_preview")
  await lobby_preview.init({ canvas, width: 256, height: 256, backgroundAlpha: 0, resolution: 1, antialias: true })
  lobby_player.position.set(128, 128)
  lobby_player.scale.set(48 / 64)
  lobby_preview.stage.addChild(lobby_player)
  document.addEventListener("pointermove", ({ clientX, clientY }) => {
    const { left, top, width, height } = canvas.getBoundingClientRect()
    lobby_player.angle = Math.atan2(clientY - top - height / 2, clientX - left - width / 2) * 180 / Math.PI + 90
  })
  update_lobby_character()
}
