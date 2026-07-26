package packets

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"server/engine"
	"time"
)

type PacketType uint8

const (
	HANDSHAKE PacketType = iota
	CHARACTER_POSITION
	CHARACTER_ATTACK
	ALLY_ATTACK
	WORLDSTATE
	DAMAGED
	TILES
	SELECT_SLOT
	CHAT_MESSAGE
	CHANGE_INVENTORY
	SET_HEALTH
	CHARACTER_DEAD
	LOOT_LOOTED
	DROP_ITEM
)

var tokens map[uint32]*engine.Character = make(map[uint32]*engine.Character)
var Clients map[uint32]*Client = make(map[uint32]*Client)

func PropogateWorldState() {
	delta := time.Second / 5
	ticker := time.NewTicker(delta)
	for {
		for _, client := range Clients {
			data := client.simulation.Pack(byte(WORLDSTATE))
			select {
			case client.send <- data:
			default:
				fmt.Println("Dropping world state for blocked client:", client.id)
			}
		}
		<-ticker.C
	}
}

func handshakePacket(client *Client, data []byte) {
	token := binary.LittleEndian.Uint32(data)
	fmt.Println("Token recieved: ", token)

	// we would get the character from the hub server
	// if they are a guest it would make them a default char
	// otherwise fetch their real char
	tokens[token] = engine.DefaultCharacter(client.simulation, &client.send, token)
	// okay back to action

	client.id = token
	client.character = tokens[token]
	Clients[token] = client
	delete(tokens, token)

	engine.Worlds[0].AddCharacter(token, client.character)
	client.instance = engine.Worlds[0]
	go client.simulation.StartSimulation(client.id, client.instance, client.character)
	go client.UpdateCells()
}

func setCharacter(client *Client) {
	x, y := client.instance.GetBeachpoint()
	client.character.Move(x, y, 0)
	client.character.Apply()
	client.sendCharacter()
	client.sendCell()
}

func HandlePacket(client *Client, r io.Reader) {
	var p = make([]byte, 1)
	r.Read(p)

	var packet_type PacketType = PacketType(p[0])

	switch packet_type {
	case HANDSHAKE:
		var data = make([]byte, 4)
		r.Read(data)
		handshakePacket(client, data)
		setCharacter(client)

		fmt.Println("Handshake complete")
	case CHARACTER_POSITION:
		var x_bytes = make([]byte, 4)
		var y_bytes = make([]byte, 4)
		var angle_bytes = make([]byte, 2)
		r.Read(x_bytes)
		r.Read(y_bytes)
		r.Read(angle_bytes)

		x := math.Float32frombits(binary.LittleEndian.Uint32(x_bytes))
		y := math.Float32frombits(binary.LittleEndian.Uint32(y_bytes))
		angle := binary.LittleEndian.Uint16(angle_bytes)

		client.instance.MoveCharacter(client.character, x, y, angle)
	case CHARACTER_ATTACK:
		var x_bytes = make([]byte, 4)
		var y_bytes = make([]byte, 4)
		var target_x_bytes = make([]byte, 4)
		var target_y_bytes = make([]byte, 4)
		var angle_bytes = make([]byte, 2)
		r.Read(x_bytes)
		r.Read(y_bytes)
		r.Read(target_x_bytes)
		r.Read(target_y_bytes)
		r.Read(angle_bytes)

		x := math.Float32frombits(binary.LittleEndian.Uint32(x_bytes))
		y := math.Float32frombits(binary.LittleEndian.Uint32(y_bytes))
		targetX := math.Float32frombits(binary.LittleEndian.Uint32(target_x_bytes))
		targetY := math.Float32frombits(binary.LittleEndian.Uint32(target_y_bytes))
		angle := binary.LittleEndian.Uint16(angle_bytes)

		client.characterAttack(x, y, targetX, targetY, angle)
	case SELECT_SLOT:
		var idx_bytes = make([]byte, 1)
		r.Read(idx_bytes)
		client.selectSlot(idx_bytes[0])
	case CHANGE_INVENTORY:
		var toBytes = make([]byte, 1)
		var fromBytes = make([]byte, 1)
		r.Read(toBytes)
		r.Read(fromBytes)
		client.changeInventory(toBytes[0], fromBytes[0])
	case DROP_ITEM:
		var slotBytes = make([]byte, 1)
		r.Read(slotBytes)
		client.dropItem(slotBytes[0])
	case CHAT_MESSAGE:
		var msg_len = make([]byte, 1)
		r.Read(msg_len)

		var msg_bytes = make([]byte, int(msg_len[0]))
		r.Read(msg_bytes)

		client.ProcessMessage(string(msg_bytes))
	}
}
