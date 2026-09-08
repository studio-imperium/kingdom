function switch_screen(name) {
  document.querySelectorAll(".lobby_screen").forEach(screen => screen.classList.toggle("hidden", screen.id !== `${name}_screen`))
  if (name === "leaderboard") load_leaderboard()
  if (name === "login" || name === "register") document.querySelector(`#${name}_screen .enemy_preview_slot`).append(document.getElementById("enemy_preview"))
}
