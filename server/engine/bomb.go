package engine

type bomb struct {
	id       uint32
	typeID   uint8
	sourceID uint32
	friendly bool
	damage   float32
	position Position
	origin   Position
	timer    float32
}

func newBomb(
	id uint32,
	typeID uint8,
	source combatant,
	position Position,
	damage float32,
	timer float32,
) *bomb {
	return &bomb{
		id:       id,
		typeID:   typeID,
		sourceID: source.combatID(),
		friendly: source.isFriendly(),
		damage:   damage,
		position: position,
		origin:   source.combatPosition(),
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
			w.visitEnemies(bomb.friendly, func(target combatant) bool {
				if intersects(
					bomb.position,
					radius,
					target.combatPosition(),
					target.combatRadius(w.catalog),
				) {
					events = append(events, w.damageTarget(target, bomb.sourceID, bomb.damage)...)
				}
				return false
			})
		}
		delete(w.bombs, id)
	}

	return events
}
