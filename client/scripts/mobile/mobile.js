let mobile_controls = null

async function configure_mobile() {
  const is_mobile = window.matchMedia("(any-pointer: coarse)").matches
  if (is_mobile) {
    const main_interface = document.getElementById("lobby_main")
    if (!main_interface) {
      if (!mobile_controls) mobile_controls = { left: create_joystick("left", app.canvas), right: create_joystick("right", app.canvas) }
      return
    }
    const utilities_list = document.getElementById("utilities")
    const map_button = document.getElementById("map_button")
    const leaderboard_button = document.getElementById("leaderboard_button")
    const graveyard_button = document.getElementById("graveyard_button")
    const login_preview = document.getElementById("login_preview")
    const register_preview = document.getElementById("register_preview")

    utilities_list.classList.add("column")
    map_button.classList.add("hidden")
    login_preview.classList.add("hidden")
    register_preview.classList.add("hidden")
    main_interface.removeChild(leaderboard_button)
    main_interface.removeChild(graveyard_button)
    utilities_list.appendChild(leaderboard_button)
    utilities_list.appendChild(graveyard_button)
  }
}
