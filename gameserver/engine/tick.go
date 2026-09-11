package engine

import (
	"gameserver/engine/simulation"
	"time"
)

func (engine *Engine) Tick(delta time.Duration) {
	for _, character := range engine.Characters {
		if !character.Dead && !character.disconnected {
			character.Tick(float32(delta) / float32(time.Second))
		}
	}
	for id, npc := range engine.Npcs {
		if npc.Dead {
			delete(engine.Npcs, id)
			continue
		}
		clear(npc.nearby)
		for id, character := range engine.Characters {
			if !character.Dead && !character.disconnected && Distance(npc, character) <= simulation.RenderDistance {
				npc.nearby[id] = character
			}
		}
		npc.UpdateTarget()
		if len(npc.nearby) > 0 {
			npc.Tick(delta)
			npc.Move(delta)
		}
	}
	for id, projectile := range engine.Projectiles {
		projectile.Tick(delta)
		if Distance(projectile, projectile.origin) > float64(projectileData[projectile.id].Range) {
			projectile.Dead = true
		}
		if !projectile.Dead {
			for targetID, character := range engine.Characters {
				if character.Dead || character.disconnected || (!projectile.evil && targetID == projectile.owner) {
					continue
				}
				if _, hit := projectile.hitlist[targetID]; hit {
					continue
				}
				if hitboxesIntersect(projectile, character, false) {
					projectile.hitlist[targetID] = character
					Hit(projectile, character)
					if projectile.Dead {
						break
					}
				}
			}
			if !projectile.evil && !projectile.Dead {
				for targetID, npc := range engine.Npcs {
					if npc.Dead {
						continue
					}
					if _, hit := projectile.hitlist[targetID]; hit {
						continue
					}
					if hitboxesIntersect(projectile, npc, true) {
						projectile.hitlist[targetID] = npc
						Hit(projectile, npc)
						if projectile.Dead {
							break
						}
					}
				}
			}
		}
		if projectile.Dead {
			delete(engine.Projectiles, id)
		}
	}
	for id, bomb := range engine.Bombs {
		bomb.Tick(delta)
		if bomb.timer > 0 {
			continue
		}
		for _, character := range engine.Characters {
			if !character.Dead && !character.disconnected && (bomb.evil || character.id != bomb.owner) && withinRange(bomb, character, false) {
				Splode(bomb, character)
			}
		}
		if !bomb.evil {
			for _, npc := range engine.Npcs {
				if !npc.Dead && withinRange(bomb, npc, true) {
					Splode(bomb, npc)
				}
			}
		}
		bomb.Dead = true
		delete(engine.Bombs, id)
	}
	for id, loot := range engine.Loot {
		loot.Tick(delta)
		if loot.Dead || loot.timer <= 0 {
			delete(engine.Loot, id)
			continue
		}
		for playerID := range loot.eligible {
			character := engine.Characters[playerID]
			if character == nil || character.Dead || character.disconnected {
				continue
			}
			if loot.pickupDelay > 0 {
				continue
			}
			distance := Distance(character, loot)
			if loot.pickupNeedsExit {
				loot.pickupNeedsExit = distance < 1
				continue
			}
			if distance < 1 && character.AddItemOrBust(loot.loot) {
				loot.Dead = true
				character.Apply()
				for recipientID := range loot.eligible {
					if recipient := engine.Characters[recipientID]; recipient != nil {
						recipient.Send(loot.Looted())
					}
				}
				delete(engine.Loot, id)
				break
			}
		}
	}
	for id, character := range engine.Characters {
		if character.Dead {
			for _, recipient := range engine.Characters {
				recipient.SendMessage(0, "System", character.username+" has died!")
			}
		}
		if character.Dead || character.disconnected {
			engine.RemoveCharacter(id)
		}
	}
}
