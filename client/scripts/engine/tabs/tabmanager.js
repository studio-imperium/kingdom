let current_tab = null

const inventoryNode = document.getElementById("inventory")
const menuNode = document.getElementById("menu")

function toggle_inventory() {
  if (current_tab != inventoryNode) {
    open_tab(inventoryNode)
    const rect = preview_canvas.getBoundingClientRect()
    preview.renderer.resize(rect.width, rect.height)
    update_preview()
  } else {
    close_tab()
  }
}

function close_tab() {
  cancel_inventory_drag()
  if (current_tab != null) {
    current_tab.classList.add("hidden")
    current_tab = null
  }
}
function open_tab(tab) {
  close_tab()
  current_tab = tab
  current_tab.classList.remove("hidden")
}

document.addEventListener("keydown", (e) => {
  if (chat_focused || e.repeat) {
    return
  }
  if (e.key == "Escape") {
    current_tab ? close_tab() : open_tab(menuNode)
    const tooltip = document.getElementById("controls_tooltip")
    if (tooltip.textContent == "[ESC] to open menu") tooltip.classList.add("hidden")
  }
  if (e.key == "e") {
    toggle_inventory()
    document.getElementById("controls_tooltip").textContent = "[ESC] to open menu"
  }
})
