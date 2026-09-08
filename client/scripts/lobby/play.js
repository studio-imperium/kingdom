async function play(event) {
  event.preventDefault()
  const form = event.currentTarget, button = form.querySelector('[type="submit"]'), message = form.querySelector('[role="alert"]')
  button.disabled = true
  message.classList.add("hidden")
  try {
    await account_ready
    if (!account_session) return
    await restore_session()
    update_account_buttons()
    const input = form.elements.username
    const name = input.value.trim()
    input.setCustomValidity(!name || new TextEncoder().encode(name).length > 255 ? "Enter a name up to 255 bytes long." : "")
    if (!input.reportValidity()) return
    const servers = form.elements.servers
    if (servers.disabled || !servers.value) return
    sessionStorage.setItem("kingdom.player_name", name)
    sessionStorage.setItem("kingdom.server", servers.value)
    sessionStorage.setItem("kingdom.character_id", account_session.guest ? "0" : account_session.data.characters[0].id)
    location.href = "/game.html"
  } catch (error) {
    message.textContent = error.message
    message.classList.remove("hidden")
  } finally {
    button.disabled = false
  }
}
