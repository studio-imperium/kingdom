package engine

import (
	"math"

	"kingdoms/engine/assets"
)

type projectile struct {
	id       uint32
	typeID   uint8
	ownerID  uint32
	hostile  bool
	damage   float32
	position Position
	origin   Position
	angle    uint16
	hits     map[uint32]struct{}
}

func newProjectile(
	id uint32,
	typeID uint8,
	ownerID uint32,
	position Position,
	angle uint16,
	hostile bool,
	damage float32,
) *projectile {
	return &projectile{
		id:       id,
		typeID:   typeID,
		ownerID:  ownerID,
		hostile:  hostile,
		damage:   damage,
		position: position,
		origin:   position,
		angle:    angle,
		hits:     make(map[uint32]struct{}),
	}
}

func (p *projectile) tick(seconds float32, data assets.Projectile) {
	radians := (float32(p.angle) - 90) * math.Pi / 180
	distance := data.Speed * seconds * 60 / 16
	p.position.X += float32(math.Cos(float64(radians))) * distance
	p.position.Y += float32(math.Sin(float64(radians))) * distance
}

func (p *projectile) state() ProjectileState {
	return ProjectileState{
		ID:    p.id,
		Type:  p.typeID,
		X:     p.position.X,
		Y:     p.position.Y,
		Angle: p.angle,
	}
}

func (w *World) tickProjectiles(seconds float32) []Event {
	events := make([]Event, 0)

	for id, projectile := range w.projectiles {
		data, valid := w.catalog.Projectile(projectile.typeID)
		if !valid {
			delete(w.projectiles, id)
			continue
		}

		projectile.tick(seconds, data)
		if distance(projectile.position, projectile.origin) > data.Range {
			delete(w.projectiles, id)
			continue
		}

		hit := false
		if projectile.hostile {
			for targetID, character := range w.characters {
				if character.dead || projectile.hit(targetID) ||
					!intersects(projectile.position, data.Hitbox, character.position, 0.5) {
					continue
				}
				projectile.hits[targetID] = struct{}{}
				events = append(events, w.damageCharacter(character, projectile.damage)...)
				hit = true
				if !data.Piercing {
					break
				}
			}
		} else {
			for targetID, npc := range w.npcs {
				npcData, _ := w.catalog.NPC(npc.typeID)
				if npc.dead || projectile.hit(targetID) ||
					!intersects(projectile.position, data.Hitbox, npc.position, npcData.Hitbox) {
					continue
				}
				projectile.hits[targetID] = struct{}{}
				events = append(events, w.damageNPC(npc, projectile.ownerID, projectile.damage)...)
				hit = true
				if !data.Piercing {
					break
				}
			}
		}

		if hit && !data.Piercing {
			delete(w.projectiles, id)
		}
	}

	return events
}

func (p *projectile) hit(id uint32) bool {
	_, hit := p.hits[id]
	return hit
}
