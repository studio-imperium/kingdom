package assets

type ProjectileSpawn struct {
	ID    uint8   `json:"id"`
	Angle float32 `json:"angle"`
	X     float32 `json:"x"`
	Y     float32 `json:"y"`
}

type BombSpawn struct {
	ID uint8   `json:"id"`
	X  float32 `json:"x"`
	Y  float32 `json:"y"`
}

type Summon struct {
	ID uint8   `json:"id"`
	X  float32 `json:"x"`
	Y  float32 `json:"y"`
}

type Attack struct {
	Animation   uint8             `json:"animation"`
	Reload      float32           `json:"reload"`
	Wait        float32           `json:"wait,omitempty"`
	Projectiles []ProjectileSpawn `json:"projectiles,omitempty"`
	Bombs       []BombSpawn       `json:"bombs,omitempty"`
	Summons     []Summon          `json:"summons,omitempty"`
}

type NPCMode struct {
	Duration  float32  `json:"duration"`
	MaxHealth float32  `json:"max_health,omitempty"`
	MinHealth float32  `json:"min_health,omitempty"`
	SingleUse bool     `json:"single_use,omitempty"`
	Priority  bool     `json:"priority,omitempty"`
	Movement  string   `json:"movement"`
	Speed     float32  `json:"speed"`
	Attacks   []Attack `json:"attacks,omitempty"`
}

type NPC struct {
	ID     uint8     `json:"id"`
	Name   string    `json:"display"`
	Health float32   `json:"health"`
	Loot   uint16    `json:"loot"`
	Range  float32   `json:"range"`
	Hitbox float32   `json:"hitbox"`
	Modes  []NPCMode `json:"modes,omitempty"`
}

type Spawn struct {
	Display string   `json:"display"`
	Chance  float32  `json:"chance"`
	NPCs    []Summon `json:"npcs"`
}

type SpawnCollection struct {
	Display string  `json:"display"`
	Spawns  []Spawn `json:"spawns"`
}

type Loot struct {
	Item      uint8   `json:"loot"`
	Chance    float32 `json:"chance"`
	Soulbound bool    `json:"soulbound"`
}

type LootTable struct {
	Display string `json:"display"`
	Entries []Loot `json:"entries"`
}

type Stats struct {
	Health float32 `json:"health,omitempty"`
	Regen  float32 `json:"regen,omitempty"`
	Speed  float32 `json:"speed,omitempty"`
	Damage float32 `json:"damage,omitempty"`
	Reload float32 `json:"reload,omitempty"`
}

type Item struct {
	ID      uint8    `json:"id"`
	Slot    string   `json:"type"`
	Stats   Stats    `json:"stats"`
	OnUse   string   `json:"on_use,omitempty"`
	Attacks []Attack `json:"attacks,omitempty"`
}

type Projectile struct {
	ID       uint8   `json:"id"`
	Speed    float32 `json:"speed"`
	Range    float32 `json:"range"`
	Damage   float32 `json:"damage"`
	Piercing bool    `json:"piercing"`
	Hitbox   float32 `json:"hitbox"`
}

type Bomb struct {
	ID      uint8   `json:"id"`
	Airtime float32 `json:"airtime"`
	Damage  float32 `json:"damage"`
	Radius  uint8   `json:"radius"`
}
