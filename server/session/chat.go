package session

import (
	"strconv"
	"strings"

	"kingdoms/protocol"
)

func (s *Server) handleChat(client *client, message string) {
	words := strings.Fields(message)
	if len(words) == 0 {
		return
	}
	if !strings.HasPrefix(words[0], "/") {
		s.broadcast(protocol.EncodeChat(client.id, client.username, message))
		return
	}

	switch words[0] {
	case "/tp":
		s.teleport(client, words)
	case "/spawn":
		s.spawn(client, words)
	case "/loot":
		s.loot(client, words)
	default:
		client.send(protocol.EncodeChat(0, "System", "Invalid command"))
	}
}

func (s *Server) teleport(client *client, words []string) {
	if len(words) != 2 {
		client.send(protocol.EncodeChat(0, "System", "Usage: /tp <player>"))
		return
	}

	target := s.clientByUsername(words[1])
	if target == nil {
		client.send(protocol.EncodeChat(0, "System", "Player not found"))
		return
	}

	character, moved := s.world.Teleport(client.id, target.id)
	if !moved {
		return
	}
	client.send(protocol.EncodeCharacter(character))
	client.send(protocol.EncodeChat(0, "System", "Teleported to "+target.username))
}

func (s *Server) spawn(client *client, words []string) {
	if !client.admin || len(words) < 2 || len(words) > 3 {
		client.send(protocol.EncodeChat(0, "System", "Usage: /spawn [amount] <npc>"))
		return
	}

	amount := 1
	npcWord := words[1]
	if len(words) == 3 {
		var err error
		amount, err = strconv.Atoi(words[1])
		if err != nil || amount < 1 || amount > 100 {
			client.send(protocol.EncodeChat(0, "System", "Invalid amount"))
			return
		}
		npcWord = words[2]
	}

	npcID, err := strconv.ParseUint(npcWord, 10, 8)
	if err != nil {
		client.send(protocol.EncodeChat(0, "System", "Invalid NPC"))
		return
	}
	position, exists := s.world.CharacterPosition(client.id)
	if !exists {
		return
	}

	spawned := 0
	for range amount {
		if _, valid := s.world.SpawnNPC(uint8(npcID), position); valid {
			spawned++
		}
	}
	if spawned == 0 {
		client.send(protocol.EncodeChat(0, "System", "Invalid NPC"))
		return
	}
	client.send(protocol.EncodeChat(
		0,
		"System",
		"Spawned "+strconv.Itoa(spawned)+" "+strconv.FormatUint(npcID, 10),
	))
}

func (s *Server) loot(client *client, words []string) {
	if !client.admin || len(words) != 2 {
		client.send(protocol.EncodeChat(0, "System", "Usage: /loot <item>"))
		return
	}

	itemID, err := strconv.ParseUint(words[1], 10, 8)
	if err != nil {
		client.send(protocol.EncodeChat(0, "System", "Invalid item"))
		return
	}
	position, exists := s.world.CharacterPosition(client.id)
	if !exists {
		return
	}
	position.Y++
	if _, valid := s.world.SpawnLoot(uint8(itemID), client.id, position); !valid {
		client.send(protocol.EncodeChat(0, "System", "Invalid item"))
		return
	}
	client.send(protocol.EncodeChat(0, "System", "Looted "+words[1]))
}

func (s *Server) broadcast(packet []byte) {
	for _, client := range s.clientList() {
		client.send(packet)
	}
}

func (s *Server) clientByUsername(username string) *client {
	for _, client := range s.clientList() {
		if client.username == username {
			return client
		}
	}
	return nil
}
