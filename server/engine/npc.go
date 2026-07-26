package engine

import (
	"math/rand/v2"

	"kingdoms/engine/assets"
)

type npc struct {
	id             uint32
	typeID         uint8
	position       Position
	origin         Position
	health         float32
	friendly       bool
	body           assets.NPCBody
	target         combatant
	looking        bool
	destination    Position
	hasDestination bool
	movement       string
	mode           uint8
	usedModes      []bool
	modeTimer      float32
	attack         uint8
	attackTimer    float32
	damage         map[uint32]float32
	dead           bool
}

func newNPC(id uint32, spawn NPCSpawn, data assets.NPC) *npc {
	body := data.Body
	if spawn.BodyOverride != nil {
		body = spawn.BodyOverride
	}

	return &npc{
		id:        id,
		typeID:    spawn.Type,
		position:  spawn.Position,
		origin:    spawn.Position,
		health:    data.Health,
		friendly:  data.Friendly || spawn.Friendly,
		body:      body,
		movement:  "wander",
		usedModes: make([]bool, len(data.Modes)),
		damage:    make(map[uint32]float32),
	}
}

func (n *npc) state() NPCState {
	state := NPCState{
		ID:       n.id,
		Type:     n.typeID,
		X:        n.position.X,
		Y:        n.position.Y,
		Health:   n.health,
		Targeted: n.looking && n.target != nil,
		Friendly: n.friendly,
		Body:     n.body,
	}
	if n.target != nil {
		state.TargetID = n.target.combatID()
	}
	return state
}

func (n *npc) combatID() uint32 {
	return n.id
}

func (n *npc) combatPosition() Position {
	return n.position
}

func (n *npc) combatRadius(catalog *assets.Catalog) float32 {
	data, _ := catalog.NPC(n.typeID)
	return data.Hitbox
}

func (n *npc) isFriendly() bool {
	return n.friendly
}

func (n *npc) isDead() bool {
	return n.dead
}

func (n *npc) tick(seconds float32, data assets.NPC, target combatant, catalog *assets.Catalog) bool {
	n.modeTimer -= seconds
	n.attackTimer -= seconds

	if target == nil || target.isDead() {
		n.target = nil
		n.looking = false
		n.modeTimer = 0
		return false
	}

	if !n.validMode(data, n.mode) || n.modeTimer <= 0 {
		if int(n.mode) < len(n.usedModes) {
			n.usedModes[n.mode] = true
		}
		n.newMode(data)
	}

	targetPosition := target.combatPosition()
	if n.movement == "hover" && !n.hovering(data, targetPosition) {
		return false
	}

	return n.attackTimer < 0 && n.canAttack(data, targetPosition, catalog)
}

func (n *npc) validMode(data assets.NPC, index uint8) bool {
	if int(index) >= len(data.Modes) {
		return false
	}
	mode := data.Modes[index]
	if mode.SingleUse && n.usedModes[index] {
		return false
	}
	if mode.MaxHealth < n.health {
		return false
	}
	return mode.MinHealth < n.health
}

func (n *npc) newMode(data assets.NPC) {
	n.looking = false
	modes := make([]uint8, 0, len(data.Modes))

	for index, mode := range data.Modes {
		if !n.validMode(data, uint8(index)) {
			continue
		}
		modes = append(modes, uint8(index))
		if mode.Priority {
			modes = modes[len(modes)-1:]
			break
		}
	}

	if len(modes) == 0 {
		return
	}

	n.mode = modes[rand.IntN(len(modes))]
	mode := data.Modes[n.mode]
	n.modeTimer = mode.Duration
	n.movement = mode.Movement
	n.attackTimer = 0
}

func (n *npc) currentAttack(data assets.NPC) (assets.Attack, bool) {
	if int(n.mode) >= len(data.Modes) {
		return assets.Attack{}, false
	}
	attacks := data.Modes[n.mode].Attacks
	if len(attacks) == 0 {
		return assets.Attack{}, false
	}
	return attacks[int(n.attack)%len(attacks)], true
}

func (n *npc) canAttack(data assets.NPC, target Position, catalog *assets.Catalog) bool {
	attack, valid := n.currentAttack(data)
	if !valid {
		return false
	}

	var attackRange float32
	for _, spawn := range attack.Projectiles {
		projectile, _ := catalog.Projectile(spawn.ID)
		attackRange = max(attackRange, projectile.Range)
	}
	if len(attack.Bombs) > 0 {
		attackRange = max(attackRange, 32)
	}
	if len(attack.Summons) > 0 {
		attackRange = max(attackRange, data.Range)
	}
	return distance(n.position, target) < attackRange
}
