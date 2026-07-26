package engine

type bomb struct {
	id       uint32
	typeID   uint8
	ownerID  uint32
	hostile  bool
	damage   float32
	position Position
	origin   Position
	timer    float32
}

func newBomb(
	id uint32,
	typeID uint8,
	ownerID uint32,
	position Position,
	origin Position,
	hostile bool,
	damage float32,
	timer float32,
) *bomb {
	return &bomb{
		id:       id,
		typeID:   typeID,
		ownerID:  ownerID,
		hostile:  hostile,
		damage:   damage,
		position: position,
		origin:   origin,
		timer:    timer,
	}
}

func (b *bomb) state() BombState {
	return BombState{
		ID:      b.id,
		Type:    b.typeID,
		X:       b.position.X,
		Y:       b.position.Y,
		OriginX: b.origin.X,
		OriginY: b.origin.Y,
	}
}

func (w *World) tickBombs(seconds float32) []Event {
	events := make([]Event, 0)

	for id, bomb := range w.bombs {
		bomb.timer -= seconds
		if bomb.timer > 0 {
			continue
		}

		data, valid := w.catalog.Bomb(bomb.typeID)
		if valid {
			radius := float32(data.Radius)
			if bomb.hostile {
				for _, character := range w.characters {
					if !character.dead &&
						intersects(bomb.position, radius, character.position, 0.5) {
						events = append(events, w.damageCharacter(character, bomb.damage)...)
					}
				}
			} else {
				for _, npc := range w.npcs {
					npcData, _ := w.catalog.NPC(npc.typeID)
					if !npc.dead &&
						intersects(bomb.position, radius, npc.position, npcData.Hitbox) {
						events = append(events, w.damageNPC(npc, bomb.ownerID, bomb.damage)...)
					}
				}
			}
		}
		delete(w.bombs, id)
	}

	return events
}
