package engine

import (
	"bytes"
	"encoding/binary"
	"math/rand/v2"
	"sync"
	"time"
)

var Worlds []*Engine = []*Engine{
	CreateIsland(),
}

type Engine struct {
	Characters  map[uint32]*Character
	Npcs        map[uint32]*Npc
	Projectiles map[uint32]*Projectile
	Bombs       map[uint32]*Bomb
	Loot        map[uint32]*Loot
	Map         *Map
	mu          sync.Mutex
}

func (engine *Engine) Pack(packet_type uint8) []byte {
	engine.mu.Lock()
	defer engine.mu.Unlock()
	data := new(bytes.Buffer)

	data.WriteByte(packet_type)
	binary.Write(data, binary.LittleEndian, uint16(len(engine.Characters)))
	for id, character := range engine.Characters {
		binary.Write(data, binary.LittleEndian, id)
		data.Write(character.Pack())
	}
	binary.Write(data, binary.LittleEndian, uint16(len(engine.Npcs)))
	for id, npc := range engine.Npcs {
		binary.Write(data, binary.LittleEndian, id)
		data.Write(npc.Pack())
	}
	binary.Write(data, binary.LittleEndian, uint16(len(engine.Loot)))
	for id, loot := range engine.Loot {
		binary.Write(data, binary.LittleEndian, id)
		data.Write(loot.Pack())
	}
	return data.Bytes()
}

func CreateIsland() *Engine {
	engine := CreateEngine()
	engine.LoadMap("desertonly")

	return engine
}

func CreateEngine() *Engine {
	return &Engine{
		Characters:  make(map[uint32]*Character),
		Npcs:        make(map[uint32]*Npc),
		Projectiles: make(map[uint32]*Projectile),
		Bombs:       make(map[uint32]*Bomb),
		Loot:        make(map[uint32]*Loot),
	}
}

func (engine *Engine) AddCharacter(id uint32, character *Character) {
	engine.mu.Lock()
	defer engine.mu.Unlock()
	engine.Characters[id] = character
}

func (engine *Engine) RemoveCharacter(id uint32, lock bool) {
	if lock {
		engine.mu.Lock()
		defer engine.mu.Unlock()
	}

	character, exists := engine.Characters[id]
	if !exists {
		return
	}
	character.Dead = true

	for _, npc := range engine.Npcs {
		npc.ExitView(id, character)
	}

	delete(engine.Characters, id)
}

func (engine *Engine) CreateProjectile(which uint8, x float32, y float32, angle uint16, evil bool, damage float32) uint32 {
	engine.mu.Lock()
	defer engine.mu.Unlock()

	id := rand.Uint32()
	projectile := DefaultProjectile(which, x, y, angle, evil, damage)
	engine.Projectiles[id] = projectile

	return id
}

func (engine *Engine) AddProjectile(id uint32, projectile *Projectile) {
	engine.mu.Lock()
	defer engine.mu.Unlock()

	engine.Projectiles[id] = projectile
}

func (engine *Engine) CreateBomb(which uint8, x float32, y float32, origin Object, evil bool, damage float32, timer float32) uint32 {
	engine.mu.Lock()
	defer engine.mu.Unlock()

	id := rand.Uint32()
	bomb := DefaultBomb(which, x, y, origin, evil, damage, timer)
	engine.Bombs[id] = bomb

	return id
}

func (engine *Engine) AddBomb(id uint32, bomb *Bomb) {
	engine.mu.Lock()
	defer engine.mu.Unlock()

	engine.Bombs[id] = bomb
}

func (engine *Engine) RemoveNpc(id uint32) {
	engine.mu.Lock()
	defer engine.mu.Unlock()
	delete(engine.Npcs, id)
}

func (engine *Engine) HasCharacter(id uint32) bool {
	engine.mu.Lock()
	defer engine.mu.Unlock()
	_, exists := engine.Characters[id]
	return exists
}

func (engine *Engine) HasNpc(id uint32) bool {
	engine.mu.Lock()
	defer engine.mu.Unlock()
	_, exists := engine.Npcs[id]
	return exists
}

func (engine *Engine) ForEachCharacter(f func(id uint32, character *Character)) {
	engine.mu.Lock()
	defer engine.mu.Unlock()

	for id, character := range engine.Characters {
		f(id, character)
	}
}

func (engine *Engine) ForEachNpc(f func(id uint32, npc *Npc)) {
	engine.mu.Lock()
	defer engine.mu.Unlock()
	for id, npc := range engine.Npcs {
		f(id, npc)
	}
}

func (engine *Engine) MoveCharacter(character *Character, x float32, y float32, angle uint16) {
	engine.mu.Lock()
	defer engine.mu.Unlock()
	character.Move(x, y, angle)
}

func (engine *Engine) AddNpc(id uint32, npc *Npc) {
	engine.mu.Lock()
	defer engine.mu.Unlock()
	npc.entityID = id
	engine.Npcs[id] = npc
}

func (engine *Engine) SpawnNpc(which uint8, x float32, y float32) (uint32, *Npc) {
	engine.mu.Lock()
	defer engine.mu.Unlock()
	id := rand.Uint32()
	npc := DefaultNpc(which, x, y)
	npc.entityID = id
	npc.instance = engine
	engine.Npcs[id] = npc

	return id, npc
}

func (engine *Engine) Run() {
	for i := 0; i < 0; i++ {
		engine.SpawnNpc(1, 400, 400)
	}
	for i := 0; i < 0; i++ {
		engine.SpawnNpc(0, 750, 750)
	}

	delta := time.Millisecond * 50
	ticker := time.NewTicker(delta)
	for {
		engine.mu.Lock()
		for id, npc := range engine.Npcs {
			if len(npc.nearby) > 0 {
				npc.UpdateTarget()
				npc.Tick(delta)
				npc.Move(delta)

				if npc.Dead {
					delete(engine.Npcs, id)
				}
			}
		}
		for id, projectile := range engine.Projectiles {
			projectile.Tick(delta)
			if projectile.Dead {
				delete(engine.Projectiles, id)
			}
		}
		for id, bomb := range engine.Bombs {
			bomb.Tick(delta)
			if bomb.Dead {
				delete(engine.Bombs, id)
			}
		}
		for id, loot := range engine.Loot {
			loot.Tick(delta)
			if loot.Dead || loot.timer <= 0 {
				delete(engine.Loot, id)
			}
		}
		engine.mu.Unlock()
		<-ticker.C
	}
}
