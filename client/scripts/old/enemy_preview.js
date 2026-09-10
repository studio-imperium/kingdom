async function init_enemy_preview() {
  const npc = npc_data[[9, 8, 18][Math.floor(Math.random() * 3)]] // Sand Golem, Goblin Warlord, Dune Wraith.
  const canvas = document.getElementById("enemy_preview")
  canvas.setAttribute("aria-label", `${npc.display} preview`)
  const preview = new PIXI.Application()
  await preview.init({ canvas, width: 256, height: 256, backgroundAlpha: 0, resolution: 1, antialias: true })
  const enemy = new PIXI.Container(), texture_cache = {}
  for (const layer of ["misc", "hand", "body", "head"]) {
    for (const part of npc.body.filter(part => part.label === layer)) enemy.addChild(build_object(part, preview.renderer, texture_cache))
  }
  enemy.position.set(128, 128)
  enemy.scale.set(40 / 64)
  preview.stage.addChild(enemy)
  const measure = document.createElement("span")
  measure.style.cssText = "position:fixed;visibility:hidden;white-space:pre;pointer-events:none"
  measure.setAttribute("aria-hidden", "true")
  document.body.appendChild(measure)
  let pointer = { x: innerWidth / 2, y: innerHeight / 2 }
  let target_angle = enemy.angle
  function look() {
    let { x, y } = pointer
    const input = document.activeElement
    if (input instanceof HTMLInputElement) {
      const style = getComputedStyle(input), rect = input.getBoundingClientRect()
      const caret = (input.selectionDirection === "backward" ? input.selectionStart : input.selectionEnd) ?? input.value.length
      measure.style.font = style.font
      measure.style.letterSpacing = style.letterSpacing
      measure.textContent = input.type === "password" ? "•".repeat(caret) : input.value.slice(0, caret)
      const start = rect.left + input.clientLeft + parseFloat(style.paddingLeft)
      const end = rect.left + input.clientLeft + input.clientWidth - parseFloat(style.paddingRight)
      x = Math.max(start, Math.min(end, start + measure.getBoundingClientRect().width - input.scrollLeft))
      y = rect.top + rect.height / 2
    }
    const { left, top, width, height } = canvas.getBoundingClientRect()
    target_angle = Math.atan2(y - top - height / 2, x - left - width / 2) * 180 / Math.PI + 90
    if (!(input instanceof HTMLInputElement)) enemy.angle = target_angle
  }
  preview.ticker.add(({ deltaMS }) => {
    if (!(document.activeElement instanceof HTMLInputElement)) return
    const difference = ((target_angle - enemy.angle) % 360 + 540) % 360 - 180
    enemy.angle += difference * (1 - Math.exp(-deltaMS / 90))
  })
  document.addEventListener("pointermove", ({ clientX, clientY }) => {
    pointer = { x: clientX, y: clientY }
    look()
  })
  const update = () => requestAnimationFrame(look)
  for (const event of ["focusin", "focusout", "input", "selectionchange"]) document.addEventListener(event, update)
  document.addEventListener("scroll", update, true)
  window.addEventListener("resize", update)
}
