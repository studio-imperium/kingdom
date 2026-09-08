package engine

import (
	"context"
	"gameserver/engine/maps"
	"gameserver/engine/simulation"
	"time"
)

// Run owns all mutable world state. Other packages communicate through Input.
type Engine struct {
	Input        chan Packet
	Characters   map[uint32]*Character
	Npcs         map[uint32]*Npc
	Projectiles  map[uint32]*Projectile
	Bombs        map[uint32]*Bomb
	Loot         map[uint32]*Loot
	Map          *maps.Map
	simulations  map[uint32]*simulation.Simulation
	cell_npcs    map[uint16][]uint32
	active_cells map[uint16]bool
	entity_id    uint32
}

func CreateEngine() *Engine {
	return &Engine{
		Input:      make(chan Packet, 256),
		Characters: make(map[uint32]*Character), Npcs: make(map[uint32]*Npc),
		Projectiles: make(map[uint32]*Projectile), Bombs: make(map[uint32]*Bomb), Loot: make(map[uint32]*Loot),
		simulations: make(map[uint32]*simulation.Simulation),
		cell_npcs:   make(map[uint16][]uint32), active_cells: make(map[uint16]bool),
		entity_id: 1 << 31,
	}
}

func (engine *Engine) nextID() uint32 { engine.entity_id++; return engine.entity_id }

func (engine *Engine) RemoveCharacter(id uint32) {
	character, exists := engine.Characters[id]
	if !exists {
		return
	}
	for _, npc := range engine.Npcs {
		npc.ExitView(id, character)
		delete(npc.damage, id)
	}
	for lootID, loot := range engine.Loot {
		delete(loot.eligible, id)
		if len(loot.eligible) == 0 {
			delete(engine.Loot, lootID)
		}
	}
	delete(engine.Characters, id)
	delete(engine.simulations, id)
	if character.finished != nil {
		character.finished <- CharacterData{Id: character.characterID, Dead: character.Dead, Hand: character.hand, Head: character.head, Body: character.body, Inventory: character.inventory}
	}
	close(character.send)
}

func (engine *Engine) CreateProjectile(which uint8, x, y float32, angle uint16, evil bool, damage float32) uint32 {
	id := engine.nextID()
	engine.Projectiles[id] = DefaultProjectile(which, x, y, angle, evil, damage)
	return id
}

func (engine *Engine) CreateBomb(which uint8, x, y float32, origin Object, evil bool, damage, timer float32) uint32 {
	id := engine.nextID()
	engine.Bombs[id] = DefaultBomb(which, x, y, origin, evil, damage, timer)
	return id
}

func (engine *Engine) SpawnNpc(which uint8, x, y float32) (uint32, *Npc) {
	id := engine.nextID()
	npc := DefaultNpc(which, x, y)
	npc.entityID, npc.instance = id, engine
	engine.Npcs[id] = npc
	return id, npc
}

func (engine *Engine) Run(ctx context.Context) {
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()
	defer func() {
		for id := range engine.Characters {
			engine.RemoveCharacter(id)
		}
	}()
	ticks := 0
	for {
		select {
		case <-ctx.Done():
			return
		case packet := <-engine.Input:
			engine.HandlePacket(packet)
		case <-ticker.C:
			engine.Tick(50 * time.Millisecond)
			ticks++
			if ticks%10 == 0 {
				engine.UpdateCells()
			}
			if ticks%4 == 0 {
				engine.SendWorldState()
			}
		}
	}
}
