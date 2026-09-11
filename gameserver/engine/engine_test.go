package engine

import (
	"encoding/binary"
	"testing"
	"time"
)

func joinPlayer(world *Engine, id uint32) chan []byte {
	output := make(chan []byte, 256)
	world.HandlePacket(Packet{Type: JOIN, ID: id, CharacterID: int64(id), Send: output})
	return output
}

func TestJoinNameInChat(t *testing.T) {
	if err := InitAssets(testAssets.URL + "/"); err != nil {
		t.Fatal(err)
	}
	world := CreateEngine()
	output := make(chan []byte, 256)
	world.HandlePacket(Packet{Type: JOIN, ID: 1, Username: "PeckishNatureSpirit", Send: output})
	<-output // Initial character state.
	world.HandlePacket(Packet{Type: CHAT_MESSAGE, ID: 1, Message: "hello"})
	message := <-output
	if message[0] != CHAT_MESSAGE || string(message[6:6+int(message[5])]) != "PeckishNatureSpirit" {
		t.Fatal("chat did not use the join name")
	}
}

func TestFinalCharacterState(t *testing.T) {
	if err := InitAssets(testAssets.URL + "/"); err != nil {
		t.Fatal(err)
	}
	for _, dead := range []bool{false, true} {
		world := CreateEngine()
		output, finished := make(chan []byte, 256), make(chan CharacterData, 1)
		saved := &CharacterData{Id: 1234567890123456789, Level: 999, Exp: 120, Body: 1, Inventory: map[uint8]uint8{0: 8}}
		world.HandlePacket(Packet{Type: JOIN, ID: 1, CharacterID: saved.Id, Username: "PeckishNatureSpirit", Character: saved, Send: output, Finished: finished})
		if world.Characters[1].Level != 2 {
			t.Fatal("loaded level was not derived from EXP")
		}
		initial := <-output
		expOffset := 31 + int(initial[26])*2
		if len(initial) != expOffset+8 || int64(binary.LittleEndian.Uint64(initial[expOffset:])) != 120 {
			t.Fatal("initial character packet missing saved EXP")
		}
		world.Characters[1].AwardExp(105)
		update := <-output
		if update[0] != HANDSHAKE || int64(binary.LittleEndian.Uint64(update[31+int(update[26])*2:])) != 225 {
			t.Fatal("earned EXP was not sent to the client")
		}
		world.ChangeInventory(1, 1, 0)
		if dead {
			world.Characters[1].Damage(1000)
			world.Tick(time.Millisecond)
		} else {
			world.HandlePacket(Packet{Type: LEAVE, ID: 1})
		}
		select {
		case state := <-finished:
			if state.Id != saved.Id || state.Username != "PeckishNatureSpirit" || state.Level != 3 || state.Exp != 225 || state.Dead != dead || state.Inventory[1] != 8 || state.Inventory[0] != 0 {
				t.Fatalf("bad final state: %+v", state)
			}
		case <-time.After(time.Second):
			t.Fatal("no final character state")
		}
		world.HandlePacket(Packet{Type: LEAVE, ID: 1})
		if len(finished) != 0 {
			t.Fatal("character was saved twice")
		}
	}
}

func TestVisibilityDoesNotChangeWorld(t *testing.T) {
	if err := InitAssets(testAssets.URL + "/"); err != nil {
		t.Fatal(err)
	}
	world := CreateEngine()
	first, second := joinPlayer(world, 1), joinPlayer(world, 2)
	<-first
	<-second
	world.Characters[2].Move(100, 100, 0)
	world.SendWorldState()
	if count := binary.LittleEndian.Uint16((<-first)[1:]); count != 1 {
		t.Fatalf("visible characters = %d", count)
	}
	if world.Characters[2].Dead || len(world.Characters) != 2 {
		t.Fatal("culling changed world state")
	}
	departing := world.Characters[1]
	_, npc := world.SpawnNpc(0, 0, 0)
	npc.nearby[1], npc.target, npc.looking = departing, departing, departing
	world.RemoveCharacter(1)
	world.RemoveCharacter(1)
	if departing.Dead || npc.target != nil || npc.looking != nil || len(npc.nearby) != 0 {
		t.Fatal("disconnect left combat references or caused death")
	}
	if world.Characters[2].Dead {
		t.Fatal("departure killed another player")
	}
	world.RemoveCharacter(2)
}

func TestWorldCombat(t *testing.T) {
	if err := InitAssets(testAssets.URL + "/"); err != nil {
		t.Fatal(err)
	}
	world := CreateEngine()
	joinPlayer(world, 1)
	joinPlayer(world, 2)
	first, second := world.Characters[1], world.Characters[2]
	first.health, second.health = 100, 100
	first.maxHealth, second.maxHealth = 100, 100
	// An expired hostile bomb damages both players exactly once.
	id := world.CreateBomb(0, 0, 0, first, true, 5, 0)
	world.Tick(0)
	if first.health != 95 || second.health != 95 {
		t.Fatalf("bomb health: %v, %v", first.health, second.health)
	}
	if _, exists := world.Bombs[id]; exists {
		t.Fatal("exploded bomb retained")
	}
	world.Tick(0)
	if first.health != 95 || second.health != 95 {
		t.Fatal("bomb applied twice")
	}

	// A non-piercing projectile must hit only one of two overlapping NPCs.
	projectileData[0].Piercing = false
	_, a := world.SpawnNpc(0, 100, 100)
	_, b := world.SpawnNpc(0, 100, 100)
	healthA, healthB := a.health, b.health
	id = world.CreateProjectile(0, 100, 100, 0, false, 1)
	world.Projectiles[id].owner = 1
	world.Tick(0)
	if (healthA-a.health)+(healthB-b.health) != 1 {
		t.Fatal("projectile hit more than once")
	}
	if _, exists := world.Projectiles[id]; exists {
		t.Fatal("spent projectile retained")
	}
	if a.damage[1]+b.damage[1] != 1 {
		t.Fatal("damage credit lost")
	}

	// Piercing projectiles remember prior hits even across ticks.
	projectileData[0].Piercing = true
	id = world.CreateProjectile(0, 100, 100, 0, false, 1)
	world.Tick(0)
	healthA, healthB = a.health, b.health
	world.Tick(0)
	if a.health != healthA || b.health != healthB {
		t.Fatal("piercing projectile repeated damage")
	}
	world.Projectiles[id].x = 10000
	world.Tick(0)
	if _, exists := world.Projectiles[id]; exists {
		t.Fatal("out-of-range projectile retained")
	}
}

func TestLootEligibilityPickupAndExpiry(t *testing.T) {
	if err := InitAssets(testAssets.URL + "/"); err != nil {
		t.Fatal(err)
	}
	world := CreateEngine()
	joinPlayer(world, 1)
	joinPlayer(world, 2)
	first, second := world.Characters[1], world.Characters[2]
	loot := CreateLoot(8, 0, 0)
	loot.x, loot.y = 0, 0
	loot.eligible[1] = true
	world.Loot[loot.id] = loot
	first.Move(10, 10, 0)
	world.Tick(0)
	if len(second.inventory) != 1 || len(world.Loot) != 1 {
		t.Fatal("ineligible player picked up personal loot")
	}
	first.Move(0, 0, 0)
	world.Tick(0)
	if len(first.inventory) != 2 || len(world.Loot) != 0 {
		t.Fatal("eligible player did not pick up loot")
	}
	loot = CreateLoot(8, 0, 0)
	loot.x, loot.y = 0, 0
	loot.eligible[1], loot.eligible[2] = true, true
	world.Loot[loot.id] = loot
	world.Tick(0)
	if len(first.inventory)+len(second.inventory) != 4 {
		t.Fatal("shared loot was duplicated")
	}
	loot = CreateLoot(8, 100, 100)
	loot.eligible[1] = true
	loot.timer = 0
	world.Loot[loot.id] = loot
	world.Tick(0)
	if len(world.Loot) != 0 {
		t.Fatal("expired loot retained")
	}
}

func TestDeathAndSlowConnection(t *testing.T) {
	if err := InitAssets(testAssets.URL + "/"); err != nil {
		t.Fatal(err)
	}
	world := CreateEngine()
	output := joinPlayer(world, 1)
	world.Characters[1].Damage(10000)
	world.Tick(0)
	death := false
	for packet := range output {
		if packet[0] == CHARACTER_DEAD {
			death = true
		}
	}
	if !death || len(world.Characters) != 0 {
		t.Fatal("death packet or cleanup missing")
	}

	output = make(chan []byte, 1)
	world.HandlePacket(Packet{Type: JOIN, ID: 2, Send: output})
	world.Characters[2].Send([]byte{10})
	world.Tick(0)
	if len(world.Characters) != 0 {
		t.Fatal("blocked connection retained")
	}
	for range output {
	}
}

func TestIslandLoadsAndTicks(t *testing.T) {
	if err := InitAssets(testAssets.URL + "/"); err != nil {
		t.Fatal(err)
	}
	world, err := CreateIsland()
	if err != nil {
		t.Fatal(err)
	}
	output := joinPlayer(world, 1)
	if len(world.Map.Cells) == 0 {
		t.Fatal("empty map")
	}
	packet := <-output
	if packet[0] != HANDSHAKE || binary.LittleEndian.Uint32(packet[27+int(packet[26])*2:]) != 1 {
		t.Fatal("missing character identity")
	}
	tiles := 0
	for len(output) > 0 {
		if (<-output)[0] == TILES {
			tiles++
		}
	}
	if tiles == 0 {
		t.Fatal("no tiles sent")
	}
	for i := 0; i < 100; i++ {
		world.Tick(50 * time.Millisecond)
		for len(output) > 0 {
			<-output
		}
	}
	world.RemoveCharacter(1)
}
