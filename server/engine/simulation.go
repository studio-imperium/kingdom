package engine

import "kingdoms/simulation"

func (w *World) NPCObservation(id uint32) (simulation.Observation, bool) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	npc, exists := w.npcs[id]
	if !exists || npc.dead {
		return simulation.Observation{}, false
	}

	observer := simulation.Observer{
		ID:       npc.id,
		Position: simulation.Vector{X: npc.position.X, Y: npc.position.Y},
		Friendly: npc.friendly,
	}
	return simulation.Observe(w.simulationFrame(), observer), true
}

func (w *World) simulationFrame() simulation.Frame {
	frame := simulation.Frame{
		Entities:    make([]simulation.EntityState, 0, len(w.characters)+len(w.npcs)),
		Projectiles: make([]simulation.ProjectileState, 0, len(w.projectiles)),
		Bombs:       make([]simulation.BombState, 0, len(w.bombs)),
	}

	for _, character := range w.characters {
		if character.dead {
			continue
		}
		frame.Entities = append(frame.Entities, simulation.EntityState{
			ID:       character.id,
			Position: simulation.Vector{X: character.position.X, Y: character.position.Y},
			Friendly: true,
		})
	}

	for _, npc := range w.npcs {
		if npc.dead {
			continue
		}
		frame.Entities = append(frame.Entities, simulation.EntityState{
			ID:       npc.id,
			Position: simulation.Vector{X: npc.position.X, Y: npc.position.Y},
			Friendly: npc.friendly,
		})
	}

	for _, projectile := range w.projectiles {
		direction := projectile.direction()
		frame.Projectiles = append(frame.Projectiles, simulation.ProjectileState{
			ID:       projectile.id,
			Position: simulation.Vector{X: projectile.position.X, Y: projectile.position.Y},
			Direction: simulation.Vector{
				X: direction.X,
				Y: direction.Y,
			},
			Damage:   projectile.damage,
			Friendly: projectile.friendly,
		})
	}

	for _, bomb := range w.bombs {
		frame.Bombs = append(frame.Bombs, simulation.BombState{
			ID:       bomb.id,
			Position: simulation.Vector{X: bomb.position.X, Y: bomb.position.Y},
			Damage:   bomb.damage,
			Time:     bomb.timer,
			Friendly: bomb.friendly,
		})
	}

	return frame
}
