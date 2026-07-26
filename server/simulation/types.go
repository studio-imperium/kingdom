package simulation

const (
	NearbyLimit     = 16
	ProjectileLimit = 16
	MaximumSpeed    = 16
)

type Vector struct {
	X float32
	Y float32
}

type EntityState struct {
	ID       uint32
	Position Vector
	Friendly bool
}

type ProjectileState struct {
	ID        uint32
	Position  Vector
	Direction Vector
	Damage    float32
	Friendly  bool
}

type BombState struct {
	ID       uint32
	Position Vector
	Damage   float32
	Time     float32
	Friendly bool
}

type Frame struct {
	Entities    []EntityState
	Projectiles []ProjectileState
	Bombs       []BombState
}

type Observer struct {
	ID       uint32
	Position Vector
	Friendly bool
}
