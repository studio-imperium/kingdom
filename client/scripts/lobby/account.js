const session_token_key = "kingdom.session_token"
const guest_token_key = "kingdom.guest_token"
let account_session = null

async function account_request(path, body, token) {
  const response = await fetch(`https://gateway.kingdomcrushers.io${path}`, {
    method: "POST",
    headers: { ...(body && { "Content-Type": "application/json" }), ...(token && { Authorization: `Bearer ${token}` }) },
    body: body && JSON.stringify(body),
    cache: "no-store",
    signal: AbortSignal.timeout(5000),
  })
  if (response.status === 204) return null
  const data = await response.json().catch(() => null)
  if (!response.ok) throw Object.assign(new Error(data?.error || `Gateway request failed (${response.status})`), { status: response.status })
  if (!data) throw new Error("Invalid gateway response")
  return data
}

// Only expose a session after the gateway has validated the cached token.
async function guest_session() {
  const { token } = await account_request("/player/guest")
  if (typeof token !== "string" || !/^[a-f0-9]{64}$/i.test(token)) throw new Error("Invalid session token response")
  sessionStorage.setItem(guest_token_key, token)
  return account_session = { valid: true, guest: true, email: "", token }
}

async function restore_session() {
  account_session = null
  const storage = localStorage.getItem(session_token_key) ? localStorage : sessionStorage
  const key = storage === localStorage ? session_token_key : guest_token_key
  const token = storage.getItem(key)
  if (!token) return guest_session()
  try {
    let session = await account_request("/players/session", undefined, token)
    if (session.valid !== true) throw new Error("Invalid session response")
    if (!session.guest && !session.data.characters.length) {
      await account_request("/character/new", undefined, token).catch(error => { if (error.status !== 409) throw error })
      session = await account_request("/players/session", undefined, token)
    }
    return account_session = { ...session, token }
  } catch (error) {
    if (error.status !== 401) throw error // Keep the token on network/server failures.
    storage.removeItem(key)
    return guest_session()
  }
}

async function authenticate(email, password, endpoint) {
  await account_ready
  const { token } = await account_request(endpoint, { email, password })
  if (typeof token !== "string" || !/^[a-f0-9]{64}$/i.test(token)) throw new Error("Invalid session token response")
  localStorage.setItem(session_token_key, token)
  return restore_session()
}

function login(email, password) { return authenticate(email, password, "/login") }
function register(email, password) { return authenticate(email, password, "/register") }

async function logout() {
  await account_ready
  const token = account_session?.token
  try {
    if (token) await account_request("/logout", undefined, token)
  } catch (error) {
    if (error.status !== 401) throw error
  }
  localStorage.removeItem(session_token_key)
  sessionStorage.removeItem(guest_token_key)
  localStorage.removeItem("kingdom.player_name")
  sessionStorage.removeItem("kingdom.player_name")
  account_session = null
  return guest_session()
}

// Starts as soon as this script loads; callers can await account_ready.
const account_ready = restore_session().catch(error => {
  console.error("Unable to restore session:", error)
  return null
})
