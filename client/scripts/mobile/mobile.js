let mobile_controls = null
let joysticks = null
const desktop_angle_preview = angle_preview
const mobile_query = window.matchMedia("(any-pointer: coarse)")

async function configure_mobile() {
  if (document.body.classList.contains("game_over")) return
  const is_mobile = mobile_query.matches
  angle_preview = is_mobile ? () => {} : desktop_angle_preview
  if (is_mobile && preview_character) preview_character.angle = 98

  const main_interface = document.getElementById("lobby_main")
  if (!main_interface) {
    if (is_mobile) document.getElementById("controls_tooltip").classList.add("hidden")
    document.getElementById("backpack_button").classList.toggle("hidden", !is_mobile)
    document.getElementById("home_button").classList.toggle("hidden", !is_mobile)
    if (is_mobile && !joysticks) joysticks = { left: create_joystick("left", app.canvas), right: create_joystick("right", app.canvas) }
    mobile_controls = is_mobile ? joysticks : null
    attacking = false
    return
  }

  const utilities_list = document.getElementById("utilities")
  utilities_list.classList.toggle("column", is_mobile)
  document.getElementById("map_button").classList.toggle("hidden", is_mobile)
  document.getElementById("login_preview").classList.toggle("hidden", is_mobile)
  document.getElementById("register_preview").classList.toggle("hidden", is_mobile)
  const buttons = is_mobile ? utilities_list : main_interface
  buttons.appendChild(document.getElementById("leaderboard_button"))
  buttons.appendChild(document.getElementById("graveyard_button"))
}

window.addEventListener("resize", () => {
  if (!app.renderer || document.body.classList.contains("game_over")) return
  app.resize()
  configure_mobile()
})
mobile_query.addEventListener("change", () => { if (app.renderer) configure_mobile() })
