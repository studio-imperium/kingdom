(() => {
  const screen = document.getElementById("connecting_screen")
  const canvas = document.createElement("canvas"), ctx = canvas.getContext("2d")
  canvas.setAttribute("aria-hidden", "true")
  canvas.style.cssText = "position:absolute;inset:0;width:100%;height:100%;z-index:-1;pointer-events:none;image-rendering:pixelated"
  screen.prepend(canvas)
  const image = new Image()
  image.onload = () => {
    const tile = document.createElement("canvas")
    tile.width = tile.height = 48
    const tile_ctx = tile.getContext("2d")
    tile_ctx.imageSmoothingEnabled = false
    // The game's 8×8 water sprite, at the same 48px tile scale as gameplay.
    tile_ctx.drawImage(image, 8, 8, 8, 8, 0, 0, 48, 48)
    const pattern = ctx.createPattern(tile, "repeat")
    function animate(time) {
      if (!canvas.isConnected) return
      if (canvas.width !== screen.clientWidth || canvas.height !== screen.clientHeight) {
        canvas.width = screen.clientWidth
        canvas.height = screen.clientHeight
      }
      pattern.setTransform(new DOMMatrix().translate(-Math.sin(time / 1000 * 0.8) * 24, -Math.cos(time / 1000 * 0.6) * 24))
      ctx.fillStyle = pattern
      ctx.fillRect(0, 0, canvas.width, canvas.height)
      requestAnimationFrame(animate)
    }
    requestAnimationFrame(animate)
  }
  image.src = "/assets/assets.png"
})()
