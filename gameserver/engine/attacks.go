package engine

import (
	"bytes"
	"encoding/binary"
)

func (character *Character) characterAttack(x float32, y float32, targetX float32, targetY float32, angle uint16) {
	idx := character.instance.GetHand(character.id)
	item_id := character.instance.GetSlot(character.id, idx)
	item := GetItemData(item_id)

	cooldown := character.AttackCooldown
	data := new(bytes.Buffer)
	data.WriteByte(uint8(ALLY_ATTACK))

	if item.OnUse != "" {
		character.instance.UseItem(character.id, item.OnUse)
		return
	}

	if cooldown < -0.1 {
		character.AttackCounter = 0
	}
	if cooldown <= 0.01 {
		counter := character.AttackCounter

		if len(item.Attacks) > 0 {
			attack := item.Attacks[int(counter)%len(item.Attacks)]
			animation := attack.Animation
			projectiles := attack.Projectiles
			bombs := attack.Bombs

			reload := attack.Reload / character.Reload

			character.AttackCooldown = reload

			binary.Write(data, binary.LittleEndian, character.id)
			data.WriteByte(animation)
			binary.Write(data, binary.LittleEndian, uint16(reload*1000))

			binary.Write(data, binary.LittleEndian, uint16(len(projectiles)))

			for _, projectile := range projectiles {
				baseDamage := GetProjectileData(projectile.ID).Damage

				damage := float32(baseDamage) * character.Power

				id := character.instance.CreateProjectile(projectile.ID, projectile.X+x, projectile.Y+y, normalizeAngle(projectile.Angle+float32(angle)), false, damage)
				proj := character.instance.Projectiles[id]
				packet := proj.Pack()

				binary.Write(data, binary.LittleEndian, id)
				data.Write(packet)

				proj.owner = character.id
			}

			binary.Write(data, binary.LittleEndian, uint16(len(bombs)))

			for _, bomb := range bombs {
				baseDamage := GetBombData(bomb.ID).Damage
				damage := baseDamage
				timer := GetBombData(bomb.ID).Airtime

				id := character.instance.CreateBomb(bomb.ID, targetX, targetY, character, false, damage, timer)
				bomb := character.instance.Bombs[id]
				packet := bomb.Pack()

				binary.Write(data, binary.LittleEndian, id)
				data.Write(packet)

				bomb.owner = character.id
			}

			character.AttackCounter += 1
			character.AttackCounter %= uint8(len(item.Attacks))
			character.sendToNearby(data.Bytes(), false)
		}
	}
}
