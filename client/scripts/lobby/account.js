let account = null

function authorized_body(body, token=localStorage.getItem("kingdom.token")) {
  return {
    method: "POST",
    headers: { ...(body && { "Content-Type": "application/json" }), ...(token && { Authorization: `Bearer ${token}` }) },
    body: body && JSON.stringify(body),
    cache: "no-store",
    signal: AbortSignal.timeout(5000),
  }
}

async function account_request(path, body, token) {
  const response = await fetch(
    `https://gateway.kingdomcrushers.io${path}`,
    authorized_body(body, token)
  )
  if (response.status === 204) return null
  const data = await response.json().catch(() => null)
  if (!response.ok) throw Object.assign(new Error(data?.error || `Gateway request failed (${response.status})`), { status: response.status })
  if (!data) throw new Error("Invalid gateway response.")
  return data
}

function save_token(token) {
  if (!/^[a-f0-9]{64}$/i.test(token || "")) throw new Error("Invalid session token.")
  localStorage.setItem("kingdom.token", token)
}

async function guest_session() {
  const session = await account_request("/player/guest")
  const token = session.token

  save_token(token)
  return account = { ...await account_request("/players/session", undefined, token), token }
}

async function restore_session() {
  const token = localStorage.getItem("kingdom.token")
  if (!token) return guest_session()
  let session
  try {
    session = await account_request("/players/session", undefined, token)
  } catch (error) {
    if (error.status !== 401) throw error
    account = null
    localStorage.removeItem("kingdom.token")
    return guest_session()
  }
  if (!session.valid) throw new Error("Invalid session response.")
  if (!session.guest && !session.data.characters.length) {
    await account_request("/character/new", undefined, token).catch(error => { if (error.status !== 409) throw error })
    session = await account_request("/players/session", undefined, token)
  }
  return account = { ...session, token }
}

async function authenticate(email, password, endpoint) {
  const session = await account_request(endpoint, { email, password })
  save_token(session.token)
  return restore_session()
}

function login(email, password) {
  return authenticate(email, password, "/login")
}
function register(email, password) {
  return authenticate(email, password, "/register")
}

async function logout() {
  const token = localStorage.getItem("kingdom.token")
  if (token) {
    try {
      await account_request("/logout", undefined, token)
    } catch (error) { if (error.status !== 401) throw error }
  }

  localStorage.removeItem("kingdom.token")
  sessionStorage.removeItem("kingdom.token")
  localStorage.removeItem("kingdom.player_name")
  sessionStorage.removeItem("kingdom.player_name")
  sessionStorage.removeItem("kingdom.character_id")
  account = null

  return guest_session()
}

async function submit_account(event, action) {
  event.preventDefault()

  const form = event.currentTarget
  const error = form.querySelector('[role="alert"]')
  const confirm = form.elements.password_confirmation
  const button = form.querySelector('[type="submit"]')
  error.classList.add("hidden")

  if (confirm && confirm.value != form.elements.password.value) {
    confirm.setCustomValidity("Passwords do not match.")
    confirm.reportValidity()
    return
  }

  button.disabled = true
  try {
    await account_ready.catch(() => {})
    await lobby_ready
    if (action == "logout") {
      await logout()
      document.querySelector('#home [name="username"]').value = ""
    }
    else {
      let session
      const email = form.elements.email.value.trim()
      const password = form.elements.password.value

      if (action == "register") {
        session = await register(email, password)
      } else {
        session = await login(email, password)
      }
      if (!session?.valid || session.guest) {
        throw new Error("Invalid credentials.")
      }
    }
    form.reset()
    update_account_buttons()
    populate_graveyard()
    switch_screen("home")
  } catch (failure) {
    error.textContent = failure.message
    error.classList.remove("hidden")
  } finally {
    button.disabled = false
  }
}
