let current_tab = null

const inventoryNode = document.getElementById("inventory")

function close_tab() {
  if (current_tab != null) {
    current_tab.classList.add("hidden")
    current_tab = null
  }
}
function open_tab(tab) {
  current_tab = tab
  current_tab.classList.remove("hidden")
}

document.addEventListener("keydown", (e) => {
  if (chat_focused) {
    return
  }
  if (e.key == "Escape") {
    close_tab()
  }
  if (e.key == "e") {
    if (current_tab == null) {
      open_tab(inventoryNode)
      const rect = preview_canvas.getBoundingClientRect()
      preview.renderer.resize(rect.width, rect.height)
      update_preview()
    } else {
      close_tab()
    }
  }
})
