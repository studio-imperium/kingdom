function populate_map({ cells, size }) {
    for (const { mesh } of Object.values(tile_map)) {
        mesh.destroy()
    }

    let ids = Object.keys(tile_data)

    tile_map = {}
    for (const cell of cells) {
        for (const tile of cell.tiles) {
            add_tile(tile.x, tile.y, ids[tile.value])
        }
    }
    render_tiles()
}