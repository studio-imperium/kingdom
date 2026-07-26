package main

import (
	"encoding/binary"
	"os"
	"path/filepath"

	"kingdoms/atlas"
)

var WATER uint8 = 1
var GRASS uint8 = 2
var WOOD uint8 = 3
var STONE uint8 = 4
var DRYGRASS uint8 = 5
var SAND uint8 = 6
var SANDSTONE uint8 = 7
var COLDGRASS uint8 = 8
var SNOW uint8 = 9
var ICE uint8 = 10
var GRAVEL uint8 = 11
var RUIN uint8 = 12
var LAVA uint8 = 13

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
	world.Infect(DesertOnly, 1)

	return world
}

func main() {
	world := CreateIsland(1500)
	path := filepath.Join("./", "world.map")
	f, _ := os.Create(path)

	defer f.Close()

	cells_len := uint16(len(world.Cells))

	binary.Write(f, binary.LittleEndian, cells_len)
	for _, cell := range world.Cells {
		binary.Write(f, binary.LittleEndian, uint16(cell.Origin.X))
		binary.Write(f, binary.LittleEndian, uint16(cell.Origin.Y))
		f.Write([]byte{cell.GetBiome()})

		adj := cell.GetAdjacentCells()
		binary.Write(f, binary.LittleEndian, uint8(len(adj)))
		for _, c := range adj {
			binary.Write(f, binary.LittleEndian, uint16(c.Idx))
		}
	}
	binary.Write(f, binary.LittleEndian, uint16(world.Size))

	for _, tiledata := range world.Tiles {
		binary.Write(f, binary.LittleEndian, tiledata.Type)
		binary.Write(f, binary.LittleEndian, tiledata.CellIdx)
	}
}
