package maps

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"math/rand/v2"
)

type Tile struct {
	X, Y uint16
	Val  uint8
}

type Origin struct{ X, Y uint16 }

type Cell struct {
	Idx      uint16
	Tiles    []Tile
	Origin   Origin
	Biome    uint8
	Adjacent []*Cell
}

type Map struct {
	Size        uint16
	Cells       []*Cell
	beachpoints []Origin
}

func Load(r io.Reader) (*Map, error) {
	var count uint16
	if err := binary.Read(r, binary.LittleEndian, &count); err != nil {
		return nil, err
	}
	if count == 0 {
		return nil, fmt.Errorf("map has no cells")
	}
	m := &Map{Cells: make([]*Cell, count)}
	for i := range m.Cells {
		m.Cells[i] = &Cell{Idx: uint16(i)}
	}
	for _, cell := range m.Cells {
		var header struct {
			X, Y            uint16
			Biome, Adjacent uint8
		}
		if err := binary.Read(r, binary.LittleEndian, &header); err != nil {
			return nil, err
		}
		cell.Origin, cell.Biome = Origin{header.X, header.Y}, header.Biome
		if cell.Biome == 10 {
			m.beachpoints = append(m.beachpoints, cell.Origin)
		}
		for i := uint8(0); i < header.Adjacent; i++ {
			var idx uint16
			if err := binary.Read(r, binary.LittleEndian, &idx); err != nil {
				return nil, err
			}
			if idx >= count {
				return nil, fmt.Errorf("invalid adjacent cell %d", idx)
			}
			cell.Adjacent = append(cell.Adjacent, m.Cells[idx])
		}
	}
	if err := binary.Read(r, binary.LittleEndian, &m.Size); err != nil {
		return nil, err
	}
	if m.Size == 0 {
		return nil, fmt.Errorf("map has no tiles")
	}
	for y := 0; y < int(m.Size); y++ {
		for x := 0; x < int(m.Size); x++ {
			var tile struct {
				Val  uint8
				Cell uint16
			}
			if err := binary.Read(r, binary.LittleEndian, &tile); err != nil {
				return nil, err
			}
			if tile.Cell >= count {
				return nil, fmt.Errorf("invalid tile cell %d", tile.Cell)
			}
			cell := m.Cells[tile.Cell]
			cell.Tiles = append(cell.Tiles, Tile{uint16(x), uint16(y), tile.Val})
		}
	}
	return m, nil
}

func (m *Map) GetNearestCell(x, y float32) *Cell {
	nearest := m.Cells[0]
	distance := math.Inf(1)
	for _, cell := range m.Cells {
		d := math.Hypot(float64(x)-float64(cell.Origin.X), float64(y)-float64(cell.Origin.Y))
		if d < distance {
			nearest, distance = cell, d
		}
	}
	return nearest
}

func (m *Map) GetBeachpoint() (float32, float32) {
	point := m.Cells[0].Origin
	if len(m.beachpoints) > 0 {
		point = m.beachpoints[rand.IntN(len(m.beachpoints))]
	}
	return float32(point.X), float32(point.Y)
}

func (cell *Cell) Pack() []byte {
	data := new(bytes.Buffer)
	data.WriteByte(6)
	binary.Write(data, binary.LittleEndian, int32(cell.Origin.X))
	binary.Write(data, binary.LittleEndian, int32(cell.Origin.Y))
	binary.Write(data, binary.LittleEndian, uint16(len(cell.Tiles)))
	for _, tile := range cell.Tiles {
		binary.Write(data, binary.LittleEndian, int32(tile.X))
		binary.Write(data, binary.LittleEndian, int32(tile.Y))
		data.WriteByte(tile.Val)
	}
	return data.Bytes()
}
