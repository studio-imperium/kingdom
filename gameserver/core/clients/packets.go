package clients

import (
	"encoding/binary"
	"errors"
	"gameserver/engine"
	"math"
)

func ParsePacket(data []byte) (engine.Packet, error) {
	var packet engine.Packet
	if len(data) == 0 {
		return packet, errors.New("empty packet")
	}
	packet.Type = data[0]
	size := 0
	switch packet.Type {
	case engine.CHARACTER_POSITION:
		size = 11
	case engine.CHARACTER_ATTACK:
		size = 19
	case engine.SELECT_SLOT, engine.DROP_ITEM:
		size = 2
	case engine.CHANGE_INVENTORY:
		size = 3
	case engine.CHAT_MESSAGE:
		if len(data) < 2 {
			return packet, errors.New("truncated chat packet")
		}
		size = 2 + int(data[1])
	default:
		return packet, errors.New("invalid packet type")
	}
	if len(data) != size {
		return packet, errors.New("invalid packet length")
	}
	switch packet.Type {
	case engine.CHARACTER_POSITION, engine.CHARACTER_ATTACK:
		packet.X = math.Float32frombits(binary.LittleEndian.Uint32(data[1:5]))
		packet.Y = math.Float32frombits(binary.LittleEndian.Uint32(data[5:9]))
		packet.Angle = binary.LittleEndian.Uint16(data[size-2:]) % 360
		if packet.Type == engine.CHARACTER_ATTACK {
			packet.TargetX = math.Float32frombits(binary.LittleEndian.Uint32(data[9:13]))
			packet.TargetY = math.Float32frombits(binary.LittleEndian.Uint32(data[13:17]))
		}
		for _, value := range []float32{packet.X, packet.Y, packet.TargetX, packet.TargetY} {
			if math.IsNaN(float64(value)) || math.IsInf(float64(value), 0) {
				return packet, errors.New("invalid position")
			}
		}
	case engine.SELECT_SLOT, engine.DROP_ITEM:
		packet.Slot = data[1]
	case engine.CHANGE_INVENTORY:
		packet.Slot, packet.From = data[1], data[2]
	case engine.CHAT_MESSAGE:
		packet.Message = string(data[2:])
	}
	return packet, nil
}
