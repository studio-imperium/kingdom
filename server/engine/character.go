package engine

import "kingdoms/engine/assets"

const (
	inventorySlots = 24
	headSlot       = 24
	bodySlot       = 25
	defaultHead    = 0
	defaultBody    = 1
)

type character struct {
	id             uint32
	position       Position
	angle          uint16
	hand           uint8
	head           uint8
	body           uint8
	maxHealth      float32
	health         float32
	regen          float32
	speed          float32
	power          float32
	reload         float32
	inventory      map[uint8]uint8
	attackCounter  uint8
	attackCooldown float32
	dead           bool
}

func newCharacter(id uint32, catalog *assets.Catalog) *character {
	character := &character{
		id:        id,
		head:      defaultHead,
		body:      defaultBody,
		inventory: map[uint8]uint8{0: 8},
	}
	character.apply(catalog)
	character.health = character.maxHealth
	return character
}

func (c *character) state() CharacterState {
	inventory := make(map[uint8]uint8, len(c.inventory))
	for slot, item := range c.inventory {
		inventory[slot] = item
	}

	return CharacterState{
		ID:        c.id,
		X:         c.position.X,
		Y:         c.position.Y,
		Angle:     c.angle,
		Health:    c.health,
		MaxHealth: c.maxHealth,
		Reload:    c.reload,
		Speed:     c.speed,
		Hand:      c.hand,
		Head:      c.head,
		Body:      c.body,
		Inventory: inventory,
	}
}

func (c *character) move(x, y float32, angle uint16) {
	c.position = Position{X: x, Y: y}
	c.angle = angle
}

func (c *character) tick(seconds float32) bool {
	previousHealth := uint16(c.health)
	c.attackCooldown = max(c.attackCooldown-seconds, -1)
	c.health = min(c.health+c.regen*seconds, c.maxHealth)
	return previousHealth != uint16(c.health)
}

func (c *character) damage(amount float32) bool {
	c.health = max(c.health-amount, 0)
	if c.health == 0 {
		c.dead = true
	}
	return c.dead
}

func (c *character) apply(catalog *assets.Catalog) {
	helmet, _ := catalog.Item(c.head)
	body, _ := catalog.Item(c.body)

	c.maxHealth = 25 *
		multiplier(helmet.Stats.Health) *
		multiplier(body.Stats.Health)
	c.regen = multiplier(helmet.Stats.Regen) *
		multiplier(body.Stats.Regen)
	c.speed = multiplier(helmet.Stats.Speed) *
		multiplier(body.Stats.Speed)
	c.power = multiplier(helmet.Stats.Damage) *
		multiplier(body.Stats.Damage)
	c.reload = multiplier(helmet.Stats.Reload) *
		multiplier(body.Stats.Reload)
	c.health = min(c.health, c.maxHealth)
}

func (c *character) addItem(item uint8) bool {
	for slot := uint8(0); slot < inventorySlots; slot++ {
		if _, occupied := c.inventory[slot]; !occupied {
			c.inventory[slot] = item
			return true
		}
	}
	return false
}
