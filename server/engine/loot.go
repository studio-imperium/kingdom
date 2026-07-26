package engine

import (
	"bytes"
	"encoding/binary"
	"math/rand/v2"
	"time"
)

type Loot struct {
	id    uint32
	loot  uint8
	x     float32
	y     float32
	timer float32
	Dead  bool
}

func (l *Loot) GetX() float32 { return l.x }
func (l *Loot) GetY() float32 { return l.y }

func CreateLoot(loot uint8, x float32, y float32) *Loot {
	id := rand.Uint32()
	x += rand.Float32() + 1
	y += rand.Float32() + 1
	return &Loot{
		id:    id,
		loot:  loot,
		x:     x,
		y:     y,
		timer: 50,
		Dead:  false,
	}
}

func (loot *Loot) Pack() []byte {
	data := new(bytes.Buffer)

	data.WriteByte(loot.loot)
	binary.Write(data, binary.LittleEndian, loot.x)
	binary.Write(data, binary.LittleEndian, loot.y)

	return data.Bytes()
}

func (character *Character) AddItemOrBust(loot uint8) bool {
	for i := uint8(0); i < 24; i++ {
		if _, ok := character.inventory[i]; !ok {
			character.inventory[i] = loot
			return true
		}
	}
	return false
}

func (loot *Loot) Looted() []byte {
	data := new(bytes.Buffer)

	data.WriteByte(12)
	binary.Write(data, binary.LittleEndian, loot.id)

	return data.Bytes()
}

func (loot *Loot) Tick(delta time.Duration) {
	secs := float32(delta) / float32(time.Second)
	loot.timer -= secs
}
