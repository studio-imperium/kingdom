package engine

import (
	"fmt"
	"sync"
	"time"

	"kingdoms/engine/assets"
)

const renderDistance float32 = 16

type World struct {
	mu          sync.RWMutex
	catalog     *assets.Catalog
	gameMap     *gameMap
	characters  map[uint32]*character
	npcs        map[uint32]*npc
	projectiles map[uint32]*projectile
	bombs       map[uint32]*bomb
	loot        map[uint32]*loot
}

func NewWorld(catalog *assets.Catalog, mapName string) (*World, error) {
	file, err := assets.OpenMap(mapName)
	if err != nil {
		return nil, fmt.Errorf("open map: %w", err)
	}
	defer file.Close()

	gameMap, err := loadMap(file)
	if err != nil {
		return nil, err
	}

	return &World{
		catalog:     catalog,
		gameMap:     gameMap,
		characters:  make(map[uint32]*character),
		npcs:        make(map[uint32]*npc),
		projectiles: make(map[uint32]*projectile),
		bombs:       make(map[uint32]*bomb),
		loot:        make(map[uint32]*loot),
	}, nil
}

func (w *World) Tick(delta time.Duration) []Event {
	w.mu.Lock()
	defer w.mu.Unlock()

	seconds := float32(delta) / float32(time.Second)
	events := make([]Event, 0)

	for id, character := range w.characters {
		if !character.dead && character.tick(seconds) {
			events = append(events, Event{
				Recipients: []uint32{id},
				Message: HealthMessage{
					Health:    character.health,
					MaxHealth: character.maxHealth,
				},
			})
		}
	}

	events = append(events, w.tickNPCs(seconds)...)
	events = append(events, w.tickProjectiles(seconds)...)
	events = append(events, w.tickBombs(seconds)...)
	events = append(events, w.tickLoot(seconds)...)

	for id, npc := range w.npcs {
		if npc.dead {
			delete(w.npcs, id)
		}
	}
	return events
}
