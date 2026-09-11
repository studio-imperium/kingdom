package engine

import (
	"bytes"
	"encoding/binary"
	"gameserver/engine/simulation"
)

func Hit(projectile *Projectile, target Entity) {
	damage(target, projectile.damage, projectile.owner, projectile.evil)
	if !projectileData[projectile.id].Piercing {
		projectile.Dead = true
	}
}

func Splode(bomb *Bomb, target Entity) {
	damage(target, bomb.damage, bomb.owner, bomb.evil)
}

func damage(target Entity, amount float32, owner uint32, evil bool) {
	switch victim := target.(type) {
	case *Npc:
		if victim.Dead {
			return
		}
		victim.damage[owner] += amount
	case *Character:
		if victim.Dead {
			return
		}
	}
	target.Damage(amount)
	victim, ok := target.(*Character)
	if !ok || !victim.Dead || evil || victim.id == owner {
		return
	}
	if killer := victim.instance.Characters[owner]; killer != nil {
		killer.AwardExp(victim.Exp / 2)
	}
	for _, item := range []uint8{victim.inventory[victim.hand], victim.head, victim.body} {
		if item <= 1 {
			continue
		}
		loot := CreateLoot(item, victim.x, victim.y)
		for id, character := range victim.instance.Characters {
			if !character.Dead && !character.disconnected && Distance(victim, character) <= simulation.RenderDistance {
				loot.eligible[id] = true
			}
		}
		victim.instance.Loot[loot.id] = loot
	}
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
	angle := normalizeAngle(p.Angle + float32(Angle(npc.target, npc)) - 90)
	damage := projectileData[p.ID].Damage

	id := npc.instance.nextID()
	projectile := DefaultProjectile(p.ID, x, y, angle, true, damage)
	npc.instance.Projectiles[id] = projectile

	packet := projectile.Pack()

	return id, packet
}
func (npc *Npc) Bomb(b BombSpawnData) (uint32, []byte) {
	x := npc.target.GetX() + b.X
	y := npc.target.GetY() + b.Y
	damage := bombData[b.ID].Damage
	timer := bombData[b.ID].Airtime

	id := npc.instance.nextID()
	bomb := DefaultBomb(b.ID, x, y, npc, true, damage, timer)
	npc.instance.Bombs[id] = bomb

	packet := bomb.Pack()

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

		for _, summon := range attack.Summons {
			npc.instance.SpawnNpc(summon.ID, npc.x+summon.X, npc.y+summon.Y)
		}

		for _, character := range npc.nearby {
			character.Send(data.Bytes())
		}

		npc.attack += 1
		npc.attackTimer = attack.Reload + attack.Wait
	}
}
