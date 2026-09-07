package engine

import (
	"embed"
	"gameserver/engine/maps"
	"math/rand/v2"
)

//go:embed assets/maps/*.map
var mapFiles embed.FS

func CreateIsland() (*Engine, error) {
	engine := CreateEngine()
	f, err := mapFiles.Open("assets/maps/desertonly.map")
	if err != nil {
		return nil, err
	}
	defer f.Close()
	engine.Map, err = maps.Load(f)
	return engine, err
}

func (engine *Engine) UpdateCells() {
	if engine.Map == nil {
		return
	}
	active := make(map[uint16]bool)
	for id, character := range engine.Characters {
		if character.Dead || character.disconnected {
			continue
		}
		nearest := engine.Map.GetNearestCell(character.x, character.y)
		cells := append([]*maps.Cell{nearest}, nearest.Adjacent...)
		for _, cell := range cells {
			active[cell.Idx] = true
			for _, discovered := range append([]*maps.Cell{cell}, cell.Adjacent...) {
				if !engine.simulations[id].Cells[discovered.Idx] {
					engine.simulations[id].Cells[discovered.Idx] = true
					character.Send(discovered.Pack())
				}
			}
		}
	}
	for idx := range active {
		if engine.active_cells[idx] {
			continue
		}
		live := engine.cell_npcs[idx][:0]
		for _, id := range engine.cell_npcs[idx] {
			if npc := engine.Npcs[id]; npc != nil && !npc.Dead {
				live = append(live, id)
			}
		}
		engine.cell_npcs[idx] = live
		if len(live) > 0 {
			continue
		}
		cell := engine.Map.Cells[idx]
		seed, odds := rand.Float32(), float32(0)
		for _, spawn := range biomeSpawns[cell.Biome] {
			odds += spawn.Chance
			if seed > odds {
				continue
			}
			for _, npc := range spawn.Npcs {
				id, _ := engine.SpawnNpc(npc.ID, npc.X+float32(cell.Origin.X), npc.Y+float32(cell.Origin.Y))
				engine.cell_npcs[idx] = append(engine.cell_npcs[idx], id)
			}
			break
		}
	}
	engine.active_cells = active
}
