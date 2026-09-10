const screens = document.querySelectorAll(".screen")

function switch_screen(screen_name = "home") {
  const new_screen = document.getElementById(screen_name)
  const slot = new_screen.querySelector(".preview_slot")

  if (slot) {
    slot.appendChild(document.getElementById("character_preview"))

    if (screen_name == "home") {
      update_preview()
    }
    if (screen_name == "login") {
      update_preview(20,26,27)
    }
    if (screen_name == "register") {
      update_preview(17,22,23)
    }
  }

  for (const screen of screens) {
    screen.classList.toggle("hidden", screen != new_screen)
  }
}
