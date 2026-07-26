package engine

import "kingdoms/engine/assets"

type Position struct {
	X float32
	Y float32
}

type NPCBody = assets.NPCBody

type NPCSpawn struct {
	Type         uint8
	Position     Position
	Friendly     bool
	BodyOverride NPCBody
}

type CharacterState struct {
	ID        uint32
	X         float32
	Y         float32
	Angle     uint16
	Health    float32
	MaxHealth float32
	Reload    float32
	Speed     float32
	Hand      uint8
	Head      uint8
	Body      uint8
	Inventory map[uint8]uint8
}

type NPCState struct {
	ID       uint32
	Type     uint8
	X        float32
	Y        float32
	Health   float32
	TargetID uint32
	Targeted bool
	Friendly bool
	Body     NPCBody
}

type ProjectileState struct {
	ID    uint32
	Type  uint8
	X     float32
	Y     float32
	Angle uint16
}

type BombState struct {
	ID      uint32
	Type    uint8
	X       float32
	Y       float32
	OriginX float32
	OriginY float32
}

type LootState struct {
	ID   uint32
	Item uint8
	X    float32
	Y    float32
}

type Tile struct {
	X    uint16
	Y    uint16
	Type uint8
}

type CellState struct {
	ID      uint16
	OriginX uint16
	OriginY uint16
	Tiles   []Tile
}

type Snapshot struct {
	Characters []CharacterState
	NPCs       []NPCState
	Loot       []LootState
}

type Message interface {
	message()
}

type Event struct {
	Recipients []uint32
	Message    Message
}

type AttackMessage struct {
	SourceID    uint32
	Animation   uint8
	Duration    float32
	Projectiles []ProjectileState
	Bombs       []BombState
}

func (AttackMessage) message() {}

type DamageMessage struct {
	TargetID uint32
}

func (DamageMessage) message() {}

type HealthMessage struct {
	Health    float32
	MaxHealth float32
}

func (HealthMessage) message() {}

type DeathMessage struct{}

func (DeathMessage) message() {}

type CharacterMessage struct {
	Character CharacterState
}

func (CharacterMessage) message() {}

type LootedMessage struct {
	LootID uint32
}

func (LootedMessage) message() {}
