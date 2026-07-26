package assets

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"path"
)

type Catalog struct {
	npcs          []NPC
	spawns        []SpawnCollection
	loot          []LootTable
	items         []Item
	projectiles   []Projectile
	bombs         []Bomb
	spawnsByBiome map[uint8][]Spawn
}

//go:embed *.json
var jsonFiles embed.FS

//go:embed maps/*.map
var mapFiles embed.FS

func Load() (*Catalog, error) {
	catalog := &Catalog{}

	if err := loadJSON("npcs.json", &catalog.npcs); err != nil {
		return nil, err
	}
	if err := loadJSON("spawns.json", &catalog.spawns); err != nil {
		return nil, err
	}
	if err := loadJSON("loot.json", &catalog.loot); err != nil {
		return nil, err
	}
	if err := loadJSON("items.json", &catalog.items); err != nil {
		return nil, err
	}
	if err := loadJSON("projectiles.json", &catalog.projectiles); err != nil {
		return nil, err
	}
	if err := loadJSON("bombs.json", &catalog.bombs); err != nil {
		return nil, err
	}
	if err := catalog.validate(); err != nil {
		return nil, err
	}

	catalog.spawnsByBiome = make(map[uint8][]Spawn, 11)
	collections := []int{4, 4, 4, 4, 3, 3, 3, 3, 2, 1, 0}
	for biome, collection := range collections {
		catalog.spawnsByBiome[uint8(biome)] = catalog.spawns[collection].Spawns
	}

	return catalog, nil
}

func Files() fs.FS {
	return jsonFiles
}

func OpenMap(name string) (fs.File, error) {
	if path.Base(name) != name {
		return nil, fmt.Errorf("invalid map name %q", name)
	}
	return mapFiles.Open("maps/" + name + ".map")
}

func (c *Catalog) NPC(id uint8) (NPC, bool) {
	if int(id) >= len(c.npcs) {
		return NPC{}, false
	}
	return c.npcs[id], true
}

func (c *Catalog) Item(id uint8) (Item, bool) {
	if int(id) >= len(c.items) {
		return Item{}, false
	}
	return c.items[id], true
}

func (c *Catalog) Projectile(id uint8) (Projectile, bool) {
	if int(id) >= len(c.projectiles) {
		return Projectile{}, false
	}
	return c.projectiles[id], true
}

func (c *Catalog) Bomb(id uint8) (Bomb, bool) {
	if int(id) >= len(c.bombs) {
		return Bomb{}, false
	}
	return c.bombs[id], true
}

func (c *Catalog) Loot(id uint16) ([]Loot, bool) {
	if int(id) >= len(c.loot) {
		return nil, false
	}
	return c.loot[id].Entries, true
}

func (c *Catalog) Spawns(biome uint8) []Spawn {
	return c.spawnsByBiome[biome]
}

func loadJSON(name string, target any) error {
	data, err := jsonFiles.ReadFile(name)
	if err != nil {
		return fmt.Errorf("read %s: %w", name, err)
	}
	if err := json.Unmarshal(data, target); err != nil {
		return fmt.Errorf("parse %s: %w", name, err)
	}
	return nil
}
