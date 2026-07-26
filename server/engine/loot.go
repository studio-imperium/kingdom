package engine

import "math/rand/v2"

type loot struct {
	id       uint32
	item     uint8
	ownerID  uint32
	position Position
	timer    float32
}

func newLoot(id uint32, item uint8, ownerID uint32, position Position) *loot {
	position.X += rand.Float32() + 1
	position.Y += rand.Float32() + 1
	return &loot{
		id:       id,
		item:     item,
		ownerID:  ownerID,
		position: position,
		timer:    50,
	}
}

func (l *loot) state() LootState {
	return LootState{
		ID:   l.id,
		Item: l.item,
		X:    l.position.X,
		Y:    l.position.Y,
	}
}

func (w *World) tickLoot(seconds float32) []Event {
	events := make([]Event, 0)

	for lootID, loot := range w.loot {
		loot.timer -= seconds
		if loot.timer <= 0 {
			delete(w.loot, lootID)
			continue
		}

		for characterID, character := range w.characters {
			if character.dead || loot.ownerID != 0 && loot.ownerID != characterID ||
				distance(character.position, loot.position) >= 1 ||
				!character.addItem(loot.item) {
				continue
			}

			character.apply(w.catalog)
			delete(w.loot, lootID)
			events = append(events,
				Event{
					Recipients: []uint32{characterID},
					Message:    CharacterMessage{Character: character.state()},
				},
				Event{
					Recipients: []uint32{characterID},
					Message:    LootedMessage{LootID: lootID},
				},
			)
			break
		}
	}

	return events
}
