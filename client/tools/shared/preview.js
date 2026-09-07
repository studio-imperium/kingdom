const OBJECT_SIZE = 8

let __pixi_loaded__ = false
function loadScript(src) {
  return new Promise((resolve, reject) => {
    const s = document.createElement("script")
    s.src = src
    s.onload = resolve
    s.onerror = () => reject(new Error("failed to load " + src))
    document.head.appendChild(s)
  })
}
async function ensurePixi() {
  if (__pixi_loaded__) return
  await loadScript("/client/scripts/pkgs/pixi.js")
  await loadScript("/client/scripts/pkgs/filters.js")
  __pixi_loaded__ = true
}

let __sheet_texture__ = null
async function loadSheet() {
  if (__sheet_texture__) return __sheet_texture__
  __sheet_texture__ = await PIXI.Assets.load("/client/assets/assets.png")
  __sheet_texture__.source.scaleMode = "nearest"
  return __sheet_texture__
}

function makeSprite(frame, doOutline = true) {
  const { w, h, x, y } = frame
  const base = __sheet_texture__
  const safeW = Math.max(1, w | 0)
  const safeH = Math.max(1, h | 0)
  const sx = Math.max(0, x | 0)
  const sy = Math.max(0, y | 0)
  const tex = new PIXI.Texture({
    source: base.source,
    frame: new PIXI.Rectangle(sx, sy, safeW, safeH),
  })
  const sprite = new PIXI.Sprite(tex)
  sprite.scale.set(OBJECT_SIZE)
  if (doOutline) {
    sprite.filters = [
      new PIXI.filters.OutlineFilter({
        thickness: 3,
        color: "black",
        quality: 0.1,
        alpha: 1,
      }),
    ]
  }
  return sprite
}

function buildObject(obj) {
  if (!obj) return new PIXI.Container()
  if (obj.type === "container") {
    const container = new PIXI.Container()
    container.x = (obj.x || 0) * OBJECT_SIZE
    container.y = (obj.y || 0) * OBJECT_SIZE
    container.angle = obj.angle || 0
    const s = obj.scale !== undefined ? obj.scale : 1
    container.scale.set(s)
    container.label = obj.label || ""
    for (const child of obj.children || []) {
      container.addChild(buildObject(child))
    }
    return container
  }
  const sprite = makeSprite(obj.sprite || { w: 0, h: 0, x: 0, y: 0 }, obj.outline !== false)
  sprite.x = (obj.x || 0) * OBJECT_SIZE
  sprite.y = (obj.y || 0) * OBJECT_SIZE
  sprite.angle = obj.angle || 0
  const s = obj.scale !== undefined ? obj.scale : 1
  sprite.scale.set(OBJECT_SIZE * s)
  sprite.label = obj.label || ""
  return sprite
}

const TILE_SIZE = 8

function makeHitboxCircle(tileRadius, color = 0xff4444, alpha = 0.18) {
  const g = new PIXI.Graphics()
  const r = tileRadius * OBJECT_SIZE * TILE_SIZE
  g.circle(0, 0, r).fill({ color, alpha }).stroke({ color, width: 1, alpha: 0.9 })
  g.circle(0, 0, 1).fill({ color, alpha: 0.9 })
  return g
}

function buildCharacter(handItem, headItem, bodyItem) {
  const character = new PIXI.Container()
  character.sortableChildren = true
  if (handItem && handItem.hand && Object.keys(handItem.hand).length) {
    character.addChild(buildObject(handItem.hand))
  }
  if (bodyItem && bodyItem.equipped && Object.keys(bodyItem.equipped).length) {
    character.addChild(buildObject(bodyItem.equipped))
  }
  if (headItem && headItem.equipped && Object.keys(headItem.equipped).length) {
    character.addChild(buildObject(headItem.equipped))
  }
  return character
}

async function makePreview(hostEl, opts = {}) {
  await ensurePixi()
  await loadSheet()

  const app = new PIXI.Application()
  await app.init({
    background: opts.background || 0x0c0c0c,
    width: hostEl.clientWidth || 600,
    height: hostEl.clientHeight || 480,
    antialias: false,
    resolution: window.devicePixelRatio || 1,
    autoDensity: true,
  })
  app.canvas.style.imageRendering = "pixelated"
  hostEl.appendChild(app.canvas)

  const root = new PIXI.Container()
  app.stage.addChild(root)

  function center() {
    root.x = app.screen.width / 2
    root.y = app.screen.height / 2
  }
  center()

  let zoom = opts.zoom || 1
  function applyZoom() {
    root.scale.set(zoom)
  }
  applyZoom()

  const ro = new ResizeObserver(() => {
    const w = hostEl.clientWidth || 600
    const h = hostEl.clientHeight || 480
    app.renderer.resize(w, h)
    center()
  })
  ro.observe(hostEl)

  return {
    app,
    root,
    setZoom(z) {
      zoom = Math.max(0.25, Math.min(8, z))
      applyZoom()
    },
    getZoom() {
      return zoom
    },
    clear() {
      root.removeChildren()
    },
    add(child) {
      root.addChild(child)
    },
  }
}
