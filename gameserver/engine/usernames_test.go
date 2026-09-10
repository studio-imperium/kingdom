package engine

import (
	"encoding/binary"
	"testing"
)

func TestWorldUsernames(t *testing.T) {
	if err := InitAssets(testAssets.URL + "/"); err != nil {
		t.Fatal(err)
	}
	world := CreateEngine()
	output := joinPlayer(world, 1)
	joinPlayer(world, 2)
	<-output
	world.Characters[2].username = "Éowyn 🐉"
	world.Characters[2].Move(100, 100, 0)
	world.SendWorldState()
	if packet := <-output; len(packet) != 26 {
		t.Fatal("sent a name for an invisible character")
	}
	world.Characters[2].Move(0, 0, 0)
	// A full output queue must not mark an unsent name as delivered.
	for len(output) < cap(output) {
		output <- nil
	}
	world.SendWorldState()
	if world.simulations[1].KnownCharacters[2] {
		t.Fatal("dropped name marked as delivered")
	}
	for len(output) > 0 {
		<-output
	}
	world.SendWorldState()
	packet := <-output
	const offset = 7 + 19*2 // Existing world fields, including both characters.
	name := world.Characters[2].username
	if len(packet) != offset+7+len(name) || binary.LittleEndian.Uint16(packet[offset:]) != 1 ||
		binary.LittleEndian.Uint32(packet[offset+2:]) != 2 || int(packet[offset+6]) != len(name) || string(packet[offset+7:]) != name {
		t.Fatalf("incorrect UTF-8 username trailer: %v", packet[offset:])
	}
	world.SendWorldState()
	if packet := <-output; len(packet) != offset {
		t.Fatal("username resent on subsequent world update")
	}
	world.Characters[2].Move(100, 100, 0)
	world.SendWorldState()
	<-output
	world.Characters[2].Move(0, 0, 0)
	world.SendWorldState()
	if packet := <-output; len(packet) != offset {
		t.Fatal("cached username resent on reentry")
	}
}
