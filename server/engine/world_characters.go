package engine

import "fmt"

func (w *World) AddCharacter(id uint32) (CharacterState, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if _, exists := w.characters[id]; exists {
		return CharacterState{}, fmt.Errorf("character %d already exists", id)
	}
	if _, exists := w.npcs[id]; exists {
		return CharacterState{}, fmt.Errorf("character %d already exists", id)
	}

	character := newCharacter(id, w.catalog)
	spawn := w.gameMap.beach()
	character.move(spawn.X, spawn.Y, 0)
	w.characters[id] = character
	return character.state(), nil
}

func (w *World) RemoveCharacter(id uint32) {
	w.mu.Lock()
	defer w.mu.Unlock()
	delete(w.characters, id)
}

func (w *World) MoveCharacter(id uint32, x, y float32, angle uint16) bool {
	if !finite(x, y) {
		return false
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	character, exists := w.characters[id]
	if !exists || character.dead {
		return false
	}
	character.move(x, y, angle)
	return true
}

func (w *World) Teleport(id, targetID uint32) (CharacterState, bool) {
	w.mu.Lock()
	defer w.mu.Unlock()
	character, exists := w.characters[id]
	target, targetExists := w.characters[targetID]
	if !exists || !targetExists || character.dead || target.dead {
		return CharacterState{}, false
	}
	character.move(target.position.X, target.position.Y, 0)
	return character.state(), true
}

func (w *World) SelectSlot(id uint32, slot uint8) (CharacterState, bool) {
	w.mu.Lock()
	defer w.mu.Unlock()
	character, exists := w.characters[id]
	if !exists || !character.selectSlot(slot) {
		return CharacterState{}, false
	}
	return character.state(), true
}

func (w *World) ChangeInventory(id uint32, to, from uint8) (CharacterState, bool) {
	w.mu.Lock()
	defer w.mu.Unlock()
	character, exists := w.characters[id]
	if !exists || !character.changeInventory(w.catalog, to, from) {
		return CharacterState{}, false
	}
	return character.state(), true
}

func (w *World) DropItem(id uint32, slot uint8) (CharacterState, bool) {
	w.mu.Lock()
	defer w.mu.Unlock()
	character, exists := w.characters[id]
	if !exists {
		return CharacterState{}, false
	}
	item, dropped := character.drop(w.catalog, slot)
	if !dropped {
		return CharacterState{}, false
	}
	w.addLoot(item, 0, character.position)
	return character.state(), true
}

func (w *World) CharacterPosition(id uint32) (Position, bool) {
	w.mu.RLock()
	defer w.mu.RUnlock()
	character, exists := w.characters[id]
	if !exists {
		return Position{}, false
	}
	return character.position, true
}
