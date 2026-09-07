package engine

import (
	"bytes"
	"encoding/binary"
	"time"
)

type Bomb struct {
	owner  uint32
	id     uint8
	evil   bool
	damage float32
	x      float32
	y      float32
	origin *Point
	timer  float32
	radius float32
	Dead   bool
}

func (b Bomb) GetX() float32 { return b.x }
func (b Bomb) GetY() float32 { return b.y }

func (bomb *Bomb) Pack() []byte {
	data := new(bytes.Buffer)

	data.WriteByte(bomb.id)
	binary.Write(data, binary.LittleEndian, bomb.x)
	binary.Write(data, binary.LittleEndian, bomb.y)
	binary.Write(data, binary.LittleEndian, bomb.origin.x)
	binary.Write(data, binary.LittleEndian, bomb.origin.y)

	return data.Bytes()
}

func (bomb *Bomb) Tick(delta time.Duration) {
	bomb.timer -= float32(delta) / float32(time.Second)
}

func DefaultBomb(id uint8, x float32, y float32, origin Object, evil bool, damage float32, timer float32) *Bomb {
	return &Bomb{
		id:     id,
		evil:   evil,
		damage: damage,
		x:      x,
		y:      y,
		origin: &Point{origin.GetX(), origin.GetY()},
		timer:  timer,
		Dead:   false,
	}
}
