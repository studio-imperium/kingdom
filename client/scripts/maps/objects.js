const object_layer = new PIXI.RenderLayer()
app.stage.addChild(object_layer)

let object_data = {}
let object_map = {}

async function init_objects() {
    const object_json = await (await fetch("/assets/objects.json")).json()
    
    for (const data of object_json) {
        textures[data.id] = await load_asset(data.path)
        object_data[data.id] = data
    }

    // const head = await PIXI.Assets.load('/assets/objects/head.png')
    // const body = await PIXI.Assets.load('/assets/objects/body.png')
    // const weapon = await PIXI.Assets.load('/assets/objects/weapon.png')

    // for (let texture of [head,body,weapon]) {
    //     texture.source.scaleMode = "nearest"
    //     texture.source.addressMode = "repeat"
    // }

    // const headsprite = create_sprite(head, -3.5,-2)
    // const bodysprite = create_sprite(body, -5.5,0)
    // const weaponsprite = create_sprite(weapon, 5,-6)
    // // weaponsprite.angle = 23.5
    // weaponsprite.scale = 1
    // const container = new PIXI.Container()

    // container.addChild(weaponsprite)
    // container.addChild(bodysprite)
    // container.addChild(headsprite)
    // app.stage.addChild(container)
    // object_layer.attach(container)
    // object_layer.sortRenderLayerChildren()

    // window.addEventListener("keydown", (e) => {
    //     if (e.key == "q") {
    //         container.angle -= 10
    //     }
    //     if (e.key == "e") {
    //         container.angle += 10
    //     }
    // })
}

function create_sprite(texture, x, y) {
    const sprite = new PIXI.Sprite(texture)
    sprite.position.set(x, y)

    const width = 2
    const color = "black"
    const quality = 0.5
    const alpha = 0.8
    
    const outline = new PIXI.filters.OutlineFilter(
        width,
        color,
        quality,
        alpha,
    )
    const shadow = new PIXI.filters.OutlineFilter(
        width,
        color,
        quality/2,
        alpha/2,
    )
    const shadow2 = new PIXI.filters.OutlineFilter(
        width*2,
        color,
        quality/2,
        alpha/4,
    )
    
    sprite.filters = [
        outline,
        shadow,
        shadow2
    ]
    return sprite
}

function add_object(x, y, object_id, object_key = `${x},${y}`) {
    const data = object_data[object_id]
    const texture = textures[object_id]
    const sprite = create_sprite(texture, x*8, y*8)
    sprite.scale = data.scale

    if (object_map[object_key]) {
        object_map[object_key].sprite.destroy()
        delete( object_map[object_key])
    }
    else {
        object_map[object_key] = {
            "x" : x,
            "y" : y,
            "object" : object_id,
            "sprite" : sprite,
        }

        sprite.zIndex = sprite.scale.x
        app.stage.addChild(sprite)
        object_layer.attach(sprite)
        object_layer.sortRenderLayerChildren()
    }
}
