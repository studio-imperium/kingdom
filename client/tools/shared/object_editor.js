function createObjectEditor(treeHost, inspectorHost, opts = {}) {
  const listeners = []
  let root = null
  let selectedPath = []

  function fire() {
    for (const l of listeners) l()
  }

  function onChange(cb) {
    listeners.push(cb)
  }

  function getRoot() {
    return root
  }

  function setRoot(newRoot) {
    root = newRoot
    selectedPath = []
    render()
  }

  function getNodeAt(path) {
    let n = root
    for (const i of path) {
      if (!n) return null
      n = (n.children || [])[i]
    }
    return n
  }

  function selectedNode() {
    return getNodeAt(selectedPath)
  }

  function pathEq(a, b) {
    if (a.length !== b.length) return false
    for (let i = 0; i < a.length; i++) if (a[i] !== b[i]) return false
    return true
  }

  function renderTree() {
    treeHost.innerHTML = ""
    if (!root) return
    treeHost.appendChild(renderNode(root, []))
  }

  function renderNode(node, path) {
    const wrap = document.createElement("div")
    wrap.className = path.length ? "tree-node" : ""

    const row = document.createElement("div")
    row.className = "tree-row"
    if (pathEq(path, selectedPath)) row.classList.add("selected")

    const label = document.createElement("span")
    label.className = "label"
    if (node.type === "container") {
      label.textContent = `▾ ${node.label || "(container)"}`
    } else {
      label.textContent = `◆ ${node.label || "(sprite)"}`
    }
    row.appendChild(label)

    const tag = document.createElement("span")
    tag.className = "tag"
    tag.textContent = node.type === "container" ? "container" : "sprite"
    row.appendChild(tag)

    row.addEventListener("click", (e) => {
      e.stopPropagation()
      selectedPath = path
      render()
    })

    wrap.appendChild(row)

    if (node.type === "container" && Array.isArray(node.children)) {
      for (let i = 0; i < node.children.length; i++) {
        wrap.appendChild(renderNode(node.children[i], path.concat(i)))
      }
    }
    return wrap
  }

  function fieldRow(parent, labelText, input) {
    const label = document.createElement("label")
    label.textContent = labelText
    parent.appendChild(label)
    parent.appendChild(input)
  }

  function numInput(value, onInput, step = 0.5) {
    const inp = document.createElement("input")
    inp.type = "number"
    inp.step = step
    inp.value = value !== undefined ? value : ""
    inp.addEventListener("input", () => {
      const v = inp.value === "" ? undefined : Number(inp.value)
      onInput(v)
    })
    return inp
  }

  function textInput(value, onInput) {
    const inp = document.createElement("input")
    inp.type = "text"
    inp.value = value !== undefined ? value : ""
    inp.addEventListener("input", () => {
      onInput(inp.value)
    })
    return inp
  }

  function renderInspector() {
    inspectorHost.innerHTML = ""
    if (!root) {
      inspectorHost.innerHTML = '<div class="muted">No object loaded</div>'
      return
    }
    const node = selectedNode()
    if (!node) {
      inspectorHost.innerHTML = '<div class="muted">Select a node in the tree.</div>'
      return
    }

    const actions = document.createElement("div")
    actions.className = "row"

    if (node.type === "container") {
      const addContainer = document.createElement("button")
      addContainer.className = "small"
      addContainer.textContent = "+ container"
      addContainer.addEventListener("click", () => {
        node.children = node.children || []
        node.children.push({ type: "container", label: "group", children: [] })
        fire(); render()
      })
      actions.appendChild(addContainer)

      const addSprite = document.createElement("button")
      addSprite.className = "small"
      addSprite.textContent = "+ sprite"
      addSprite.addEventListener("click", () => {
        node.children = node.children || []
        node.children.push({
          label: "new_sprite",
          sprite: { w: 8, h: 8, x: 0, y: 0 },
          x: -4,
          y: -4,
        })
        fire(); render()
      })
      actions.appendChild(addSprite)
    }

    if (selectedPath.length > 0) {
      const del = document.createElement("button")
      del.className = "small danger"
      del.textContent = "delete"
      del.addEventListener("click", () => {
        const parentPath = selectedPath.slice(0, -1)
        const idx = selectedPath[selectedPath.length - 1]
        const parent = getNodeAt(parentPath)
        parent.children.splice(idx, 1)
        selectedPath = parentPath
        fire(); render()
      })
      actions.appendChild(del)
    }

    inspectorHost.appendChild(actions)

    const grid = document.createElement("div")
    grid.className = "field-grid"

    fieldRow(grid, "label", textInput(node.label, (v) => { node.label = v; fire(); renderTree() }))

    if (node.type !== "container" && selectedPath.length > 0) {
      const typeSel = document.createElement("select")
      const optS = document.createElement("option"); optS.value = "sprite"; optS.textContent = "sprite"
      const optC = document.createElement("option"); optC.value = "container"; optC.textContent = "container"
      typeSel.appendChild(optS); typeSel.appendChild(optC)
      typeSel.value = "sprite"
      typeSel.addEventListener("change", () => {
        if (typeSel.value === "container") {
          delete node.sprite
          delete node.outline
          node.type = "container"
          node.children = []
        }
        fire(); render()
      })
      fieldRow(grid, "type", typeSel)
    }

    fieldRow(grid, "x", numInput(node.x, (v) => { node.x = v; fire(); renderTree() }))
    fieldRow(grid, "y", numInput(node.y, (v) => { node.y = v; fire(); renderTree() }))
    fieldRow(grid, "angle", numInput(node.angle, (v) => { node.angle = v; fire() }, 1))
    fieldRow(grid, "scale", numInput(node.scale, (v) => { node.scale = v; fire() }, 0.05))

    if (node.type !== "container") {
      const hr = document.createElement("hr")
      grid.appendChild(document.createElement("div"))
      grid.appendChild(hr)

      const sprite = node.sprite = node.sprite || { w: 0, h: 0, x: 0, y: 0 }
      fieldRow(grid, "sprite.x", numInput(sprite.x, (v) => { sprite.x = v || 0; fire() }, 1))
      fieldRow(grid, "sprite.y", numInput(sprite.y, (v) => { sprite.y = v || 0; fire() }, 1))
      fieldRow(grid, "sprite.w", numInput(sprite.w, (v) => { sprite.w = v || 0; fire() }, 1))
      fieldRow(grid, "sprite.h", numInput(sprite.h, (v) => { sprite.h = v || 0; fire() }, 1))

      const outlineCb = document.createElement("input")
      outlineCb.type = "checkbox"
      outlineCb.checked = node.outline !== false
      outlineCb.addEventListener("change", () => {
        if (outlineCb.checked) delete node.outline
        else node.outline = false
        fire()
      })
      fieldRow(grid, "outline", outlineCb)
    }

    inspectorHost.appendChild(grid)
  }

  function render() {
    renderTree()
    renderInspector()
  }

  return {
    setRoot,
    getRoot,
    onChange,
    render,
  }
}
