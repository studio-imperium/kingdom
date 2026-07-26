package engine

func (w *World) Attack(
	id uint32,
	x, y, targetX, targetY float32,
	attackAngle uint16,
) []Event {
	if !finite(x, y, targetX, targetY) {
		return nil
	}
	w.mu.Lock()
	defer w.mu.Unlock()

	character, exists := w.characters[id]
	if !exists || character.dead {
		return nil
	}

	itemID := character.inventory[character.hand]
	item, valid := w.catalog.Item(itemID)
	if !valid {
		return nil
	}
	if item.OnUse != "" {
		return w.useItem(character, item.OnUse)
	}
	if len(item.Attacks) == 0 || character.attackCooldown > 0.1 {
		return nil
	}
	if character.attackCooldown < -0.1 {
		character.attackCounter = 0
	}

	attack := item.Attacks[int(character.attackCounter)%len(item.Attacks)]
	reload := attack.Reload / character.reload
	character.attackCooldown = reload
	character.attackCounter = (character.attackCounter + 1) % uint8(len(item.Attacks))

	message := AttackMessage{
		SourceID:  id,
		Animation: attack.Animation,
		Duration:  reload,
	}

	for _, spawn := range attack.Projectiles {
		data, _ := w.catalog.Projectile(spawn.ID)
		projectileID := uniqueID(w.projectiles)
		projectile := newProjectile(
			projectileID,
			spawn.ID,
			character,
			Position{X: x + spawn.X, Y: y + spawn.Y},
			(uint16(spawn.Angle)+attackAngle)%360,
			data.Damage*character.power,
		)
		w.projectiles[projectileID] = projectile
		message.Projectiles = append(message.Projectiles, projectile.state())
	}

	for _, spawn := range attack.Bombs {
		data, _ := w.catalog.Bomb(spawn.ID)
		bombID := uniqueID(w.bombs)
		bomb := newBomb(
			bombID,
			spawn.ID,
			character,
			Position{X: targetX + spawn.X, Y: targetY + spawn.Y},
			data.Damage,
			data.Airtime,
		)
		w.bombs[bombID] = bomb
		message.Bombs = append(message.Bombs, bomb.state())
	}

	return []Event{{
		Recipients: w.nearbyCharacterIDs(character.position, id, true),
		Message:    message,
	}}
}

func (w *World) useItem(character *character, action string) []Event {
	if action != "potion_heal" {
		return nil
	}
	character.health = min(character.health+5, character.maxHealth)
	if !character.consume(w.catalog, character.hand) {
		return nil
	}
	return []Event{
		{
			Recipients: []uint32{character.id},
			Message:    CharacterMessage{Character: character.state()},
		},
		{
			Recipients: []uint32{character.id},
			Message: HealthMessage{
				Health:    character.health,
				MaxHealth: character.maxHealth,
			},
		},
	}
}
