function fillSelect(sel, items, opts = {}) {
  const includeNone = opts.includeNone || false
  const noneLabel = opts.noneLabel || "(none)"
  const noneValue = opts.noneValue !== undefined ? opts.noneValue : ""
  const filter = opts.filter || (() => true)
  const labelFor = opts.labelFor || ((it) => it.display || `#${it.id}`)
  const valueFor = opts.valueFor || ((it) => it.id)

  sel.innerHTML = ""
  if (includeNone) {
    const o = document.createElement("option")
    o.value = String(noneValue)
    o.textContent = noneLabel
    sel.appendChild(o)
  }
  for (const it of items) {
    if (!filter(it)) continue
    const o = document.createElement("option")
    o.value = String(valueFor(it))
    o.textContent = `${labelFor(it)}  (id ${valueFor(it)})`
    sel.appendChild(o)
  }
}

function nextId(items) {
  let max = -1
  for (const it of items) {
    const id = typeof it.id === "number" ? it.id : -1
    if (id > max) max = id
  }
  return max + 1
}
