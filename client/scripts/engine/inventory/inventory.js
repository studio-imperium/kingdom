const hotbar = document.getElementById("hotbar")
const inventory_storage = document.getElementById("inventory_storage")
const inventory_hotbar = document.getElementById("inventory_hotbar")
const equipment_slots = document.getElementById("equipment_slots")
let head_slot
let body_slot
let dragged

function populate_hotbar() {
  for (let i = 0; i < 6; i++) {
    const slot = document.createElement("div")
    slot.className = "slot"

    const sprite = document.createElement("button")
    sprite.className = "slot_sprite"
    sprite.style.backgroundPosition = "999px -10px"

    slot.onclick = () => {
      select_slot(i)
    }

    slot.appendChild(sprite)
    hotbar.appendChild(slot)
  }
}

function create_slot(idx) {
  const slot_container = document.createElement("div")
  slot_container.className = "inventory_slot_container"

  const slot = document.createElement("div")
  slot.className = "inventory_slot"

  const sprite = document.createElement("button")
  sprite.className = "slot_sprite"
  sprite.style.backgroundPosition = "999px -10px"

  slot_container.slot = idx
  slot.slot = idx
  slot.addEventListener("pointerdown", (event) => {
    if (dragged || event.button !== 0) return
    const item_id =
      idx == 24 ? character.head : idx == 25 ? character.body : inventory[idx]
    if (!item_data[item_id] || item_id < 2) return
    event.preventDefault()
    dragged = { slot, idx, pointer_id: event.pointerId, x: event.clientX, y: event.clientY }
    slot.setPointerCapture(event.pointerId)
  })

  slot.appendChild(sprite)
  slot_container.appendChild(slot)

  return slot_container
}

function create_gear_slot(idx) {
  const container = create_slot(idx)
  const slot = container.querySelector(".inventory_slot")
  const sprite = document.createElement("button")

  slot.className = "inventory_slot gear_slot"
  sprite.className = "placeholder_sprite"
  if (idx == 24) {
    sprite.style.backgroundPosition = "0px -144px"
  } else if (idx == 25) {
    sprite.style.backgroundPosition = "-0.5px -153.5px"
  }
  slot.appendChild(sprite)
  return slot
}

function populate_inventory() {
  for (let i = 0; i < 6; i++) {
    inventory_hotbar.appendChild(create_slot(i))
  }
  for (let i = 0; i < 18; i++) {
    inventory_storage.appendChild(create_slot(i + 6))
  }

  head_slot = create_gear_slot(24)
  body_slot = create_gear_slot(25)

  equipment_slots.appendChild(head_slot)
  equipment_slots.appendChild(body_slot)
}

function reset_hotbar() {
  for (let slot of hotbar.childNodes) {
    let sprite = slot.querySelector(".slot_sprite")
    sprite.style.backgroundPosition = "999px -10px"
    slot.className = "slot"
  }
}

function reset_inventory() {
  for (let slot of inventory_storage.childNodes) {
    let sprite = slot.querySelector(".slot_sprite")
    sprite.style.backgroundPosition = "999px -10px"
  }
  for (let slot of inventory_hotbar.childNodes) {
    let sprite = slot.querySelector(".slot_sprite")
    sprite.style.backgroundPosition = "999px -10px"
  }
  for (let slot of equipment_slots.childNodes) {
    let sprite = slot.querySelector(".slot_sprite")
    let placeholder = slot.querySelector(".placeholder_sprite")
    sprite.style.backgroundPosition = "999px -10px"
    placeholder.classList.remove("hidden")
  }
}

function get_inventory_slot(slot) {
  if (slot < 6) {
    return inventory_hotbar.querySelector(`[slot="${slot}"]`)
  } else {
    return inventory_storage.querySelector(`[slot="${slot}"]`)
  }
}

function get_hotbar_slot(slot) {
  if (slot < 6) {
    return hotbar.querySelectorAll(".slot")[slot]
  }
}

function set_slot(slot_node, item_id) {
  if (!slot_node) {
    return
  }
  let data = item_data[item_id]
  let sprite_node = slot_node.querySelector(".slot_sprite")

  sprite_node.style.backgroundPosition = `${-data.sprite.x}px ${-data.sprite.y}px`
}

function set_gear_slot(slot_node, item_id) {
  let data = item_data[item_id]
  let sprite = slot_node.querySelector(".slot_sprite")
  let placeholder = slot_node.querySelector(".placeholder_sprite")
  sprite.style.backgroundPosition = `${-data.sprite.x}px ${-data.sprite.y}px`

  if (item_id < 2) {
    placeholder.classList.remove("hidden")
    sprite.style.backgroundPosition = "999px -10px"
  } else {
    placeholder.classList.add("hidden")
  }
}

function refresh_inventory(_inventory, hand, head, body) {
  inventory = _inventory

  reset_hotbar()
  reset_inventory()
  hotbar.childNodes[hand].className = "slot selected"

  set_gear_slot(head_slot, head)
  set_gear_slot(body_slot, body)

  for (let slot of Object.keys(inventory)) {
    let node = get_inventory_slot(slot)
    set_slot(node, inventory[slot])

    if (slot < 6) {
      set_slot(get_hotbar_slot(slot), inventory[slot])
    }
  }
}

populate_hotbar()
populate_inventory()

function cancel_inventory_drag() {
  if (!dragged) return
  const drag = dragged
  dragged = null
  drag.sprite?.remove()
  drag.slot.classList.remove("dragging")
  drag.target?.classList.remove("hovered")
  if (drag.slot.hasPointerCapture(drag.pointer_id)) drag.slot.releasePointerCapture(drag.pointer_id)
}

document.addEventListener("pointermove", (event) => {
  if (!dragged || event.pointerId !== dragged.pointer_id) return
  if (!dragged.sprite) {
    // Small finger movements while tapping should not move or drop an item.
    if (Math.hypot(event.clientX - dragged.x, event.clientY - dragged.y) < 6) return
    dragged.sprite = dragged.slot.querySelector(".slot_sprite").cloneNode()
    dragged.sprite.classList.add("inventory_drag")
    document.body.appendChild(dragged.sprite)
    dragged.slot.classList.add("dragging")
  }
  dragged.sprite.style.left = `${event.clientX}px`
  dragged.sprite.style.top = `${event.clientY}px`
  dragged.target?.classList.remove("hovered")
  dragged.target = document.elementFromPoint(event.clientX, event.clientY)?.closest(".inventory_slot")
  if (dragged.target !== dragged.slot) dragged.target?.classList.add("hovered")
})

document.addEventListener("pointerup", (event) => {
  if (!dragged || event.pointerId !== dragged.pointer_id) return
  const drag = dragged
  const element = document.elementFromPoint(event.clientX, event.clientY)
  const target = element?.closest(".inventory_slot")
  cancel_inventory_drag()
  if (!drag.sprite || !element || drag.slot.closest(".hidden")) return
  if (target && target !== drag.slot) change_inventory(Number(target.slot), drag.idx)
  else if (!target) drop_item(drag.idx)
})

for (const type of ["pointercancel", "lostpointercapture"]) {
  document.addEventListener(type, (event) => {
    if (event.pointerId === dragged?.pointer_id) cancel_inventory_drag()
  })
}
window.addEventListener("blur", cancel_inventory_drag)
window.addEventListener("resize", cancel_inventory_drag)
document.addEventListener("visibilitychange", () => {
  if (document.hidden) cancel_inventory_drag()
})
