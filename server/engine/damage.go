package engine

func (w *World) damageCharacter(character *character, damage float32) []Event {
	if character.dead {
		return nil
	}

	died := character.damage(damage)
	events := []Event{
		{
			Recipients: []uint32{character.id},
			Message:    DamageMessage{TargetID: character.id},
		},
		{
			Recipients: []uint32{character.id},
			Message: HealthMessage{
				Health:    character.health,
				MaxHealth: character.maxHealth,
			},
		},
	}
	if died {
		events = append(events, Event{
			Recipients: []uint32{character.id},
			Message:    DeathMessage{},
		})
	}
	return events
}

func (w *World) damageNPC(npc *npc, ownerID uint32, damage float32) []Event {
	if npc.dead {
		return nil
	}

	npc.health -= damage
	if ownerID != 0 {
		npc.damage[ownerID] += damage
	}

	events := []Event{{
		Recipients: w.nearbyCharacterIDs(npc.position, 0, false),
		Message:    DamageMessage{TargetID: npc.id},
	}}
	if npc.health > 0 {
		return events
	}

	npc.health = 0
	npc.dead = true
	w.dropNPCLoot(npc)
	return events
}

func intersects(a Position, aRadius float32, b Position, bRadius float32) bool {
	x := a.X - b.X
	y := a.Y - b.Y
	radius := aRadius + bRadius
	return x*x+y*y <= radius*radius
}
