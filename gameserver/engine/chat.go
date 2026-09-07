package engine

import (
	"bytes"
	"encoding/binary"
	"strconv"
	"strings"
)

func (character *Character) SendMessage(id uint32, sender, msg string) {
	if len(sender) > 255 {
		sender = sender[:255]
	}
	if len(msg) > 255 {
		msg = msg[:255]
	}
	data := new(bytes.Buffer)
	data.WriteByte(CHAT_MESSAGE)
	binary.Write(data, binary.LittleEndian, id)
	data.WriteByte(uint8(len(sender)))
	data.WriteString(sender)
	data.WriteByte(uint8(len(msg)))
	data.WriteString(msg)
	character.Send(data.Bytes())
}

func (character *Character) ProcessMessage(msg string) {
	words := strings.Fields(msg)
	if len(words) == 0 {
		return
	}
	if msg[0] != '/' {
		for _, recipient := range character.instance.Characters {
			recipient.SendMessage(character.id, character.username, msg)
		}
		return
	}
	if words[0] == "/tp" && len(words) == 2 {
		for _, other := range character.instance.Characters {
			if words[1] == other.username {
				character.Move(other.x, other.y, 0)
				character.Apply()
				character.SendMessage(1, "System", "Teleported to "+other.username)
				return
			}
		}
	}
	if character.admin && len(words) >= 2 {
		which, err := strconv.Atoi(words[1])
		if err == nil && which >= 0 {
			switch words[0] {
			case "/spawn":
				amount := 1
				if len(words) == 3 {
					amount, err = strconv.Atoi(words[2])
				}
				if err == nil && which < len(npcData) && amount > 0 && amount <= 100 {
					for i := 0; i < amount; i++ {
						character.instance.SpawnNpc(uint8(which), character.x, character.y)
					}
					character.SendMessage(1, "System", "Spawned "+strconv.Itoa(amount)+" "+strconv.Itoa(which))
					return
				}
			case "/loot":
				if which < len(itemData) {
					loot := CreateLoot(uint8(which), character.x, character.y+1)
					loot.eligible[character.id] = true
					character.instance.Loot[loot.id] = loot
					character.SendMessage(1, "System", "Looted "+strconv.Itoa(which))
					return
				}
			}
		}
	}
	character.SendMessage(0, "System", "Invalid command")
}
