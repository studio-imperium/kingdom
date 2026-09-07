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

func TestVisibilityDoesNotChangeWorld(t *testing.T) {
	if err := InitAssets(); err != nil {
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
	if err := InitAssets(); err != nil {
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
	if err := InitAssets(); err != nil {
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
	if err := InitAssets(); err != nil {
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
	if err := InitAssets(); err != nil {
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
	if packet[0] != HANDSHAKE || binary.LittleEndian.Uint32(packet[len(packet)-4:]) != 1 {
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
