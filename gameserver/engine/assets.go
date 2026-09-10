package engine

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type ProjectileSpawnData struct {
	ID    uint8   `json:"id"`
	Angle float32 `json:"angle"`
	X     float32 `json:"x"`
	Y     float32 `json:"y"`
}

type BombSpawnData struct {
	ID uint8   `json:"id"`
	X  float32 `json:"x"`
	Y  float32 `json:"y"`
}

type SummonData struct {
	ID uint8   `json:"id"`
	X  float32 `json:"x"`
	Y  float32 `json:"y"`
}

type AttackData struct {
	Animation   uint8                 `json:"animation"`
	Reload      float32               `json:"reload"`
	Wait        float32               `json:"wait,omitempty"`
	Projectiles []ProjectileSpawnData `json:"projectiles,omitempty"`
	Bombs       []BombSpawnData       `json:"bombs,omitempty"`
	Summons     []SummonData          `json:"summons,omitempty"`
}

type NpcModeData struct {
	Duration  float32      `json:"duration"`
	MaxHealth float32      `json:"max_health,omitempty"`
	MinHealth float32      `json:"min_health,omitempty"`
	SingleUse bool         `json:"single_use,omitempty"`
	Priority  bool         `json:"priority,omitempty"`
	Movement  string       `json:"movement"`
	Speed     float32      `json:"speed"`
	Attacks   []AttackData `json:"attacks,omitempty"`
}

type NpcData struct {
	ID     uint8         `json:"id"`
	Name   string        `json:"display"`
	Health float32       `json:"health"`
	Exp    int64         `json:"exp"`
	Loot   uint16        `json:"loot"`
	Range  float32       `json:"range"`
	Hitbox float32       `json:"hitbox"`
	Modes  []NpcModeData `json:"modes,omitempty"`
}

type SpawnData struct {
	Display string       `json:"display"`
	Chance  float32      `json:"chance"`
	Npcs    []SummonData `json:"npcs"`
}

type SpawnCollection struct {
	Display string      `json:"display"`
	Spawns  []SpawnData `json:"spawns"`
}

type LootData struct {
	Loot   uint8   `json:"loot"`
	Chance float32 `json:"chance"`
	SB     bool    `json:"soulbound"`
}

type LootTable struct {
	Display string     `json:"display"`
	Entries []LootData `json:"entries"`
}

type Stats struct {
	Health float32 `json:"health,omitempty"`
	Regen  float32 `json:"regen,omitempty"`
	Speed  float32 `json:"speed,omitempty"`
	Damage float32 `json:"damage,omitempty"`
	Reload float32 `json:"reload,omitempty"`
}

type ItemData struct {
	ID      uint8        `json:"id"`
	Slot    string       `json:"type"`
	Stats   Stats        `json:"stats"`
	OnUse   string       `json:"on_use,omitempty"`
	Attacks []AttackData `json:"attacks,omitempty"`
}

type ProjectileData struct {
	ID       uint8   `json:"id"`
	Speed    float32 `json:"speed"`
	Range    float32 `json:"range"`
	Damage   float32 `json:"damage"`
	Piercing bool    `json:"piercing"`
	Hitbox   float32 `json:"hitbox"`
}

type BombData struct {
	ID      uint8   `json:"id"`
	Airtime float32 `json:"airtime"`
	Damage  float32 `json:"damage"`
	Radius  uint8   `json:"radius"`
}

var npcData []NpcData
var spawnsData []SpawnCollection
var lootData []LootTable
var itemData []ItemData
var projectileData []ProjectileData
var bombData []BombData

var biomeSpawns map[uint8][]SpawnData = map[uint8][]SpawnData{}

func GetNpcData(id uint8) NpcData {
	return npcData[id]
}
func spawnsByCollection(idx int) []SpawnData {
	if idx < 0 || idx >= len(spawnsData) {
		return nil
	}
	return spawnsData[idx].Spawns
}
func GetLootData(id uint16) []LootData {
	return lootData[id].Entries
}
func GetItemData(id uint8) ItemData {
	return itemData[id]
}
func GetProjectileData(id uint8) ProjectileData {
	return projectileData[id]
}
func GetBombData(id uint8) BombData {
	return bombData[id]
}

func InitAssets(baseURL string) error {
	client := &http.Client{Timeout: 5 * time.Second}
	for _, asset := range []struct {
		name string
		data any
	}{
		{"npcs", &npcData}, {"spawns", &spawnsData}, {"loot", &lootData},
		{"items", &itemData}, {"projectiles", &projectileData}, {"bombs", &bombData},
	} {
		response, err := client.Get(baseURL + asset.name + ".json")
		if err != nil {
			return fmt.Errorf("load %s: %w", asset.name, err)
		}
		if response.StatusCode != http.StatusOK {
			response.Body.Close()
			return fmt.Errorf("load %s: HTTP %d", asset.name, response.StatusCode)
		}
		err = json.NewDecoder(response.Body).Decode(asset.data)
		response.Body.Close()
		if err != nil {
			return fmt.Errorf("load %s: %w", asset.name, err)
		}
	}

	// desert
	biomeSpawns[0] = spawnsByCollection(4)
	biomeSpawns[1] = spawnsByCollection(4)
	biomeSpawns[2] = spawnsByCollection(4)

	biomeSpawns[3] = spawnsByCollection(4)
	biomeSpawns[4] = spawnsByCollection(3)

	biomeSpawns[5] = spawnsByCollection(3)
	biomeSpawns[6] = spawnsByCollection(3)
	biomeSpawns[7] = spawnsByCollection(3)

	// forest
	biomeSpawns[8] = spawnsByCollection(2)
	biomeSpawns[9] = spawnsByCollection(1)

	// beach
	biomeSpawns[10] = spawnsByCollection(0)

	fmt.Println("Initialized", len(npcData), "NPCs")
	fmt.Println("Initialized", len(spawnsData), "Spawns")
	fmt.Println("Initialized", len(lootData), "Loot Tables")
	fmt.Println("Initialized", len(itemData), "Items")
	fmt.Println("Initialized", len(projectileData), "Projectiles")
	fmt.Println("Initialized", len(bombData), "Bombs")
	return nil
}
