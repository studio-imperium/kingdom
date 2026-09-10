let leaderboard = []
let ping = {}

async function ping_servers(servers) {
  for (const server of servers) {
    let start = performance.now()

    try {
      await fetch("https://" + server.address + "/assets/tiles.json", {
        method: "HEAD",
        cache: "no-store",
        redirect: "error",
        signal: AbortSignal.timeout(3000)
      })
      ping[server.address] = Math.round(performance.now() - start)
    } catch {
      ping[server.address] = 999
    }
  }
}

const server_api = "https://gateway.kingdomcrushers.io/gameservers"
const select = document.getElementById("servers")
let server;

async function init_servers() {
  const response = await fetch(server_api)
  if (!response.ok) throw new Error("Unable to load servers.")
  const servers = await response.json()
  await ping_servers(servers)

  const on_change = () => {
    let server_ping = ping[select.value]

    server = select.value

    select.style.color = (
      server_ping < 75 ? "#76dc8c" :
        server_ping < 200 ? "#ffb347" :
          "#ff6b6b")
  }
  select.addEventListener("change", on_change)

  select.innerHTML = ""
  for (const server of servers) {
    select.add(new Option(server.name + " " + server.players_online + "/" + server.max_players, server.address))
  }
  select.disabled = !servers.length
  if (!servers.length) select.add(new Option("No servers available", ""))
  on_change()
}



async function fetch_leaderboards() {
  leaderboard = await (await fetch("https://gateway.kingdomcrushers.io/leaderboard")).json()
  const leaderboard_list = document.getElementById("leaderboard_list")
  leaderboard_list.replaceChildren()

  let i = 1;

  const make_entry = (character) => {
    const container = document.createElement("div")
    const rank = document.createElement("p")
    const name = document.createElement("label")
    const group = document.createElement("group")
    const slots = document.createElement("div")
    const preview = document.createElement("canvas")

    if (character.dead) {
      name.classList.add("error")
    }

    rank.innerHTML = i + "."
    name.textContent = character.username + ", level " + character.level
    render_character(preview, 135, character.weapon, character.head, character.body)

    container.appendChild(rank)
    for (var item_id of [character.weapon, character.head, character.body]) {
      let data = item_data[item_id]

      const slot = document.createElement("div")
      slot.className = "slot"

      if (item_id > 1) {
        const sprite = document.createElement("button")
        sprite.className = "slot_sprite"
        sprite.style.backgroundPosition = `${-data.sprite.x}px ${-data.sprite.y}px`

        slot.append(sprite)
      }
      slots.appendChild(slot)
    }
    group.appendChild(name)
    group.appendChild(slots)
    container.appendChild(group)
    container.appendChild(preview)

    return container;
  }

  for (const character of leaderboard) {
    leaderboard_list.appendChild(make_entry(character))
    i++
  }
}



function populate_graveyard() {
  const graveyard = account.data.graveyard
  const graveyard_list = document.getElementById("graveyard_list")
  graveyard_list.replaceChildren()

  let i = 1;

  const make_entry = (character) => {
    const weapon = character.inventory[character.hand] || 0
    const container = document.createElement("div")
    const rank = document.createElement("p")
    const name = document.createElement("label")
    const group = document.createElement("group")
    const slots = document.createElement("div")
    const preview = document.createElement("canvas")

    if (character.dead) {
      name.classList.add("error")
    }

    rank.innerHTML = i + "."
    name.textContent = character.username + ", level " + character.level
    render_character(preview, 135, weapon, character.head, character.body)

    container.appendChild(rank)
    for (var item_id of [weapon, character.head, character.body]) {
      let data = item_data[item_id]

      const slot = document.createElement("div")
      slot.className = "slot"

      if (item_id > 1) {
        const sprite = document.createElement("button")
        sprite.className = "slot_sprite"
        sprite.style.backgroundPosition = `${-data.sprite.x}px ${-data.sprite.y}px`

        slot.append(sprite)
      }
      slots.appendChild(slot)
    }
    group.appendChild(name)
    group.appendChild(slots)
    container.appendChild(group)
    container.appendChild(preview)

    return container;
  }

  for (const character of graveyard) {
    graveyard_list.appendChild(make_entry(character))
    i++
  }
}
