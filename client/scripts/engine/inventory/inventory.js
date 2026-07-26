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
  slot.draggable = true

  slot.addEventListener("dragstart", (event) => {
    const sprite = slot.querySelector(".slot_sprite")
    const drag = sprite.cloneNode()
    const item_id =
      idx == 24 ? character.head : idx == 25 ? character.body : inventory[idx]
    const data = item_data[item_id]

    console.log(item_id)

    if (!data || item_id < 2) {
      event.preventDefault()
      return
    }

    dragged = idx
    event.target.classList.add("dragging")

    drag.style.width = "40px"
    drag.style.height = "40px"
    drag.style.backgroundSize = "2048px 2048px"
    drag.style.backgroundPosition = `${0}px ${-90 * 4}px`

    document.body.appendChild(drag)

    event.dataTransfer.setDragImage(drag, 20, 20)

    requestAnimationFrame(() => {
      drag.remove()
    })
  })
  slot.addEventListener("dragend", (event) => {
    event.target.classList.remove("dragging")
  })
  slot.addEventListener("dragover", (event) => {
    event.preventDefault()
  })
  slot.addEventListener("dragenter", (event) => {
    event.target.classList.add("hovered")
  })
  slot.addEventListener("dragleave", (event) => {
    event.target.classList.remove("hovered")
  })
  slot.addEventListener("drop", (event) => {
    event.preventDefault()
    event.stopPropagation()

    event.target.classList.remove("hovered")
    change_inventory(idx, dragged)
    dragged = null
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

document.body.addEventListener("dragover", (event) => {
  if (dragged != null) event.preventDefault()
})
document.body.addEventListener("drop", (event) => {
  if (dragged == null) return
  event.preventDefault()
  drop_item(dragged)
  dragged = null
})
