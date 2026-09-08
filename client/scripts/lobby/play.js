document.querySelector(".lobby_actions .button").addEventListener("click", async () => {
  await account_ready
  if (!account_session) return
  await restore_session()
  const input = document.querySelector(".text_input")
  const name = input.value.trim()
  input.setCustomValidity(!name || new TextEncoder().encode(name).length > 255 ? "Enter a name up to 255 bytes long." : "")
  if (!input.reportValidity()) return
  const servers = document.getElementById("servers")
  if (servers.disabled || !servers.value) return
  sessionStorage.setItem("kingdom.player_name", name)
  sessionStorage.setItem("kingdom.server", servers.value)
  if (account_session.guest) {
    sessionStorage.setItem("kingdom.character_id", "0")
  } else {
    if (!account_session.data.characters.length) {
      await account_request("/character/new", undefined, account_session.token)
      await restore_session()
    }
    const characters = account_session.data.characters
    const selected = characters.find(character => character.id === sessionStorage.getItem("kingdom.character_id")) || characters[0]
    sessionStorage.setItem("kingdom.character_id", selected.id)
  }
  location.href = "/game.html"
})
