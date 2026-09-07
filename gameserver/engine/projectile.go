package engine

import (
	"bytes"
	"encoding/binary"
	"math"
	"time"
)

type Projectile struct {
	owner   uint32
	id      uint8
	evil    bool
	damage  float32
	x       float32
	y       float32
	origin  *Point
	angle   uint16
	hitlist map[uint32]Object
	Dead    bool
}

func (p Projectile) GetX() float32 { return p.x }
func (p Projectile) GetY() float32 { return p.y }

func (projectile *Projectile) Pack() []byte {
	data := new(bytes.Buffer)

	data.WriteByte(projectile.id)
	binary.Write(data, binary.LittleEndian, projectile.x)
	binary.Write(data, binary.LittleEndian, projectile.y)
	binary.Write(data, binary.LittleEndian, projectile.angle)

	return data.Bytes()
}

func (projectile *Projectile) Tick(delta time.Duration) {
	deltaMs := float32(delta) / float32(time.Millisecond)
	deltaTime := deltaMs * (60.0 / 1000.0)
	speed := float32(projectileData[projectile.id].Speed)
	rad := (float32(projectile.angle) - 90) * (math.Pi / 180)
	dx := float32(math.Cos(float64(rad)))
	dy := float32(math.Sin(float64(rad)))

	projectile.x += dx * speed * deltaTime / 16
	projectile.y += dy * speed * deltaTime / 16
}

func DefaultProjectile(id uint8, x float32, y float32, angle uint16, evil bool, damage float32) *Projectile {
	return &Projectile{
		id:      id,
		evil:    evil,
		damage:  damage,
		x:       x,
		y:       y,
		origin:  &Point{x, y},
		angle:   angle,
		hitlist: make(map[uint32]Object),
		Dead:    false,
	}
}
