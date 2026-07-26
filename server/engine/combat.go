package engine

import (
	"bytes"
	"encoding/binary"
	"math/rand/v2"
)

func Hit(projectile *Projectile, target Entity) {
	target.Damage(projectile.damage)
	if !projectileData[projectile.id].Piercing {
		projectile.Dead = true
	}
}

func Splode(bomb *Bomb, target Entity) {
	target.Damage(bomb.damage)
}

func hitboxesIntersect(projectile *Projectile, entity Entity, is_npc bool) bool {
	projHitbox := float32(projectileData[projectile.id].Hitbox)
	entityHitbox := entity.GetHitbox()

	dx := projectile.x - entity.GetX()
	dy := projectile.y - entity.GetY()
	radius := projHitbox + entityHitbox
	return (dx*dx + dy*dy) <= radius*radius
}

func withinRange(bomb *Bomb, entity Entity, is_npc bool) bool {
	bombRadius := float32(bombData[bomb.id].Radius)
	entityHitbox := entity.GetHitbox()

	dx := bomb.x - entity.GetX()
	dy := bomb.y - entity.GetY()
	radius := bombRadius + entityHitbox
	return (dx*dx + dy*dy) <= radius*radius
}

// npc attacks
func (npc *Npc) Projectile(p ProjectileSpawnData) (uint32, []byte) {
	x := p.X + npc.x
	y := p.Y + npc.y
	raw := int(p.Angle) + int(Angle(npc.target, npc)) - 90
	angle := uint16(((raw % 360) + 360) % 360)
	damage := projectileData[p.ID].Damage

	id := rand.Uint32()
	projectile := DefaultProjectile(p.ID, x, y, angle, true, damage)
	npc.instance.Projectiles[id] = projectile

	packet := projectile.Pack()

	for _, character := range npc.nearby {
		character.Simulation.AddProjectile(id, projectile)
	}

	return id, packet
}
func (npc *Npc) Bomb(b BombSpawnData) (uint32, []byte) {
	x := npc.target.GetX() + b.X
	y := npc.target.GetY() + b.Y
	damage := bombData[b.ID].Damage
	timer := bombData[b.ID].Airtime

	id := rand.Uint32()
	bomb := DefaultBomb(b.ID, x, y, npc, true, damage, timer)
	npc.instance.Bombs[id] = bomb

	packet := bomb.Pack()

	for _, character := range npc.nearby {
		character.Simulation.AddBomb(id, bomb)
	}

	return id, packet
}

func (npc *Npc) NewAttack() {
	if npc.CanAttack() {
		attack := npc.GetAttack()

		data := new(bytes.Buffer)
		data.WriteByte(3)
		binary.Write(data, binary.LittleEndian, npc.entityID)
		data.WriteByte(attack.Animation)
		binary.Write(data, binary.LittleEndian, uint16(1000*attack.Reload))

		binary.Write(data, binary.LittleEndian, uint16(len(attack.Projectiles)))
		for _, projectile := range attack.Projectiles {
			id, packet := npc.Projectile(projectile)

			binary.Write(data, binary.LittleEndian, id)
			data.Write(packet)
		}
		binary.Write(data, binary.LittleEndian, uint16(len(attack.Bombs)))
		for _, bomb := range attack.Bombs {
			id, packet := npc.Bomb(bomb)

			binary.Write(data, binary.LittleEndian, id)
			data.Write(packet)
		}

		npc.instance.mu.Unlock()
		for _, summon := range attack.Summons {
			npc.instance.SpawnNpc(summon.ID, npc.x+summon.X, npc.y+summon.Y)
		}
		npc.instance.mu.Lock()

		for _, character := range npc.nearby {
			*character.send <- data.Bytes()
		}

		npc.attack += 1
		npc.attackTimer = attack.Reload + attack.Wait
	}
}
