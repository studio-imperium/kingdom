let game_over = false

async function show_death() {
  if (game_over) return
  game_over = true
  socket?.close(1000, "Character died")
  app.stop()
  app.resizeTo = null
  app.canvas.style.width = "100vw"
  app.canvas.style.height = "100vh"
  attacking = false
  close_tab()
  document.activeElement?.blur()
  document.getElementById("connecting_screen")?.remove()
  document.body.classList.add("game_over")

  const { hand, head, body } = character
  document.getElementById("death_name").textContent = (sessionStorage.getItem("kingdom.player_name") || "Guest") + ", level " + level
  // const slots = document.getElementById("death_gear")
  // for (const item of [hand, head, body]) {
  //   const slot = document.createElement("div")
  //   slot.className = "slot"
  //   if (item > 1) {
  //     const sprite = document.createElement("button")
  //     sprite.type = "button"
  //     sprite.tabIndex = -1
  //     sprite.className = "slot_sprite"
  //     slot.appendChild(sprite)
  //     set_slot(slot, item)
  //   }
  //   slots.appendChild(slot)
  // }
  await preview_ready
  preview.stop()
  render_character(document.getElementById("death_preview"), 135, hand, head, body)
  document.getElementById("death_screen").classList.remove("hidden")
  document.getElementById("death_continue").focus({ preventScroll: true })
}

for (const type of ["keydown", "keyup"]) {
  document.addEventListener(type, event => {
    if (game_over) event.stopImmediatePropagation()
  }, true)
}
