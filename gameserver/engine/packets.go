package engine

const (
	HANDSHAKE uint8 = iota
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
	JOIN
	LEAVE
)

// Packet contains decoded input. Send is supplied only when core joins a player.
type Packet struct {
	Type                   uint8
	ID                     uint32
	CharacterID            int64
	Username               string
	Admin                  bool
	Send                   chan<- []byte
	X, Y, TargetX, TargetY float32
	Angle                  uint16
	Slot, From             uint8
	Message                string
}

func (engine *Engine) HandlePacket(packet Packet) {
	if packet.Type == JOIN {
		if _, exists := engine.Characters[packet.ID]; exists {
			close(packet.Send)
			return
		}
		character := DefaultCharacter(engine, packet.Send, packet.ID)
		character.characterID, character.username, character.admin = packet.CharacterID, packet.Username, packet.Admin
		engine.Characters[packet.ID] = character
		engine.addView(packet.ID)
		if engine.Map != nil {
			character.x, character.y = engine.Map.GetBeachpoint()
		}
		character.Apply()
		engine.UpdateCells()
		return
	}
	if packet.Type == LEAVE {
		engine.RemoveCharacter(packet.ID)
		return
	}
	character, exists := engine.Characters[packet.ID]
	if !exists || character.Dead || character.disconnected {
		return
	}
	switch packet.Type {
	case CHARACTER_POSITION:
		character.Move(packet.X, packet.Y, packet.Angle)
	case CHARACTER_ATTACK:
		character.characterAttack(packet.X, packet.Y, packet.TargetX, packet.TargetY, packet.Angle)
	case SELECT_SLOT:
		if packet.Slot < 24 {
			engine.SelectSlot(packet.ID, packet.Slot)
			character.Send(character.PackFull(HANDSHAKE))
		}
	case CHANGE_INVENTORY:
		if packet.Slot <= 25 && packet.From <= 25 {
			engine.ChangeInventory(packet.ID, packet.Slot, packet.From)
		}
	case DROP_ITEM:
		if packet.Slot <= 25 {
			engine.DropItem(packet.ID, packet.Slot)
		}
	case CHAT_MESSAGE:
		character.ProcessMessage(packet.Message)
	}
}
