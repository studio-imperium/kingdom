package engine

import (
	"bytes"
	"encoding/binary"
	"gameserver/engine/simulation"
)

func (engine *Engine) addView(id uint32) { engine.simulations[id] = simulation.CreateSimulation() }

func (engine *Engine) SendWorldState() {
	for id, character := range engine.Characters {
		view := engine.simulations[id]
		view.Reset(character.x, character.y)
		var names []uint32
		for id, other := range engine.Characters {
			if !other.Dead && !other.disconnected && view.Visible(other.x, other.y) {
				view.Characters[id] = other.Pack()
				if id != character.id && !view.KnownCharacters[id] {
					names = append(names, id)
				}
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
		packet := bytes.NewBuffer(view.Pack())
		if len(names) > 0 {
			binary.Write(packet, binary.LittleEndian, uint16(len(names)))
			for _, id := range names {
				name := engine.Characters[id].username
				binary.Write(packet, binary.LittleEndian, id)
				packet.WriteByte(uint8(len(name)))
				packet.WriteString(name)
			}
		}
		select {
		case character.send <- packet.Bytes():
			for _, id := range names {
				view.KnownCharacters[id] = true
			}
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
