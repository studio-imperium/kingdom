package protocol

import (
	"bytes"
	"encoding/binary"
	"math"

	"kingdoms/engine"
)

type packetType uint8

const (
	typeHandshake packetType = iota
	typeCharacterPosition
	typeCharacterAttack
	typeReceiveAttack
	typeWorldState
	typeDamaged
	typeTiles
	typeSelectSlot
	typeChatMessage
	typeChangeInventory
	typeSetHealth
	typeCharacterDead
	typeLootLooted
	typeDropItem
)

func EncodeMessage(message engine.Message) []byte {
	switch message := message.(type) {
	case engine.AttackMessage:
		return encodeAttack(message)
	case engine.DamageMessage:
		writer := newPacket(typeDamaged)
		writer.uint32(message.TargetID)
		return writer.Bytes()
	case engine.HealthMessage:
		writer := newPacket(typeSetHealth)
		writer.uint16(health(message.Health))
		writer.uint16(health(message.MaxHealth))
		return writer.Bytes()
	case engine.DeathMessage:
		return []byte{byte(typeCharacterDead)}
	case engine.CharacterMessage:
		return EncodeCharacter(message.Character)
	case engine.LootedMessage:
		writer := newPacket(typeLootLooted)
		writer.uint32(message.LootID)
		return writer.Bytes()
	default:
		return nil
	}
}

func EncodeCharacter(character engine.CharacterState) []byte {
	writer := newPacket(typeHandshake)
	writer.float32(character.X)
	writer.float32(character.Y)
	writer.uint16(character.Angle)
	writer.uint16(health(character.Health))
	writer.uint16(health(character.MaxHealth))
	writer.float32(character.Reload)
	writer.float32(character.Speed)
	writer.uint8(character.Hand)
	writer.uint8(character.Head)
	writer.uint8(character.Body)
	writer.uint8(uint8(len(character.Inventory)))
	for slot, item := range character.Inventory {
		writer.uint8(slot)
		writer.uint8(item)
	}
	return writer.Bytes()
}

func EncodeSnapshot(snapshot engine.Snapshot) []byte {
	writer := newPacket(typeWorldState)
	writer.uint16(uint16(len(snapshot.Characters)))
	for _, character := range snapshot.Characters {
		writer.uint32(character.ID)
		writer.float32(character.X)
		writer.float32(character.Y)
		writer.uint16(character.Angle)
		writer.uint16(health(character.Health))
		writer.uint8(character.Inventory[character.Hand])
		writer.uint8(character.Head)
		writer.uint8(character.Body)
	}

	writer.uint16(uint16(len(snapshot.NPCs)))
	for _, npc := range snapshot.NPCs {
		writer.uint32(npc.ID)
		writer.uint8(npc.Type)
		writer.float32(npc.X)
		writer.float32(npc.Y)
		writer.float32(npc.Health)
		if npc.Targeted {
			writer.uint8(1)
			writer.uint32(npc.TargetID)
		} else {
			writer.uint8(0)
		}
	}

	writer.uint16(uint16(len(snapshot.Loot)))
	for _, loot := range snapshot.Loot {
		writer.uint32(loot.ID)
		writer.uint8(loot.Item)
		writer.float32(loot.X)
		writer.float32(loot.Y)
	}
	return writer.Bytes()
}

func EncodeCell(cell engine.CellState) []byte {
	writer := newPacket(typeTiles)
	writer.int32(int32(cell.OriginX))
	writer.int32(int32(cell.OriginY))
	writer.uint16(uint16(len(cell.Tiles)))
	for _, tile := range cell.Tiles {
		writer.int32(int32(tile.X))
		writer.int32(int32(tile.Y))
		writer.uint8(tile.Type)
	}
	return writer.Bytes()
}

func EncodeChat(id uint32, sender, message string) []byte {
	senderBytes := truncate(sender)
	messageBytes := truncate(message)
	writer := newPacket(typeChatMessage)
	writer.uint32(id)
	writer.uint8(uint8(len(senderBytes)))
	writer.Write(senderBytes)
	writer.uint8(uint8(len(messageBytes)))
	writer.Write(messageBytes)
	return writer.Bytes()
}

func encodeAttack(message engine.AttackMessage) []byte {
	writer := newPacket(typeReceiveAttack)
	writer.uint32(message.SourceID)
	writer.uint8(message.Animation)
	writer.uint16(uint16(message.Duration * 1000))
	writer.uint16(uint16(len(message.Projectiles)))
	for _, projectile := range message.Projectiles {
		writer.uint32(projectile.ID)
		writer.uint8(projectile.Type)
		writer.float32(projectile.X)
		writer.float32(projectile.Y)
		writer.uint16(projectile.Angle)
	}
	writer.uint16(uint16(len(message.Bombs)))
	for _, bomb := range message.Bombs {
		writer.uint32(bomb.ID)
		writer.uint8(bomb.Type)
		writer.float32(bomb.X)
		writer.float32(bomb.Y)
		writer.float32(bomb.OriginX)
		writer.float32(bomb.OriginY)
	}
	return writer.Bytes()
}

func truncate(value string) []byte {
	data := []byte(value)
	if len(data) > 255 {
		return data[:255]
	}
	return data
}

func health(value float32) uint16 {
	return uint16(min(max(value, 0), math.MaxUint16))
}

type packetWriter struct {
	bytes.Buffer
}

func newPacket(packetType packetType) *packetWriter {
	writer := &packetWriter{}
	writer.uint8(uint8(packetType))
	return writer
}

func (w *packetWriter) uint8(value uint8) {
	w.WriteByte(value)
}

func (w *packetWriter) uint16(value uint16) {
	var data [2]byte
	binary.LittleEndian.PutUint16(data[:], value)
	w.Write(data[:])
}

func (w *packetWriter) uint32(value uint32) {
	var data [4]byte
	binary.LittleEndian.PutUint32(data[:], value)
	w.Write(data[:])
}

func (w *packetWriter) int32(value int32) {
	w.uint32(uint32(value))
}

func (w *packetWriter) float32(value float32) {
	w.uint32(math.Float32bits(value))
}
