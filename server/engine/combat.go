package engine

import "kingdoms/engine/assets"

type combatant interface {
	combatID() uint32
	combatPosition() Position
	combatRadius(*assets.Catalog) float32
	isFriendly() bool
	isDead() bool
}

func (w *World) visitEnemies(friendly bool, visit func(combatant) bool) {
	if !friendly {
		for _, character := range w.characters {
			if !character.dead && visit(character) {
				return
			}
		}
	}

	for _, npc := range w.npcs {
		if !npc.dead && npc.friendly != friendly && visit(npc) {
			return
		}
	}
}

func (w *World) nearestEnemy(source combatant, limit float32) combatant {
	var nearest combatant
	nearestDistance := limit

	w.visitEnemies(source.isFriendly(), func(target combatant) bool {
		if target.combatID() == source.combatID() {
			return false
		}
		currentDistance := distance(source.combatPosition(), target.combatPosition())
		if currentDistance < nearestDistance {
			nearest = target
			nearestDistance = currentDistance
		}
		return false
	})

	return nearest
}

func (w *World) damageTarget(target combatant, sourceID uint32, damage float32) []Event {
	switch target := target.(type) {
	case *character:
		return w.damageCharacter(target, damage)
	case *npc:
		return w.damageNPC(target, sourceID, damage)
	default:
		return nil
	}
}
