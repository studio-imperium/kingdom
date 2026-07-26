package protocol

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

const maxPacketSize = 1024

type Command interface {
	command()
}

type HandshakeCommand struct {
	ID uint32
}

func (HandshakeCommand) command() {}

type MoveCommand struct {
	X     float32
	Y     float32
	Angle uint16
}

func (MoveCommand) command() {}

type AttackCommand struct {
	X       float32
	Y       float32
	TargetX float32
	TargetY float32
	Angle   uint16
}

func (AttackCommand) command() {}

type SelectSlotCommand struct {
	Slot uint8
}

func (SelectSlotCommand) command() {}

type ChangeInventoryCommand struct {
	To   uint8
	From uint8
}

func (ChangeInventoryCommand) command() {}

type ChatCommand struct {
	Message string
}

func (ChatCommand) command() {}

type DropItemCommand struct {
	Slot uint8
}

func (DropItemCommand) command() {}

func Decode(reader io.Reader) (Command, error) {
	data, err := io.ReadAll(io.LimitReader(reader, maxPacketSize+1))
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, fmt.Errorf("empty packet")
	}
	if len(data) > maxPacketSize {
		return nil, fmt.Errorf("packet too large")
	}

	decoder := bytes.NewReader(data)
	rawType, err := decoder.ReadByte()
	if err != nil {
		return nil, err
	}

	var command Command
	switch packetType(rawType) {
	case typeHandshake:
		command, err = read[HandshakeCommand](decoder)
	case typeCharacterPosition:
		command, err = read[MoveCommand](decoder)
	case typeCharacterAttack:
		command, err = read[AttackCommand](decoder)
	case typeSelectSlot:
		command, err = read[SelectSlotCommand](decoder)
	case typeChangeInventory:
		command, err = read[ChangeInventoryCommand](decoder)
	case typeChatMessage:
		var length uint8
		if err = binary.Read(decoder, binary.LittleEndian, &length); err == nil {
			message := make([]byte, length)
			_, err = io.ReadFull(decoder, message)
			command = ChatCommand{Message: string(message)}
		}
	case typeDropItem:
		command, err = read[DropItemCommand](decoder)
	default:
		return nil, fmt.Errorf("unknown packet type %d", rawType)
	}

	if err != nil {
		return nil, err
	}
	if decoder.Len() != 0 {
		return nil, fmt.Errorf("unexpected packet data")
	}
	return command, nil
}

func read[T any](reader io.Reader) (T, error) {
	var value T
	err := binary.Read(reader, binary.LittleEndian, &value)
	return value, err
}
