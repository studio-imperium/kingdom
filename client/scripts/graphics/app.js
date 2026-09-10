const OBJECT_SIZE = 8
const TILE_SIZE = 8
const app = new PIXI.Application()

async function init() {
  await app.init({
    background: "#1f1f1f",
    resizeTo: window,
    width: window.innerWidth,
    height: window.innerHeight,
    useContextAlpha: false,
    antialias: true,
    autoDensity: true,
    resolution: 1,
  })
  document.body.appendChild(app.canvas)
  configure_mobile()

  app.stage.scale = 48
  app.canvas.style.imageRendering = "pixelated"
  app.canvas.style.imageRendering = "crisp-edges"

  await PIXI.Assets.load("/assets/myriad-pro.ttf")

  await load_textures()
  connect()
}
