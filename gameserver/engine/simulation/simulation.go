package simulation

import (
	"bytes"
	"encoding/binary"
)

const RenderDistance = 16

// A simulation contains only the packets visible to one player, never world entities.
type Simulation struct {
	Characters      map[uint32][]byte
	Npcs            map[uint32][]byte
	Loot            map[uint32][]byte
	Cells           map[uint16]bool
	KnownCharacters map[uint32]bool
	x, y            float32
}

func CreateSimulation() *Simulation {
	return &Simulation{
		Characters:      make(map[uint32][]byte),
		Npcs:            make(map[uint32][]byte),
		Loot:            make(map[uint32][]byte),
		Cells:           make(map[uint16]bool),
		KnownCharacters: make(map[uint32]bool),
	}
}

func (simulation *Simulation) Reset(x, y float32) {
	simulation.x, simulation.y = x, y
	clear(simulation.Characters)
	clear(simulation.Npcs)
	clear(simulation.Loot)
}

func (simulation *Simulation) Visible(x, y float32) bool {
	dx, dy := x-simulation.x, y-simulation.y
	return dx*dx+dy*dy <= RenderDistance*RenderDistance
}

func (simulation *Simulation) Pack() []byte {
	data := new(bytes.Buffer)
	data.WriteByte(4)
	for _, entities := range []map[uint32][]byte{simulation.Characters, simulation.Npcs, simulation.Loot} {
		binary.Write(data, binary.LittleEndian, uint16(len(entities)))
		for id, packet := range entities {
			binary.Write(data, binary.LittleEndian, id)
			data.Write(packet)
		}
	}
	return data.Bytes()
}
