package main

import (
	"bytes"
	"encoding/binary"
	"io"
	"log"
	"os"

	"github.com/studio-imperium/atlas"
)

// Tiles
const WATER = 1
const GRASS = 2
const WOOD = 3
const STONE = 4
const DRYGRASS = 5
const SAND = 6
const SANDSTONE = 7
const COLDGRASS = 8
const SNOW = 9
const ICE = 10
const GRAVEL = 11
const RUIN = 12
const LAVA = 13

var Beach []atlas.Biome = []atlas.Biome{
	atlas.NewBiome(
		atlas.NewFill(GRASS),
	),
	atlas.NewBiome(
		atlas.NewFill(GRASS),
	),
	atlas.NewBiome(
		atlas.NewVoronoi(6, GRASS, SAND),
		atlas.NewSelectiveBorder(SAND, GRASS),
		atlas.NewBorder(SAND),
		atlas.NewSelectiveExternalBorder(SAND, SAND),
	),
	atlas.NewBiome(
		atlas.NewFill(WATER),
	),
}
var Sandy []atlas.Biome = []atlas.Biome{
	atlas.NewBiome(
		atlas.NewCropCircle(7, SANDSTONE, DRYGRASS),
		atlas.NewSelectiveBorder(SAND, DRYGRASS),
		atlas.NewSelectiveBorder(SAND, SANDSTONE),
		atlas.NewSelectiveBorder(SANDSTONE, DRYGRASS),
		atlas.NewSelectiveBorder(SANDSTONE, DRYGRASS),
	),
	atlas.NewBiome(
		atlas.NewCropCircle(3.5, DRYGRASS, SAND),
		atlas.NewSelectiveBorder(SANDSTONE, SAND),
	),
	atlas.NewBiome(
		atlas.NewVoronoi(20, SANDSTONE, DRYGRASS),
		atlas.NewSelectiveBorder(SAND, SANDSTONE),
		atlas.NewSelectiveBorder(SAND, SANDSTONE),
	),
}
var Sandy2 []atlas.Biome = []atlas.Biome{
	atlas.NewBiome(
		atlas.NewCropCircle(3.5, DRYGRASS, SAND),
		atlas.NewSelectiveBorder(SANDSTONE, SAND),
	),
	atlas.NewBiome(
		atlas.NewVoronoi(20, SANDSTONE, DRYGRASS),
		atlas.NewSelectiveBorder(SAND, SANDSTONE),
		atlas.NewSelectiveBorder(SAND, SANDSTONE),
	),
}
var Sandy3 []atlas.Biome = []atlas.Biome{
	atlas.NewBiome(
		atlas.NewCropCircle(3.5, DRYGRASS, SAND),
		atlas.NewSelectiveBorder(SANDSTONE, SAND),
	),
	atlas.NewBiome(
		atlas.NewVoronoi(20, SANDSTONE, DRYGRASS),
		atlas.NewSelectiveBorder(SAND, SANDSTONE),
		atlas.NewSelectiveBorder(SAND, SANDSTONE),
	),
	atlas.NewBiome(
		atlas.NewCropCircle(7, SANDSTONE, DRYGRASS),
		atlas.NewSelectiveBorder(SAND, DRYGRASS),
		atlas.NewSelectiveBorder(SAND, SANDSTONE),
		atlas.NewSelectiveBorder(SANDSTONE, DRYGRASS),
		atlas.NewSelectiveBorder(SANDSTONE, DRYGRASS),
	),
}
var Snowy []atlas.Biome = []atlas.Biome{
	atlas.NewBiome(
		atlas.NewCropCircle(3, ICE, SNOW),
		atlas.NewSelectiveBorder(SNOW, ICE),
	),
	atlas.NewBiome(
		atlas.NewVoronoi(10, ICE, COLDGRASS),
		atlas.NewSelectiveBorder(SNOW, ICE),
		atlas.NewSelectiveBorder(SNOW, ICE),
	),
}
var Snowy2 []atlas.Biome = []atlas.Biome{
	atlas.NewBiome(
		atlas.NewPattern(10, ICE, COLDGRASS),
		atlas.NewSelectiveBorder(SNOW, ICE),
	),
	atlas.NewBiome(
		atlas.NewPattern(4.1, ICE, COLDGRASS),
		atlas.NewSelectiveBorder(SNOW, ICE),
		atlas.NewSelectiveBorder(SNOW, COLDGRASS),
	),
}
var Glaciers []atlas.Biome = []atlas.Biome{
	atlas.NewBiome(
		atlas.NewVoronoi(40, STONE, STONE, STONE, STONE, STONE, STONE, STONE, SNOW),
		atlas.NewSelectiveBorder(STONE, SNOW),
		atlas.NewBorder(SNOW),
	),
	atlas.NewBiome(
		atlas.NewVoronoi(40, STONE, STONE, STONE, SNOW, ICE),
		atlas.NewSelectiveBorder(STONE, SNOW),
		atlas.NewSelectiveBorder(SNOW, ICE),
		atlas.NewBorder(SNOW),
	),
	atlas.NewBiome(
		atlas.NewPattern(27, STONE, SNOW),
		atlas.NewSelectiveBorder(SNOW, ICE),
		atlas.NewSelectiveBorder(STONE, STONE),
		atlas.NewBorder(ICE),
		atlas.NewSelectiveExternalBorder(SNOW, ICE),
	),
}
var Hot []atlas.Biome = []atlas.Biome{
	atlas.NewBiome(
		atlas.NewPattern(3.3, LAVA, GRAVEL),
	),
	atlas.NewBiome(
		atlas.NewPattern(15, LAVA, RUIN),
		atlas.NewSelectiveBorder(GRAVEL, LAVA),
	),
	atlas.NewBiome(
		atlas.NewVoronoi(20, RUIN, LAVA),
		atlas.NewSelectiveBorder(GRAVEL, LAVA),
	),
}

func AppendBiomes(biomes ...[]atlas.Biome) []atlas.Biome {
	if len(biomes) == 1 {
		return biomes[0]
	} else {
		return append(biomes[0], AppendBiomes(biomes[1:]...)...)
	}
}

var Island = AppendBiomes(Hot, Hot, Glaciers, Snowy2, Snowy, Sandy, Sandy2, Sandy3, Beach)
var DesertOnly = AppendBiomes(Sandy, Sandy2, Sandy3, Beach)

func CreateIsland(size int) *atlas.World {
	world := atlas.NewWorld(size, 2000, 11)
	world.InfectFrom(DesertOnly, 1, atlas.Point{X: float64(size / 2), Y: float64(size / 2)})

	return world
}

func main() {
	var data bytes.Buffer
	writeWorld(&data, CreateIsland(1500))
	if err := os.WriteFile("world.map", data.Bytes(), 0644); err != nil {
		log.Fatal(err)
	}
}

func writeWorld(f io.Writer, world *atlas.World) {
	cellIndices := make(map[*atlas.Cell]uint16, len(world.Cells))
	for idx, cell := range world.Cells {
		cellIndices[cell] = uint16(idx)
	}
	type tileData struct {
		Type    uint8
		CellIdx uint16
	}
	tiles := make([]tileData, world.Size*world.Size)

	cells_len := uint16(len(world.Cells))

	binary.Write(f, binary.LittleEndian, cells_len)
	for _, cell := range world.Cells {
		binary.Write(f, binary.LittleEndian, uint16(cell.Origin.X))
		binary.Write(f, binary.LittleEndian, uint16(cell.Origin.Y))
		f.Write([]byte{uint8(cell.GetBiome())})

		adj := cell.GetAdjacentCells()
		binary.Write(f, binary.LittleEndian, uint8(len(adj)))
		for _, c := range adj {
			binary.Write(f, binary.LittleEndian, cellIndices[c])
		}
		for _, tile := range cell.Tiles {
			tiles[tile.X+tile.Y*world.Size] = tileData{uint8(tile.Value), cellIndices[cell]}
		}
	}
	binary.Write(f, binary.LittleEndian, uint16(world.Size))

	for _, tiledata := range tiles {
		binary.Write(f, binary.LittleEndian, tiledata.Type)
		binary.Write(f, binary.LittleEndian, tiledata.CellIdx)
	}
}
