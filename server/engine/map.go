package engine

import (
	"encoding/binary"
	"fmt"
	"io"
	"math/rand/v2"
)

type cell struct {
	id        uint16
	origin    Position
	biome     uint8
	tiles     []Tile
	adjacent  []uint16
	activated bool
}

type gameMap struct {
	size    uint16
	cells   []*cell
	beaches []Position
}

func loadMap(reader io.Reader) (*gameMap, error) {
	var cellCount uint16
	if err := binary.Read(reader, binary.LittleEndian, &cellCount); err != nil {
		return nil, fmt.Errorf("read cell count: %w", err)
	}
	if cellCount == 0 {
		return nil, fmt.Errorf("map has no cells")
	}

	gameMap := &gameMap{cells: make([]*cell, cellCount)}
	for index := range gameMap.cells {
		gameMap.cells[index] = &cell{
			id: uint16(index),
		}
	}

	for _, cell := range gameMap.cells {
		var x uint16
		var y uint16
		var adjacentCount uint8

		if err := binary.Read(reader, binary.LittleEndian, &x); err != nil {
			return nil, fmt.Errorf("read cell %d x: %w", cell.id, err)
		}
		if err := binary.Read(reader, binary.LittleEndian, &y); err != nil {
			return nil, fmt.Errorf("read cell %d y: %w", cell.id, err)
		}
		if err := binary.Read(reader, binary.LittleEndian, &cell.biome); err != nil {
			return nil, fmt.Errorf("read cell %d biome: %w", cell.id, err)
		}
		if err := binary.Read(reader, binary.LittleEndian, &adjacentCount); err != nil {
			return nil, fmt.Errorf("read cell %d adjacency count: %w", cell.id, err)
		}

		cell.origin = Position{X: float32(x), Y: float32(y)}
		if cell.biome == 10 {
			gameMap.beaches = append(gameMap.beaches, cell.origin)
		}

		cell.adjacent = make([]uint16, adjacentCount)
		for index := range cell.adjacent {
			if err := binary.Read(reader, binary.LittleEndian, &cell.adjacent[index]); err != nil {
				return nil, fmt.Errorf("read cell %d adjacency: %w", cell.id, err)
			}
			if int(cell.adjacent[index]) >= len(gameMap.cells) {
				return nil, fmt.Errorf("cell %d references cell %d", cell.id, cell.adjacent[index])
			}
		}
	}

	if err := binary.Read(reader, binary.LittleEndian, &gameMap.size); err != nil {
		return nil, fmt.Errorf("read map size: %w", err)
	}

	for y := uint16(0); y < gameMap.size; y++ {
		for x := uint16(0); x < gameMap.size; x++ {
			var tileType uint8
			var cellID uint16
			if err := binary.Read(reader, binary.LittleEndian, &tileType); err != nil {
				return nil, fmt.Errorf("read tile %d,%d: %w", x, y, err)
			}
			if err := binary.Read(reader, binary.LittleEndian, &cellID); err != nil {
				return nil, fmt.Errorf("read tile cell %d,%d: %w", x, y, err)
			}
			if int(cellID) >= len(gameMap.cells) {
				return nil, fmt.Errorf("tile %d,%d references cell %d", x, y, cellID)
			}
			gameMap.cells[cellID].tiles = append(gameMap.cells[cellID].tiles, Tile{
				X:    x,
				Y:    y,
				Type: tileType,
			})
		}
	}

	if len(gameMap.beaches) == 0 {
		return nil, fmt.Errorf("map has no beach cells")
	}

	return gameMap, nil
}

func (m *gameMap) beach() Position {
	return m.beaches[rand.IntN(len(m.beaches))]
}

func (m *gameMap) nearest(position Position) *cell {
	nearest := m.cells[0]
	nearestDistance := distance(position, nearest.origin)
	for _, candidate := range m.cells[1:] {
		candidateDistance := distance(position, candidate.origin)
		if candidateDistance < nearestDistance {
			nearest = candidate
			nearestDistance = candidateDistance
		}
	}
	return nearest
}

func (m *gameMap) nearby(position Position) []*cell {
	nearest := m.nearest(position)
	ids := map[uint16]struct{}{nearest.id: {}}

	for _, adjacent := range nearest.adjacent {
		ids[adjacent] = struct{}{}
		for _, outer := range m.cells[adjacent].adjacent {
			ids[outer] = struct{}{}
		}
	}

	cells := make([]*cell, 0, len(ids))
	for id := range ids {
		cells = append(cells, m.cells[id])
	}
	return cells
}

func (c *cell) state() CellState {
	tiles := make([]Tile, len(c.tiles))
	copy(tiles, c.tiles)
	return CellState{
		ID:      c.id,
		OriginX: uint16(c.origin.X),
		OriginY: uint16(c.origin.Y),
		Tiles:   tiles,
	}
}
