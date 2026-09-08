async function load_servers() {
  const select = document.getElementById("servers")
  const update_color = () => select.style.setProperty("--ping-color", select.selectedOptions[0].style.getPropertyValue("--ping-color"))
  select.disabled = true
  select.addEventListener("change", update_color)
  try {
    const response = await fetch("https://gateway.kingdomcrushers.io/gameservers", { signal: AbortSignal.timeout(5000) })
    if (!response.ok) throw new Error(response.status)
    const servers = await response.json()
    select.replaceChildren(...servers.map(({ name, address, players_online, max_players }) => new Option(`${name} (${players_online ?? "0"}/${max_players})`, address)))
    if (!servers.length) return select.replaceChildren(new Option("No servers available", ""))
    select.disabled = false
    await Promise.all(servers.map(async (server, i) => {
      const option = select.options[i]
      const url = new URL(/^[a-z]+:\/\//i.test(server.address) ? server.address : `https://${server.address}`)
      url.protocol = url.protocol.replace("wss:", "https:").replace("ws:", "http:")
      const probe = () => fetch(new URL("/assets/tiles.json", url), { method: "HEAD", cache: "no-store", redirect: "error", signal: AbortSignal.timeout(3000) })
      let quality = "bad"
      try {
        await probe() // Warm the connection before measuring the HTTP round trip.
        const start = performance.now(), result = await probe()
        if (!result.ok) throw new Error(result.status)
        const ping = Math.round(performance.now() - start)
        quality = ping > 200 ? "bad" : ping > 100 ? "neutral" : "good"
      } catch {}
      option.style.setProperty("--ping-color", `var(--ping-${quality})`)
      update_color()
    }))
  } catch {
    select.replaceChildren(new Option("Unable to load servers", ""))
    select.disabled = true
  }
}
