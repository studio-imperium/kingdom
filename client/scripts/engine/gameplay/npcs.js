const npcs = {}

class Npc {
  constructor(id, x, y, health) {
    this.object = build_npc(id)
    this.object.angle = 0
    this.object.x = x
    this.object.y = y
    this.animator = new Animator(this.object)
    this.interpolator = new Interpolator(this.object)
    this.colorAnimator = new ColorAnimator(this.object)
    this.which = id

    add_object(this.object)
  }

  update(x, y, health) {
    this.object.health = health
    this.interpolator.add_npc_frame(x, y)
  }

  damage() {
    this.colorAnimator.animate(0xffb3b3, 300)
  }

  kill(id) {
    this.colorAnimator.animate(0xffb3b3, 30000)
    this.animator.animate(0, 0.2)
    this.interpolator.frames = []
    this.interpolator.last_frame = Date.now()

    setTimeout(() => {
      this.object.destroy()
      delete npcs[id]
    }, 300)
  }
}

function build_npc(npc_id) {
  let npc = new PIXI.Container()

  for (let bodypart of npc_data[npc_id].body) {
    let obj = build_object(bodypart)

    if (bodypart.label == "hand") {
      hand_layer.attach(obj)
    } else if (bodypart.label == "body") {
      body_layer.attach(obj)
    } else if (bodypart.label == "head") {
      head_layer.attach(obj)
    } else {
      misc_layer.attach(obj)
    }
    npc.addChild(obj)
  }

  return npc
}
