package simulation

import "math"

const (
	MovementVectorSize         = 4
	ProjectileActionVectorSize = 3
	ActionVectorSize           = MovementVectorSize + ProjectileLimit*ProjectileActionVectorSize
)

type Movement struct {
	Up    float32
	Down  float32
	Left  float32
	Right float32
}

type ProjectileAction struct {
	Angle float32
	Power float32
}

type Action struct {
	Movement    Movement
	Count       uint8
	Projectiles [ProjectileLimit]ProjectileAction
}

type ActionVector struct {
	Count      uint8
	Continuous [ActionVectorSize]float32
}

type ProjectileLimits struct {
	MaximumDamage float32
	MinimumSpeed  float32
	MaximumSpeed  float32
}

type ResolvedProjectile struct {
	Angle  float32
	Damage float32
	Speed  float32
}

func (m Movement) Sanitize() Movement {
	return Movement{
		Up:    clamp(m.Up, 0, 1),
		Down:  clamp(m.Down, 0, 1),
		Left:  clamp(m.Left, 0, 1),
		Right: clamp(m.Right, 0, 1),
	}
}

func (m Movement) Velocity() Vector {
	m = m.Sanitize()
	return Vector{
		X: m.Right - m.Left,
		Y: m.Down - m.Up,
	}.limit(1).scale(MaximumSpeed)
}

func (m Movement) Half() Movement {
	return Movement{
		Up:    m.Up / 2,
		Down:  m.Down / 2,
		Left:  m.Left / 2,
		Right: m.Right / 2,
	}
}

func (m Movement) Sum() float32 {
	m = m.Sanitize()
	return m.Up + m.Down + m.Left + m.Right
}

func (a Action) Sanitize() Action {
	a.Movement = a.Movement.Sanitize()
	a.Count = min(a.Count, ProjectileLimit)
	for index := range a.Projectiles {
		a.Projectiles[index].Angle = normalizeAngle(a.Projectiles[index].Angle)
		a.Projectiles[index].Power = clamp(a.Projectiles[index].Power, 0, 1)
	}
	return a
}

func (a Action) Vector() ActionVector {
	a = a.Sanitize()
	vector := ActionVector{Count: a.Count}
	vector.Continuous[0] = a.Movement.Up
	vector.Continuous[1] = a.Movement.Down
	vector.Continuous[2] = a.Movement.Left
	vector.Continuous[3] = a.Movement.Right

	index := MovementVectorSize
	for projectileIndex := range int(a.Count) {
		projectile := a.Projectiles[projectileIndex]
		direction := angleDirection(projectile.Angle)
		vector.Continuous[index] = direction.X
		vector.Continuous[index+1] = direction.Y
		vector.Continuous[index+2] = projectile.Power
		index += ProjectileActionVectorSize
	}
	return vector
}

func (v ActionVector) Action() Action {
	action := Action{
		Count: min(v.Count, ProjectileLimit),
		Movement: Movement{
			Up:    v.Continuous[0],
			Down:  v.Continuous[1],
			Left:  v.Continuous[2],
			Right: v.Continuous[3],
		},
	}

	index := MovementVectorSize
	for projectileIndex := range int(action.Count) {
		action.Projectiles[projectileIndex] = ProjectileAction{
			Angle: angleDirectionInverse(Vector{
				X: v.Continuous[index],
				Y: v.Continuous[index+1],
			}),
			Power: v.Continuous[index+2],
		}
		index += ProjectileActionVectorSize
	}
	return action.Sanitize()
}

func (p ProjectileAction) Resolve(limits ProjectileLimits) ResolvedProjectile {
	power := clamp(p.Power, 0, 1)
	maximumDamage := max(clampFinite(limits.MaximumDamage), 0)
	minimumSpeed := max(clampFinite(limits.MinimumSpeed), 0)
	maximumSpeed := max(clampFinite(limits.MaximumSpeed), minimumSpeed)
	return ResolvedProjectile{
		Angle:  normalizeAngle(p.Angle),
		Damage: maximumDamage * power,
		Speed:  minimumSpeed + (maximumSpeed-minimumSpeed)*power,
	}
}

func (v Vector) scale(amount float32) Vector {
	return Vector{X: v.X * amount, Y: v.Y * amount}
}

func normalizeAngle(angle float32) float32 {
	if !finite(angle) {
		return 0
	}
	angle = float32(math.Mod(float64(angle), 360))
	if angle < 0 {
		angle += 360
	}
	return angle
}

func angleDirection(angle float32) Vector {
	radians := (normalizeAngle(angle) - 90) * math.Pi / 180
	return Vector{
		X: float32(math.Cos(float64(radians))),
		Y: float32(math.Sin(float64(radians))),
	}
}

func angleDirectionInverse(direction Vector) float32 {
	direction = direction.unit()
	if direction == (Vector{}) {
		return 0
	}
	radians := math.Atan2(float64(direction.Y), float64(direction.X))
	return normalizeAngle(float32(radians*180/math.Pi) + 90)
}
