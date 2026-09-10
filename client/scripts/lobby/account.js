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

async function account_request(path, body) {
  const response = await fetch(
    `https://gateway.kingdomcrushers.io${path}`,
    authorized_body(body)
  )
  return await response.json()
}

async function guest_session() {
  const session = await account_request("/player/guest")
  const token = session.token

  localStorage.setItem("kingdom.token", token)
  account = session

  document.getElementById("login_button").classList.remove("hidden")
  document.getElementById("logout_button").classList.add("hidden")

  return account
}

async function restore_session() {
  let session = await account_request("/players/session")

  if (!session.valid) {
    localStorage.removeItem("kingdom.token")
    return guest_session()
  }
  if (!session.guest && !session.data.characters.length) {
    await account_request("/character/new")
    session = await account_request("/players/session")
  }
  account = session
  document.getElementById("login_button").classList.add("hidden")
  document.getElementById("logout_button").classList.remove("hidden")
  await populate_graveyard()
}

async function authenticate(email, password, endpoint) {
  const session = await account_request(endpoint, { email, password })
  localStorage.setItem("kingdom.token", session.token)
  return restore_session()
}

function login(email, password) {
  return authenticate(email, password, "/login")
}
function register(email, password) {
  return authenticate(email, password, "/register")
}

async function logout() {
  let token
  if (token = localStorage.getItem("kingdom.token")) {
    try {
      await account_request("/logout", undefined, token)
    } catch (error) {}
  }

  localStorage.removeItem("kingdom.token")
  sessionStorage.removeItem("kingdom.token")
  account = null

  return guest_session()
}

async function submit_account(event, action) {
  event.preventDefault()

  const form = event.currentTarget
  const error = form.querySelector('[role="alert"]')
  const confirm = form.elements.password_confirmation

  if (confirm && confirm.value != form.elements.password.value) {
    confirm.setCustomValidity("Passwords do not match.")
    confirm.reportValidity()
    return
  }

  try {
    if (action == "logout") {
      await logout()
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
    switch_screen("home")
  } catch (failure) {
    error.textContent = failure.message
    error.classList.remove("hidden")
  }
}
