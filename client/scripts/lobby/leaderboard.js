let leaderboard_loading = false

function leaderboard_preview(character) {
  const preview = new PIXI.Container(), player = new PIXI.Container()
  preview.addChild(player)
  player.position.set(48, 48)
  player.scale.set(40 / 64)
  try {
    for (const part of [item_data[character.weapon]?.hand, item_data[character.body || 1]?.equipped, item_data[character.head || 0]?.equipped]) {
      if (part) player.addChild(build_object(part, lobby_preview.renderer, lobby_texture_cache))
    }
    const canvas = lobby_preview.renderer.extract.canvas({ target: preview, frame: new PIXI.Rectangle(0, 0, 96, 96) })
    canvas.setAttribute("role", "img")
    canvas.setAttribute("aria-label", `${character.username || "Unnamed"} character`)
    return canvas
  } finally {
    preview.destroy({ children: true })
  }
}

async function load_leaderboard() {
  if (leaderboard_loading) return
  leaderboard_loading = true
  const rows = document.getElementById("leaderboard_rows"), status = document.getElementById("leaderboard_status"), retry = document.getElementById("leaderboard_retry")
  rows.replaceChildren()
  status.textContent = "Loading leaderboard…"
  retry.classList.add("hidden")
  try {
    const response = await fetch("https://gateway.kingdomcrushers.io/leaderboard", { cache: "no-store", signal: AbortSignal.timeout(5000) })
    if (!response.ok) throw new Error("Unable to load leaderboard. Please try again.")
    const characters = await response.json()
    await lobby_ready
    const fragment = document.createDocumentFragment()
    characters.forEach((character, index) => {
      const row = document.createElement("tr")
      for (const value of [String(index + 1), character.username || "Unnamed", leaderboard_preview(character), character.level]) {
        const cell = document.createElement("td")
        cell.append(value)
        row.append(cell)
      }
      if (character.dead) {
        const dead = document.createElement("small")
        dead.textContent = "Dead"
        row.lastChild.append(dead)
        row.lastChild.title = "Level at death"
      }
      fragment.append(row)
    })
    rows.replaceChildren(fragment)
    status.textContent = characters.length ? "" : "No characters yet."
  } catch (error) {
    status.textContent = "Unable to load leaderboard. Please try again."
    retry.classList.remove("hidden")
    console.error("Unable to load leaderboard:", error)
  } finally {
    leaderboard_loading = false
  }
}
