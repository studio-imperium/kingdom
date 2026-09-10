function build_character(hand, head, body) {
  let character = new PIXI.Container()
  character.sortableChildren = true

  if (hand && item_data[hand].hand) {
    const hand_obj = build_object(item_data[hand].hand)
    character.addChild(hand_obj)
    hand_layer.attach(hand_obj)
  }

  const body_obj = body
    ? build_object(item_data[body].equipped)
    : build_object(item_data[1].equipped)
  character.addChild(body_obj)
  body_layer.attach(body_obj)

  const head_obj = head
    ? build_object(item_data[head].equipped)
    : build_object(item_data[0].equipped)
  character.addChild(head_obj)
  head_layer.attach(head_obj)

  return character
}
