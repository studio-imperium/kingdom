let holding = false
let visited = new Set()
let last_mouse = { x: 0, y: 0 }

function init_draw() {
    function paint(e) {
        if (map_preview) return
        const [x, y] = localize(e.offsetX, e.offsetY)
        const key = `${x},${y}`
        if (visited.has(key)) return
        visited.add(key)
        const size = brush_size(selected_tool)
        for (let dx = 0; dx < size; dx++) {
            for (let dy = 0; dy < size; dy++) add_tile(x + dx, y + dy, selected_tile)
        }
        render_tiles()
    }
    app.canvas.addEventListener("mousedown", e => {
        if (e.button !== 0) return
        holding = true
        visited.clear()
        last_mouse = { x: e.clientX, y: e.clientY }
        if (selected_tool !== 'move') paint(e)
    })
    app.canvas.addEventListener("mousemove", e => {
        if (!holding) return
        if (selected_tool === 'move') {
            app.stage.x += e.clientX - last_mouse.x
            app.stage.y += e.clientY - last_mouse.y
            last_mouse = { x: e.clientX, y: e.clientY }
        } else paint(e)
    })
    window.addEventListener("mouseup", () => { holding = false })
    window.addEventListener("blur", () => { holding = false })
    app.canvas.addEventListener("wheel", event => {
        event.preventDefault()
        const center = { x: app.screen.width / 2, y: app.screen.height / 2 }
        const world = app.stage.toLocal(center)
        const scale = map_preview
            ? Math.max(0.001, Math.min(64, app.stage.scale.x * (event.deltaY < 0 ? 1.2 : 1 / 1.2)))
            : Math.max(0.7, Math.min(64, app.stage.scale.x + (event.deltaY < 0 ? 1 : -1)))
        app.stage.scale.set(scale)
        app.stage.position.set(center.x - world.x * scale, center.y - world.y * scale)
    }, { passive: false })
}

function brush_size(tool) {
    return Number(tool.slice(-1))
}

function localize(x, y) {
    const point = app.stage.toLocal({ x, y })
    return [Math.floor(point.x), Math.floor(point.y)]
}
