let map_preview = null
let preview_file = null

function create_map_preview(data, offset, size) {
    const palette = document.createElement("canvas")
    palette.width = tile_data.length
    palette.height = 1
    const palette_context = palette.getContext("2d")
    tile_data.forEach(({ id, x, y, size }, index) => {
        if (index) palette_context.drawImage(textures[id].source.resource, x + 8, y + 8, size, size, index, 0, 1, 1)
    })
    const colors = new Uint32Array(palette_context.getImageData(0, 0, palette.width, 1).data.buffer)
    const canvas = document.createElement("canvas")
    canvas.width = canvas.height = Math.min(size, 1024)
    const context = canvas.getContext("2d")
    const image = context.createImageData(canvas.width, canvas.height)
    const pixels = new Uint32Array(image.data.buffer)
    for (let y = 0; y < canvas.height; y++) {
        for (let x = 0; x < canvas.width; x++) {
            const tile = Math.floor((y + 0.5) * size / canvas.height) * size + Math.floor((x + 0.5) * size / canvas.width)
            pixels[y * canvas.width + x] = colors[data.getUint8(offset + tile * 3)]
        }
    }
    context.putImageData(image, 0, 0)
    const sprite = new PIXI.Sprite(PIXI.Texture.from(canvas))
    sprite.texture.source.scaleMode = "nearest"
    sprite.width = sprite.height = size
    return sprite
}

async function import_map(file) {
    if (!file) return
    try {
        const data = new DataView(await file.arrayBuffer())
        const cells = data.getUint16(0, true)
        let offset = 2
        for (let i = 0; i < cells; i++) offset += 6 + data.getUint8(offset + 5) * 2
        const size = data.getUint16(offset, true)
        offset += 2
        if (!cells || !size || data.byteLength !== offset + size * size * 3) throw new Error("Invalid map size.")
        // Validate before replacing the current terrain.
        for (let i = offset; i < data.byteLength; i += 3) {
            if (!tile_data[data.getUint8(i)] || data.getUint16(i + 1, true) >= cells) throw new Error("Invalid map tile.")
        }
        const preview = file.size > 1024 * 1024 ? create_map_preview(data, offset, size) : null
        reset_map()
        tile_blending = file.size <= 1024 * 1024
        if (preview) {
            map_preview = preview
            preview_file = file
            app.stage.addChild(preview)
            select_tool("move", document.querySelector('[title="Pan"]'))
            document.querySelectorAll('.tools button:not([title="Pan"])').forEach(button => button.disabled = true)
            document.getElementById("tile_selection").hidden = true
            document.getElementById("preview_notice").hidden = false
        } else {
            for (let y = 0; y < size; y++) {
                for (let x = 0; x < size; x++) {
                    const tile = data.getUint8(offset + (y * size + x) * 3)
                    if (tile) add_tile(x, y, tile)
                }
            }
            render_tiles()
        }
        const scale = Math.min(32, app.screen.width / (size + 4), app.screen.height / (size + 4))
        app.stage.scale.set(scale)
        app.stage.position.set((app.screen.width - size * scale) / 2, (app.screen.height - size * scale) / 2)
    } catch (error) {
        alert("Could not import this .map file: " + error.message)
    }
}
