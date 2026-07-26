package engine

import "math/rand/v2"

func (w *World) SpawnNPC(spawn NPCSpawn) (uint32, bool) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.spawnNPC(spawn)
}

func (w *World) SpawnLoot(item uint8, ownerID uint32, position Position) (uint32, bool) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if _, valid := w.catalog.Item(item); !valid {
		return 0, false
	}
	if ownerID != 0 {
		if _, exists := w.characters[ownerID]; !exists {
			return 0, false
		}
	}
	return w.addLoot(item, ownerID, position), true
}

func (w *World) spawnNPC(spawn NPCSpawn) (uint32, bool) {
	data, valid := w.catalog.NPC(spawn.Type)
	if !valid {
		return 0, false
	}
	id := w.entityID()
	w.npcs[id] = newNPC(id, spawn, data)
	return id, true
}

func (w *World) addLoot(item uint8, ownerID uint32, position Position) uint32 {
	id := uniqueID(w.loot)
	w.loot[id] = newLoot(id, item, ownerID, position)
	return id
}

func (w *World) activate(cell *cell) {
	if cell.activated {
		return
	}
	cell.activated = true

	roll := rand.Float32()
	var chance float32
	for _, spawn := range w.catalog.Spawns(cell.biome) {
		chance += spawn.Chance
		if roll > chance {
			continue
		}
		for _, npc := range spawn.NPCs {
			w.spawnNPC(NPCSpawn{
				Type: npc.ID,
				Position: Position{
					X: cell.origin.X + npc.X,
					Y: cell.origin.Y + npc.Y,
				},
			})
		}
		return
	}
}

func uniqueID[T any](values map[uint32]T) uint32 {
	for {
		id := rand.Uint32()
		if id == 0 {
			continue
		}
		if _, exists := values[id]; !exists {
			return id
		}
	}
}

func (w *World) entityID() uint32 {
	for {
		id := rand.Uint32()
		if id == 0 {
			continue
		}
		if _, exists := w.characters[id]; exists {
			continue
		}
		if _, exists := w.npcs[id]; !exists {
			return id
		}
	}
}
