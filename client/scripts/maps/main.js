const TILE_SIZE = 8
const app = new PIXI.Application()
const addr = sessionStorage.getItem("kingdom.server") || "server.kingdomcrushers.io"
const prefixs = ["wss", "https"]
let elapsed = 0

async function init() {
    await app.init({ background: "#1f1f1f", resizeTo: window, resolution: 1, antialias: true })
    document.body.appendChild(app.canvas)
    app.canvas.style.display = "block"
    app.canvas.style.imageRendering = "pixelated"
    app.stage.scale.set(32)
    await load_tiles()
    const sheet = new PIXI.Spritesheet(await PIXI.Assets.load("/assets/assets.png"), spritesheet_data)
    await sheet.parse()
    sheet.textureSource.source.scaleMode = "nearest"
    textures = sheet.textures
    tile_data.forEach(({ id, x, y }, index) => {
        if (id === "NULL") return
        const button = document.createElement("button")
        button.className = "icon" + (index === selected_tile ? " selected" : "")
        button.title = id.replaceAll("_", " ")
        button.style.backgroundPosition = `${-x - 8}px ${-y - 8}px`
        button.onclick = () => select(index, button)
        document.getElementById("tile_selection").appendChild(button)
    })
    init_draw()
    document.getElementById("import_button").disabled = false
    app.ticker.add(({ deltaMS }) => { elapsed += deltaMS / 1000; tile_animations() })
}
