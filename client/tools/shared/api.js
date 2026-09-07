async function loadAsset(name) {
  const res = await fetch(`/assets/${name}`)
  if (!res.ok) throw new Error(`load ${name}: ${res.status}`)
  return res.json()
}

async function saveAsset(name, data) {
  const res = await fetch(`/save/${name}`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(data),
  })
  if (!res.ok) {
    const txt = await res.text()
    throw new Error(`save ${name}: ${res.status} ${txt}`)
  }
  return res.json()
}

function attachSaveButton(button, statusEl, name, getData) {
  button.addEventListener("click", async () => {
    const original = button.textContent
    button.disabled = true
    button.textContent = "Saving…"
    try {
      const data = getData()
      await saveAsset(name, data)
      if (statusEl) {
        statusEl.textContent = `Saved ${name} at ${new Date().toLocaleTimeString()}`
        statusEl.style.color = "var(--good)"
      }
    } catch (e) {
      if (statusEl) {
        statusEl.textContent = "Error: " + e.message
        statusEl.style.color = "var(--danger)"
      } else {
        alert(e.message)
      }
    } finally {
      button.disabled = false
      button.textContent = original
    }
  })
}
