function create_joystick(side, surface) {
  const zone = document.createElement("div")
  zone.className = "joystick joystick_" + side
  zone.setAttribute("aria-label", side === "left" ? "Move" : "Aim and attack")
  zone.innerHTML = '<div class="joystick_line"></div><div class="joystick_base"></div><div class="joystick_thumb"></div>'
  document.body.appendChild(zone)

  const line = zone.querySelector(".joystick_line")
  const thumb = zone.querySelector(".joystick_thumb")
  const stick = { x: 0, y: 0 }
  let pointer_id = null
  let origin_x = 0, origin_y = 0
  surface.style.touchAction = "none"

  function move(event) {
    if (event.pointerId !== pointer_id) return
    const radius = zone.clientWidth / 2 - thumb.offsetWidth / 2
    const dx = event.clientX - origin_x
    const dy = event.clientY - origin_y
    const distance = Math.hypot(dx, dy)
    const scale = distance > radius ? radius / distance : 1
    const x = dx * scale, y = dy * scale

    stick.x = distance > radius * 0.15 ? x / radius : 0
    stick.y = distance > radius * 0.15 ? y / radius : 0
    thumb.style.transform = `translate(-50%, -50%) translate(${x}px, ${y}px)`
    line.style.width = Math.hypot(x, y) + "px"
    line.style.transform = `rotate(${Math.atan2(y, x)}rad)`
  }

  function reset() {
    const id = pointer_id
    pointer_id = null
    zone.classList.remove("active")
    stick.x = stick.y = 0
    thumb.style.transform = "translate(-50%, -50%)"
    line.style.width = "0px"
    if (id !== null && surface.hasPointerCapture(id)) surface.releasePointerCapture(id)
  }

  surface.addEventListener("pointerdown", (event) => {
    if (!mobile_controls || pointer_id !== null || event.button !== 0) return
    const rect = surface.getBoundingClientRect()
    if ((event.clientX < rect.left + rect.width / 2) !== (side === "left")) return
    event.preventDefault()
    pointer_id = event.pointerId
    origin_x = event.clientX
    origin_y = event.clientY
    zone.style.left = origin_x + "px"
    zone.style.top = origin_y + "px"
    zone.classList.add("active")
    surface.setPointerCapture(pointer_id)
    move(event)
  })
  surface.addEventListener("pointermove", move)
  for (const type of ["pointerup", "pointercancel", "lostpointercapture"]) {
    surface.addEventListener(type, (event) => {
      if (event.pointerId === pointer_id) reset()
    })
  }
  window.addEventListener("blur", reset)
  window.addEventListener("resize", reset)
  document.addEventListener("visibilitychange", () => {
    if (document.hidden) reset()
  })
  return stick
}
