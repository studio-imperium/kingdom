async function play(event) {
  event.preventDefault()
  const form = event.currentTarget
  const button = form.querySelector('[type="submit"]'), error = form.querySelector('[role="alert"]')
  const name = form.elements.username.value.trim(), servers = form.elements.servers
  error.classList.add("hidden")
  form.elements.username.setCustomValidity(!name || new TextEncoder().encode(name).length > 255 ? "Enter a name of 1–255 UTF-8 bytes." : "")
  if (!form.elements.username.reportValidity()) return

  button.disabled = true
  try {
    if (servers.disabled || !servers.value) throw new Error("Select an available server.")
    await account_ready.catch(() => {})
    await restore_session()
    localStorage.setItem("kingdom.player_name", name)
    sessionStorage.setItem("kingdom.player_name", name)
    sessionStorage.setItem("kingdom.server", servers.value)
    sessionStorage.setItem("kingdom.character_id", account.guest ? "0" : account.data.characters[0].id)
    location.href = "/game.html"
  } catch (failure) {
    error.textContent = failure.message
    error.classList.remove("hidden")
  } finally {
    button.disabled = false
  }
}
