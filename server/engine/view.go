package engine

func (w *World) SnapshotFor(id uint32) (Snapshot, bool) {
	w.mu.RLock()
	defer w.mu.RUnlock()
	viewer, exists := w.characters[id]
	if !exists || viewer.dead {
		return Snapshot{}, false
	}

	snapshot := Snapshot{}
	for _, character := range w.characters {
		if !character.dead && distance(viewer.position, character.position) <= renderDistance {
			snapshot.Characters = append(snapshot.Characters, character.state())
		}
	}
	for _, npc := range w.npcs {
		if !npc.dead && distance(viewer.position, npc.position) <= renderDistance {
			snapshot.NPCs = append(snapshot.NPCs, npc.state())
		}
	}
	for _, loot := range w.loot {
		visible := loot.ownerID == 0 || loot.ownerID == id
		if visible && distance(viewer.position, loot.position) <= renderDistance {
			snapshot.Loot = append(snapshot.Loot, loot.state())
		}
	}
	return snapshot, true
}

func (w *World) CellsFor(id uint32) ([]CellState, bool) {
	w.mu.Lock()
	defer w.mu.Unlock()
	character, exists := w.characters[id]
	if !exists || character.dead {
		return nil, false
	}

	cells := w.gameMap.nearby(character.position)
	states := make([]CellState, 0, len(cells))
	for _, cell := range cells {
		w.activate(cell)
		states = append(states, cell.state())
	}
	return states, true
}

func (w *World) nearbyCharacterIDs(
	position Position,
	excludedID uint32,
	exclude bool,
) []uint32 {
	ids := make([]uint32, 0)
	for id, character := range w.characters {
		if character.dead || exclude && id == excludedID {
			continue
		}
		if distance(position, character.position) <= renderDistance {
			ids = append(ids, id)
		}
	}
	return ids
}
