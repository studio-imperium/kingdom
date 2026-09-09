function export_map() {
    const tiles = Object.values(tile_map)
    if (!tiles.length) return alert("Paint some terrain before exporting.")
    let left = Infinity, top = Infinity, right = -Infinity, bottom = -Infinity
    for (const { x, y } of tiles) {
        left = Math.min(left, x); top = Math.min(top, y)
        right = Math.max(right, x); bottom = Math.max(bottom, y)
    }
    const size = Math.max(right - left, bottom - top) + 1
    if (size > 65535) return alert("Maps cannot exceed 65535 tiles per side.")
    const data = new DataView(new ArrayBuffer(10 + size * size * 3))
    // One cell: center origin, biome 0, no neighbors. Unpainted tiles stay NULL (0).
    data.setUint16(0, 1, true)
    data.setUint16(2, Math.floor(size / 2), true)
    data.setUint16(4, Math.floor(size / 2), true)
    data.setUint16(8, size, true)
    for (const { x, y, idx } of tiles) data.setUint8(10 + ((y - top) * size + x - left) * 3, idx)
    const link = document.createElement("a")
    link.href = URL.createObjectURL(new Blob([data.buffer], { type: "application/octet-stream" }))
    link.download = "terrain.map"
    link.click()
    setTimeout(() => URL.revokeObjectURL(link.href), 1000)
}
