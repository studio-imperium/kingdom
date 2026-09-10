let leaderboard_loading = false

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
    await document.fonts.load('32px "myriad-pro"')
    const fragment = document.createDocumentFragment()
    characters.forEach((character, index) => {
      const row = document.createElement("tr")
      for (const value of [String(index + 1), character_preview(character)]) {
        const cell = document.createElement("td")
        cell.append(value)
        row.append(cell)
      }
      if (character.dead) {
        const dead = document.createElement("small")
        dead.textContent = "Dead"
        row.lastChild.append(dead)
        row.lastChild.title = "Character and level at death"
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
