package engine

import "gameserver/engine/simulation"

func (engine *Engine) addView(id uint32) { engine.simulations[id] = simulation.CreateSimulation() }

func (engine *Engine) SendWorldState() {
	for id, character := range engine.Characters {
		view := engine.simulations[id]
		view.Reset(character.x, character.y)
		for id, other := range engine.Characters {
			if !other.Dead && !other.disconnected && view.Visible(other.x, other.y) {
				view.Characters[id] = other.Pack()
			}
		}
		for id, npc := range engine.Npcs {
			if !npc.Dead && view.Visible(npc.x, npc.y) {
				view.Npcs[id] = npc.Pack()
			}
		}
		for id, loot := range engine.Loot {
			if !loot.Dead && loot.eligible[character.id] && view.Visible(loot.x, loot.y) {
				view.Loot[id] = loot.Pack()
			}
		}
		select {
		case character.send <- view.Pack():
		default:
		}
	}
}

func (character *Character) sendToNearby(payload []byte, includeSelf bool) {
	for id, other := range character.instance.Characters {
		if !includeSelf && id == character.id {
			continue
		}
		if !other.Dead && Distance(character, other) <= simulation.RenderDistance {
			other.Send(payload)
		}
	}
}
