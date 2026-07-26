package engine

import (
	"time"
)

var render_distance int = 16

func (simulation *Engine) AddLoot(loot *Loot) {
	simulation.Loot[loot.id] = loot
}

func (simulation *Engine) StartSimulation(clientID uint32, instance *Engine, clientCharacter *Character) {
	delta := time.Millisecond * 50
	ticker := time.NewTicker(delta)
	for {
		if clientCharacter.Dead {
			return
		} else {
			clientCharacter.Tick(float32(delta) / float32(time.Second))
		}

		// cull characters
		simulation.ForEachCharacter(func(id uint32, character *Character) {
			if id == clientID {
				return
			}
			if character.Dead || int(Distance(character, clientCharacter)) > render_distance {
				simulation.RemoveCharacter(id, false)
			}
		})
		// cull npcs
		simulation.ForEachNpc(func(id uint32, npc *Npc) {
			if npc.Dead {
				delete(simulation.Npcs, id)
				return
			}
			if int(Distance(npc, clientCharacter)) > render_distance {
				npc.ExitView(clientID, clientCharacter)
				delete(simulation.Npcs, id)
			}
		})

		// add characters
		instance.ForEachCharacter(func(id uint32, character *Character) {
			exists := simulation.HasCharacter(id)
			if !exists && int(Distance(character, clientCharacter)) <= render_distance {
				simulation.AddCharacter(id, character)
			}
		})
		// add npcs
		instance.ForEachNpc(func(id uint32, npc *Npc) {
			exists := simulation.HasNpc(id)
			if !exists && int(Distance(npc, clientCharacter)) <= render_distance {
				npc.EnterView(clientID, clientCharacter)
				simulation.AddNpc(id, npc)
			}
		})

		hits := make([]struct {
			projectile *Projectile
			target     Entity
		}, 0)

		splodes := make([]struct {
			bomb   *Bomb
			target Entity
		}, 0)

		simulation.mu.Lock()

		for id, projectile := range simulation.Projectiles {
			out_of_range := Distance(projectile, projectile.origin) > float64(projectileData[projectile.id].Range)
			if projectile.Dead || out_of_range {
				projectile.Dead = true
				delete(simulation.Projectiles, id)
			} else if Distance(projectile, projectile.origin) > float64(projectileData[projectile.id].Range) {
				projectile.Dead = true
			} else if projectile.evil {
				if _, hit := projectile.hitlist[clientID]; hit || projectile.Dead {
					continue
				}
				if hitboxesIntersect(projectile, clientCharacter, false) {
					projectile.hitlist[clientID] = clientCharacter
					hits = append(hits, struct {
						projectile *Projectile
						target     Entity
					}{projectile: projectile, target: clientCharacter})
				}
			} else {
				for npcID, npc := range simulation.Npcs {
					if _, hit := projectile.hitlist[npcID]; hit || projectile.Dead {
						continue
					}
					if hitboxesIntersect(projectile, npc, true) {
						projectile.hitlist[npcID] = npc
						hits = append(hits, struct {
							projectile *Projectile
							target     Entity
						}{projectile: projectile, target: npc})
					}
				}
			}
		}

		for id, bomb := range simulation.Bombs {
			if bomb.timer <= 0 {
				delete(simulation.Bombs, id)
				bomb.Dead = true
				if bomb.evil {
					if withinRange(bomb, clientCharacter, false) {
						splodes = append(splodes, struct {
							bomb   *Bomb
							target Entity
						}{bomb: bomb, target: clientCharacter})
					}
				} else {
					for _, npc := range simulation.Npcs {
						if withinRange(bomb, npc, true) {
							splodes = append(splodes, struct {
								bomb   *Bomb
								target Entity
							}{bomb: bomb, target: npc})
						}
					}
				}
			}
		}

		for id, loot := range simulation.Loot {
			if loot.Dead {
				delete(simulation.Loot, id)
			} else if Distance(clientCharacter, loot) < 1 && clientCharacter.AddItemOrBust(loot.loot) {
				loot.Dead = true
				clientCharacter.Apply()
				*clientCharacter.send <- loot.Looted()
			}
		}

		for _, hit := range hits {
			Hit(hit.projectile, hit.target)
		}
		for _, splode := range splodes {
			Splode(splode.bomb, splode.target)
		}

		simulation.mu.Unlock()
		<-ticker.C
	}
}
