package engine

import (
	"bytes"
	"encoding/binary"
	"math/rand/v2"
	"time"
)

type Npc struct {
	id       uint8
	entityID uint32
	x        float32
	y        float32
	health   float32
	origin   Object
	instance *Engine

	target      Object
	looking     Entity
	movement    string
	mode        uint8
	usedModes   []bool
	modeTimer   float32
	attack      uint8
	attackTimer float32

	nearby map[uint32]*Character
	damage map[uint32]float32
	Dead   bool
}

func (npc Npc) GetX() float32      { return npc.x }
func (npc Npc) GetY() float32      { return npc.y }
func (npc Npc) GetId() uint32      { return npc.entityID }
func (npc Npc) GetHitbox() float32 { return float32(npcData[npc.id].Hitbox) }
func (npc *Npc) Damage(amount float32) {
	npc.health -= amount

	if npc.health <= 0 && !npc.Dead {
		npc.Dead = true
		npc.Death()
	}

	if len(npc.nearby) == 0 {
		return
	}

	data := new(bytes.Buffer)
	data.WriteByte(byte(5))
	binary.Write(data, binary.LittleEndian, npc.entityID)
	packet := data.Bytes()

	for _, character := range npc.nearby {
		*character.send <- packet
	}
}
func (npc *Npc) Death() {
	// we would use enemies loot pool id
	data := GetNpcData(npc.id)
	lootPool := GetLootData(data.Loot)
	SBThreshold := min(float32(200), data.Health/10.0)

	for id, char := range npc.nearby {
		damage, ok := npc.damage[id]
		for _, loot := range lootPool {
			odds := rand.Float32() <= loot.Chance
			if odds && loot.SB && ok && damage >= SBThreshold {
				l := CreateLoot(loot.Loot, npc.x, npc.y)
				char.Simulation.AddLoot(l)
			}
		}
	}
	for _, loot := range lootPool {
		odds := rand.Float32() <= loot.Chance
		if odds && !loot.SB {
			l := CreateLoot(loot.Loot, npc.x, npc.y)
			for _, char := range npc.nearby {
				char.Simulation.AddLoot(l)
			}
		}
	}
}

func (npc *Npc) Look(obj Entity) {
	npc.looking = obj
}

func (npc *Npc) Pack() []byte {
	data := new(bytes.Buffer)

	binary.Write(data, binary.LittleEndian, npc.id)
	binary.Write(data, binary.LittleEndian, npc.x)
	binary.Write(data, binary.LittleEndian, npc.y)
	binary.Write(data, binary.LittleEndian, npc.health)

	if npc.looking != nil {
		data.WriteByte(1)
		binary.Write(data, binary.LittleEndian, npc.looking.GetId())
	} else {
		data.WriteByte(0)
	}
	return data.Bytes()
}

func (npc *Npc) UpdateTarget() {
	min_dist := GetNpcData(npc.id).Range
	found_character := false

	for _, character := range npc.nearby {
		dist := float32(Distance(npc, character))
		if dist < min_dist {
			npc.target = character
			min_dist = dist
			found_character = true
		}
	}

	if !found_character {
		if _, ok := npc.target.(*Character); ok {
			npc.target = nil
			npc.Look(nil)
		}
	}
}

func (npc *Npc) EnterView(id uint32, character *Character) {
	npc.nearby[id] = character
}
func (npc *Npc) ExitView(id uint32, character *Character) {
	delete(npc.nearby, id)
	if target, ok := npc.target.(*Character); ok && target == character {
		npc.target = nil
	}
}

func (npc *Npc) Data() NpcData {
	return GetNpcData(npc.id)
}

func (npc *Npc) ValidMode(idx uint8) bool {
	data := GetNpcData(npc.id)
	mode := data.Modes[idx]

	if mode.SingleUse && npc.usedModes[idx] {
		return false
	}
	if mode.MaxHealth < npc.health {
		return false
	}
	if mode.MinHealth >= npc.health {
		return false
	}
	return true
}

func (npc *Npc) NewMode() {
	npc.Look(nil)
	data := GetNpcData(npc.id)
	pool := make([]uint8, 0)

	for idx, mode := range data.Modes {
		if npc.ValidMode(uint8(idx)) {
			pool = append(pool, uint8(idx))

			if mode.Priority {
				pool = []uint8{uint8(idx)}
				break
			}
		}
	}

	if len(pool) > 0 {
		mode := pool[rand.IntN(len(pool))]
		npc.mode = mode
		npc.modeTimer = data.Modes[mode].Duration
		npc.movement = data.Modes[mode].Movement
		npc.attackTimer = 0
	}
}
func (npc *Npc) Tick(delta time.Duration) {
	deltaMs := float32(delta) / float32(time.Millisecond)
	npc.modeTimer -= deltaMs / 1000.0
	npc.attackTimer -= deltaMs / 1000.0

	if npc.health <= 0 {
		npc.Dead = true
	}

	if npc.InCombat() {
		if !npc.ValidMode(npc.mode) || npc.modeTimer <= 0 {
			npc.usedModes[npc.mode] = true
			npc.NewMode()
		}

		if npc.movement == "hover" && !npc.Hovering() {
			return
		}

		if npc.attackTimer < 0 {
			npc.NewAttack()
		}
	} else {
		npc.modeTimer = 0
	}
}

func (npc *Npc) GetAttack() *AttackData {
	d := npcData[npc.id]
	mode := d.Modes[npc.mode]
	attackLen := len(mode.Attacks)

	if attackLen > 0 {
		return &mode.Attacks[int(npc.attack)%attackLen]
	}
	return nil
}

func Max(a float32, b float32) float32 {
	if a > b {
		return a
	}
	return b
}

func (npc *Npc) CanAttack() bool {
	dist := float32(Distance(npc.target, npc))
	attack := npc.GetAttack()
	var attack_range float32 = 0

	if attack != nil {
		for _, proj := range attack.Projectiles {
			attack_range = Max(attack_range, GetProjectileData(proj.ID).Range)
		}
		if len(attack.Bombs) > 0 {
			attack_range = 32
		}
		if len(attack.Summons) > 0 {
			attack_range = Max(attack_range, GetNpcData(npc.id).Range)
		}

		return (npc.InCombat() &&
			dist < attack_range)
	} else {
		return false
	}
}

func (npc *Npc) InCombat() bool {
	_, ok := npc.target.(*Character)
	return ok
}

func (npc *Npc) Move(delta time.Duration) {
	if !npc.InCombat() {
		npc.movement = "wander"
	}
	switch npc.movement {
	case "wander":
		npc.Wander(delta)
	case "chase":
		npc.Chase(delta)
	case "run":
		npc.Run(delta)
	case "overshoot":
		npc.Overshoot(delta)
	case "hover":
		npc.Hover(delta)
	case "turret":
		npc.Turret(delta)
	default:
		return
	}
}

func DefaultNpc(id uint8, x float32, y float32) *Npc {
	health := npcData[id].Health

	return &Npc{
		x:      x,
		y:      y,
		health: health,
		origin: Point{x, y},
		id:     id,

		target:    nil,
		looking:   nil,
		movement:  "wander",
		mode:      0,
		usedModes: make([]bool, len(GetNpcData(id).Modes)),
		modeTimer: 0,

		attack:      0,
		attackTimer: 0,
		nearby:      make(map[uint32]*Character),
		damage:      make(map[uint32]float32),
		Dead:        false,
	}
}
