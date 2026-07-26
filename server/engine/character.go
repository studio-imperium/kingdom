package engine

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type Character struct {
	id    uint32
	x     float32
	y     float32
	angle uint16
	hand  uint8
	head  uint8
	body  uint8

	maxHealth float32
	health    float32
	regen     float32
	Speed     float32
	Power     float32
	Reload    float32

	inventory      map[uint8]uint8
	send           *chan []byte
	AttackCounter  uint8
	AttackCooldown float32
	Simulation     *Engine
	Dead           bool
}

func (c *Character) GetX() float32      { return c.x }
func (c *Character) GetY() float32      { return c.y }
func (c *Character) GetId() uint32      { return c.id }
func (c *Character) GetHitbox() float32 { return 0.5 }
func (c *Character) Damage(amount float32) {
	c.health -= amount

	if c.health < 1 {
		*c.send <- []byte{11}
	} else if (amount > 0) {
		data := new(bytes.Buffer)
		data.WriteByte(byte(5))
		binary.Write(data, binary.LittleEndian, c.id)
		packet := data.Bytes()

		*c.send <- packet
	}

	c.SetHealth(c.health)
}

func (character *Character) SetHealth(val float32) {
	character.health = val

	data := new(bytes.Buffer)
	data.WriteByte(byte(10))
	binary.Write(data, binary.LittleEndian, uint16(val))
	binary.Write(data, binary.LittleEndian, uint16(character.maxHealth))
	packet := data.Bytes()

	*character.send <- packet
}

func (engine *Engine) GetHand(id uint32) uint8 {
	engine.mu.Lock()
	defer engine.mu.Unlock()
	char := engine.Characters[id]
	return char.hand
}

func (engine *Engine) GetSlot(id uint32, idx uint8) uint8 {
	engine.mu.Lock()
	defer engine.mu.Unlock()
	char := engine.Characters[id]
	return char.inventory[idx]
}

func (engine *Engine) ChangeInventory(id uint32, to uint8, from uint8) {
	engine.mu.Lock()
	defer engine.mu.Unlock()

	if to == from {
		return
	}

	char, ok := engine.Characters[id]
	if !ok {
		return
	}

	const (
		headSlot    uint8 = 24
		bodySlot    uint8 = 25
		defaultHead uint8 = 0
		defaultBody uint8 = 1
	)

	gearInfo := func(slot uint8) (*uint8, string, uint8, bool) {
		switch slot {
		case headSlot:
			return &char.head, "head", defaultHead, true
		case bodySlot:
			return &char.body, "body", defaultBody, true
		default:
			return nil, "", 0, false
		}
	}

	if toGear, toType, toDefault, toIsGear := gearInfo(to); toIsGear {
		fromItem, ok := char.inventory[from]
		if !ok || GetItemData(fromItem).Slot != toType {
			return
		}

		equipped := *toGear
		*toGear = fromItem

		if equipped == toDefault {
			delete(char.inventory, from)
		} else {
			char.inventory[from] = equipped
		}

		char.Apply()
		return
	}

	if fromGear, fromType, fromDefault, fromIsGear := gearInfo(from); fromIsGear {
		equipped := *fromGear
		if equipped == fromDefault {
			return
		}

		if toItem, ok := char.inventory[to]; ok {
			if GetItemData(toItem).Slot != fromType {
				return
			}
			*fromGear = toItem
		} else {
			*fromGear = fromDefault
		}

		char.inventory[to] = equipped
		char.Apply()
		return
	}

	item1, ok := char.inventory[from]
	if !ok {
		return
	}
	if item2, ok := char.inventory[to]; ok {
		char.inventory[from] = item2
	} else {
		delete(char.inventory, from)
	}
	char.inventory[to] = item1
	char.Apply()
}

func (char *Character) RemoveInventory(slot uint8) {
	if slot < 24 {
		delete(char.inventory, slot)
	} else if slot == 24 {
		char.head = 0
	} else {
		char.body = 1
	}
	char.Apply()
}

func (engine *Engine) DropItem(id uint32, slot uint8) {
	engine.mu.Lock()
	defer engine.mu.Unlock()

	char, ok := engine.Characters[id]
	if !ok {
		return
	}

	const (
		headSlot    uint8 = 24
		bodySlot    uint8 = 25
		defaultHead uint8 = 0
		defaultBody uint8 = 1
	)

	var item uint8
	switch slot {
	case headSlot:
		if char.head == defaultHead {
			return
		}
		item = char.head
		char.head = defaultHead
	case bodySlot:
		if char.body == defaultBody {
			return
		}
		item = char.body
		char.body = defaultBody
	default:
		v, has := char.inventory[slot]
		if !has {
			return
		}
		item = v
		delete(char.inventory, slot)
	}

	l := CreateLoot(item, char.x, char.y)
	if char.Simulation != nil {
		char.Simulation.AddLoot(l)
	}
	char.Apply()
}

func (engine *Engine) UseItem(id uint32, onUse string) {
	engine.mu.Lock()
	defer engine.mu.Unlock()

	char, ok := engine.Characters[id]
	if !ok {
		return
	}

	switch onUse {
	case "potion_heal":
		char.health += 5
		char.Tick(0)
		char.RemoveInventory(char.hand)
	default:
		fmt.Println("Invalid item use")
	}
}

func (engine *Engine) SelectSlot(id uint32, new uint8) {
	engine.mu.Lock()
	defer engine.mu.Unlock()
	char := engine.Characters[id]
	char.hand = new
}

func (engine *Engine) PackCharacter(id uint32, packet_type uint8) []byte {
	engine.mu.Lock()
	defer engine.mu.Unlock()
	char := engine.Characters[id]
	return char.PackFull(packet_type)
}

func (character *Character) PackFull(packet_type uint8) []byte {
	data := new(bytes.Buffer)

	data.WriteByte(packet_type)
	binary.Write(data, binary.LittleEndian, character.x)
	binary.Write(data, binary.LittleEndian, character.y)
	binary.Write(data, binary.LittleEndian, character.angle)
	binary.Write(data, binary.LittleEndian, uint16(character.health))
	binary.Write(data, binary.LittleEndian, uint16(character.maxHealth))
	binary.Write(data, binary.LittleEndian, float32(character.Reload))
	binary.Write(data, binary.LittleEndian, float32(character.Speed))
	data.WriteByte(character.hand)
	data.WriteByte(character.head)
	data.WriteByte(character.body)

	data.WriteByte(uint8(len(character.inventory)))
	for slot, id := range character.inventory {
		data.WriteByte(slot)
		data.WriteByte(id)
	}
	return data.Bytes()
}

func (character *Character) Pack() []byte {
	data := new(bytes.Buffer)

	binary.Write(data, binary.LittleEndian, character.x)
	binary.Write(data, binary.LittleEndian, character.y)
	binary.Write(data, binary.LittleEndian, character.angle)
	binary.Write(data, binary.LittleEndian, uint16(character.health))
	data.WriteByte(character.inventory[character.hand])
	data.WriteByte(character.head)
	data.WriteByte(character.body)

	return data.Bytes()
}

var weapon uint8 = 4

func DefaultCharacter(simulation *Engine, send *chan []byte, id uint32) *Character {
	weapon += 1
	weapon %= 4
	return &Character{
		id:     id,
		x:      0,
		y:      0,
		angle:  0,
		hand:   0,
		head:   0,
		body:   1,
		health: 1000,
		inventory: map[uint8]uint8{
			0: 8,
		},
		send:           send,
		AttackCounter:  0,
		AttackCooldown: 0,
		Simulation:     simulation,
		Dead:           false,
	}
}

func (character *Character) Tick(delta_sec float32) {
	character.AttackCooldown -= delta_sec

	if character.AttackCooldown < -1 {
		character.AttackCooldown = -1
	}

	tmp := uint16(character.health)
	character.health += character.regen * delta_sec

	if character.health > character.maxHealth {
		character.health = character.maxHealth
	}
	if tmp != uint16(character.health) {
		character.SetHealth(character.health)
	}
}

func (character *Character) Move(x float32, y float32, angle uint16) {
	character.x = x
	character.y = y
	character.angle = angle
}

func NonZero(a float32) float32 {
	if a != 0 {
		return a
	}
	return 1
}

func (character *Character) Apply() {
	var health float32 = 25
	var regen float32 = 1
	var speed float32 = 1
	var damage float32 = 1
	var reload float32 = 1

	helmet := GetItemData(character.head)
	body := GetItemData(character.body)

	health *= NonZero(helmet.Stats.Health)
	health *= NonZero(body.Stats.Health)

	regen *= NonZero(helmet.Stats.Regen)
	regen *= NonZero(body.Stats.Regen)

	speed *= NonZero(helmet.Stats.Speed)
	speed *= NonZero(body.Stats.Speed)

	damage *= NonZero(helmet.Stats.Damage)
	damage *= NonZero(body.Stats.Damage)

	reload *= NonZero(helmet.Stats.Reload)
	reload *= NonZero(body.Stats.Reload)

	character.maxHealth = health
	character.regen = regen
	character.Speed = speed
	character.Power = damage
	character.Reload = reload

	data := character.PackFull(0)
	*character.send <- data
}

func (character *Character) Attack(x float32, y float32, angle uint16) {
	character.x = x
	character.y = y
	character.angle = angle
}
