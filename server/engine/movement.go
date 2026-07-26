package engine

import "kingdoms/engine/assets"

func (n *npc) move(seconds float32, data assets.NPC, target *character) {
	if !n.hasTarget || target == nil {
		n.movement = "wander"
		n.wander(seconds, data)
		return
	}

	switch n.movement {
	case "chase":
		n.chase(seconds, data, target.position)
	case "run":
		n.run(seconds, data, target.position)
	case "overshoot":
		n.overshoot(seconds, data, target.position)
	case "hover":
		n.hover(seconds, data, target.position)
	case "turret":
		n.looking = true
	default:
		n.wander(seconds, data)
	}
}

func (n *npc) chase(seconds float32, data assets.NPC, target Position) {
	speed := n.speed(data)
	target = nearby(target, distance(n.position, target))
	n.position.X += direction(target.X-n.position.X) * speed * seconds
	n.position.Y += direction(target.Y-n.position.Y) * speed * seconds
}

func (n *npc) run(seconds float32, data assets.NPC, target Position) {
	speed := n.speed(data)
	target = nearby(target, distance(n.position, target))
	n.position.X += direction(n.position.X-target.X) * speed * seconds
	n.position.Y += direction(n.position.Y-target.Y) * speed * seconds
}

func (n *npc) hover(seconds float32, data assets.NPC, target Position) {
	currentDistance := distance(n.position, target)
	minimum := data.Range/2 - 1
	maximum := data.Range/2 + 1

	switch {
	case currentDistance > maximum:
		n.chase(seconds, data, target)
		n.looking = false
	case currentDistance < minimum:
		n.run(seconds, data, target)
		n.looking = false
	default:
		n.looking = true
	}
}

func (n *npc) hovering(data assets.NPC, target Position) bool {
	currentDistance := distance(n.position, target)
	return currentDistance > data.Range/2-1 && currentDistance < data.Range/2+1
}

func (n *npc) overshoot(seconds float32, data assets.NPC, target Position) {
	speed := n.speed(data)
	target = nearby(target, 1)
	n.position.X += direction(target.X-n.position.X) * speed * seconds
	n.position.Y += direction(target.Y-n.position.Y) * speed * seconds
}

func (n *npc) wander(seconds float32, data assets.NPC) {
	if !n.hasDestination || distance(n.position, n.destination) < 1 {
		n.destination = nearby(n.origin, 4)
		n.hasDestination = true
	}

	speed := n.speed(data)
	n.position.X += direction(n.destination.X-n.position.X) * speed * seconds
	n.position.Y += direction(n.destination.Y-n.position.Y) * speed * seconds
}

func (n *npc) speed(data assets.NPC) float32 {
	if int(n.mode) >= len(data.Modes) {
		return 0
	}
	return data.Modes[n.mode].Speed
}
