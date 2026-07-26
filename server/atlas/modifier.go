package atlas

import (
	"math"
	"math/rand"
	"sync"
)

type Modifier func(*Cell)

func (world *World) Infect(biomes []Biome, decay float64) {

	mindist := float64(world.Size)
	center := newPoint(world.Size/2, world.Size/2)
	var starting_cell *Cell
	for origin, cell := range world.cellByOrigin {
		if dist := distance(origin, center); dist < mindist {
			mindist = dist
			starting_cell = cell
		}
	}

	var changedCells []*Cell
	seen := make(map[*Cell]bool)
	seen[starting_cell] = true
	queue := []*Cell{
		starting_cell,
	}

	maxBiome := len(biomes) - 1
	currentBiome := 0.0
	convertBiome := func() uint8 {
		if int(currentBiome) > maxBiome {
			return uint8(maxBiome)
		}
		return uint8(math.Floor(currentBiome))
	}

	for len(queue) > 0 {
		currentBiome += decay
		var temp []*Cell

		for _, cell := range queue {
			biomeInt8 := convertBiome()
			if cell.biome <= biomeInt8 {
				changedCells = append(changedCells, cell)
				cell.biome = biomeInt8
			}

			for _, adj := range cell.GetAdjacentCells() {
				if !seen[adj] {
					seen[adj] = true
					temp = append(temp, adj)
				}
			}
		}

		queue = temp
	}

	wg := &sync.WaitGroup{}
	wg.Add(len(changedCells))
	for _, cell := range changedCells {
		biome := biomes[cell.biome]
		go func() {
			defer wg.Done()
			for _, mod := range biome.modifiers {
				cell.mu.Lock()
				mod(cell)
				for _, tile := range cell.Tiles {
					world.Tiles[tile.X+(tile.Y*world.Size)] = TileData{
						tile.Value,
						uint16(cell.Idx),
					}
				}
				cell.mu.Unlock()
			}
		}()
	}
	wg.Wait()
}

func NewFill(value uint8) Modifier {
	return func(cell *Cell) {
		for idx := range cell.Tiles {
			tile := &(cell.Tiles[idx])
			tile.Value = value
		}
	}
}

func NewCropCircle(angles float64, values ...uint8) Modifier {
	valuesLength := float64(len(values))

	getValue := func(pt Point, angle float64) float64 {
		angledX := math.Cos(angle) * pt.X
		angledY := math.Sin(angle) * pt.Y
		return math.Cos(angledX + angledY)
	}

	quasiCrystal := func(pt Point) uint8 {
		var value float64
		angle := 2.0 * 3.14156
		delta := angle / angles

		for i := 0; float64(i) < angles; i++ {
			value += getValue(pt, angle)
			angle -= delta
		}

		tileVal := value / angles
		tileVal += 1.0
		tileVal /= 2.0
		tileVal *= valuesLength

		return values[int(math.Max(0, math.Floor(tileVal)))%len(values)]
	}

	return func(cell *Cell) {
		origin := cell.Origin

		for idx := range cell.Tiles {
			tile := &(cell.Tiles[idx])
			tile.Value = quasiCrystal(tile.point().subtract(origin))
		}
	}
}

func NewPattern(angles float64, values ...uint8) Modifier {
	valuesLength := float64(len(values))

	getValue := func(pt Point, angle float64) float64 {
		angledX := math.Cos(angle) * pt.X
		angledY := math.Sin(angle) * pt.Y
		return math.Cos(angledX + angledY)
	}

	quasiCrystal := func(pt Point) uint8 {
		var value float64
		angle := 2.0 * 3.14156
		delta := angle / angles

		for i := 0; float64(i) < angles; i++ {
			value += getValue(pt, angle)
			angle -= delta
		}

		tileVal := value / angles
		tileVal += 1.0
		tileVal /= 2.0
		tileVal *= valuesLength

		return values[int(math.Max(0, math.Floor(tileVal)))%len(values)]
	}

	return func(cell *Cell) {
		for idx := range cell.Tiles {
			tile := &(cell.Tiles[idx])
			tile.Value = quasiCrystal(tile.point())
		}
	}
}

func NewVoronoi(density int, values ...uint8) Modifier {
	return func(cell *Cell) {
		rnd := rand.New(rand.NewSource(int64(density)))
		var origins []Point
		getPoint := func() Point {
			return cell.Tiles[rnd.Int()%len(cell.Tiles)].point()
		}
		getValue := func(seed int) uint8 {
			return values[seed%len(values)]
		}
		findNearest := func(pt Point) uint8 {
			nearest := origins[0]
			nearestDist := distance(nearest, pt)

			for idx := range origins {
				origin := origins[idx]
				originDist := distance(origin, pt)
				if nearestDist > originDist {
					nearest = origin
					nearestDist = originDist
				}
			}

			return getValue(int(nearest.X + nearest.Y))
		}

		for idx := 0; idx < density; idx++ {
			pt := getPoint()
			origins = append(origins, pt)
		}

		for idx := range cell.Tiles {
			tile := &(cell.Tiles[idx])
			tile.Value = findNearest(tile.point())
		}
	}
}

func NewBorder(border uint8) Modifier {
	return func(cell *Cell) {
		isBorder := func(cell *Cell, tile *Tile) bool {
			x := tile.X
			y := tile.Y
			adj := [4]Point{
				newPoint(x-1, y),
				newPoint(x+1, y),
				newPoint(x, y+1),
				newPoint(x, y-1),
			}
			for _, pt := range adj {
				if cell.grid[pt] == nil {
					return true
				}
			}

			return false
		}

		for idx := range cell.Tiles {
			tile := &(cell.Tiles[idx])

			if isBorder(cell, tile) {
				tile.Value = border
			}
		}
	}
}

func NewSelectiveBorder(border uint8, around uint8) Modifier {
	return func(cell *Cell) {
		isBorder := func(cell *Cell, tile *Tile) bool {
			x := tile.X
			y := tile.Y
			adj := [4]Point{
				newPoint(x-1, y),
				newPoint(x+1, y),
				newPoint(x, y+1),
				newPoint(x, y-1),
			}
			for _, pt := range adj {
				if cell.grid[pt] == nil || cell.grid[pt].Value != around {
					return true
				}
			}

			return false
		}

		var toChange []*Tile
		for idx := range cell.Tiles {
			tile := &(cell.Tiles[idx])

			if tile.Value == around && isBorder(cell, tile) {
				toChange = append(toChange, tile)
			}
		}

		for _, tile := range toChange {
			tile.Value = border
		}
	}
}

func NewSelectiveExternalBorder(border uint8, around uint8) Modifier {
	return func(cell *Cell) {
		isBorder := func(cell *Cell, tile *Tile) bool {
			x := tile.X
			y := tile.Y
			adj := [4]Point{
				newPoint(x-1, y),
				newPoint(x+1, y),
				newPoint(x, y+1),
				newPoint(x, y-1),
			}
			for _, pt := range adj {
				if cell.grid[pt] != nil && cell.grid[pt].Value == around {
					return true
				}
			}

			return false
		}

		var toChange []*Tile
		for idx := range cell.Tiles {
			tile := &(cell.Tiles[idx])

			if tile.Value != around && isBorder(cell, tile) {
				toChange = append(toChange, tile)
			}
		}

		for _, tile := range toChange {
			tile.Value = border
		}
	}
}
