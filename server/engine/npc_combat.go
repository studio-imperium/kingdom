package engine

import (
	"math/rand/v2"

	"kingdoms/engine/assets"
)

func (w *World) tickNPCs(seconds float32) []Event {
	events := make([]Event, 0)

	for _, npc := range w.npcs {
		if npc.dead {
			continue
		}
		data, _ := w.catalog.NPC(npc.typeID)
		npc.updateTarget(w.characters, data)
		target := w.characters[npc.targetID]

		if npc.tick(seconds, data, target, w.catalog) {
			events = append(events, w.npcAttack(npc, data, target))
		}
		npc.move(seconds, data, target)
	}

	return events
}

func (w *World) npcAttack(npc *npc, data assets.NPC, target *character) Event {
	attack, _ := npc.currentAttack(data)
	message := AttackMessage{
		SourceID:  npc.id,
		Animation: attack.Animation,
		Duration:  attack.Reload,
	}

	for _, spawn := range attack.Projectiles {
		projectileData, _ := w.catalog.Projectile(spawn.ID)
		projectileID := uniqueID(w.projectiles)
		rawAngle := int(spawn.Angle) + int(angle(target.position, npc.position)) - 90
		projectileAngle := uint16((rawAngle%360 + 360) % 360)
		projectile := newProjectile(
			projectileID,
			spawn.ID,
			0,
			Position{X: npc.position.X + spawn.X, Y: npc.position.Y + spawn.Y},
			projectileAngle,
			true,
			projectileData.Damage,
		)
		w.projectiles[projectileID] = projectile
		message.Projectiles = append(message.Projectiles, projectile.state())
	}

	for _, spawn := range attack.Bombs {
		bombData, _ := w.catalog.Bomb(spawn.ID)
		bombID := uniqueID(w.bombs)
		bomb := newBomb(
			bombID,
			spawn.ID,
			0,
			Position{X: target.position.X + spawn.X, Y: target.position.Y + spawn.Y},
			npc.position,
			true,
			bombData.Damage,
			bombData.Airtime,
		)
		w.bombs[bombID] = bomb
		message.Bombs = append(message.Bombs, bomb.state())
	}

	for _, summon := range attack.Summons {
		w.spawnNPC(summon.ID, Position{
			X: npc.position.X + summon.X,
			Y: npc.position.Y + summon.Y,
		})
	}

	npc.attack++
	npc.attackTimer = attack.Reload + attack.Wait
	return Event{
		Recipients: w.nearbyCharacterIDs(npc.position, 0, false),
		Message:    message,
	}
}

func (w *World) dropNPCLoot(npc *npc) {
	data, _ := w.catalog.NPC(npc.typeID)
	table, _ := w.catalog.Loot(data.Loot)
	threshold := min(float32(200), data.Health/10)

	for _, entry := range table {
		if entry.Soulbound {
			for characterID, damage := range npc.damage {
				if damage >= threshold && rand.Float32() <= entry.Chance {
					w.addLoot(entry.Item, characterID, npc.position)
				}
			}
		} else if rand.Float32() <= entry.Chance {
			w.addLoot(entry.Item, 0, npc.position)
		}
	}
}
