package simulation

import "sort"

const (
	EntityVectorSize     = 4
	ProjectileVectorSize = 6
	BombVectorSize       = 5
	ObservationSize      = NearbyLimit * (EntityVectorSize + ProjectileVectorSize + BombVectorSize)
)

type Entity struct {
	Exists   bool
	Position Vector
	Friendly bool
}

type Projectile struct {
	Exists    bool
	Position  Vector
	Direction Vector
	Damage    float32
}

type Bomb struct {
	Exists   bool
	Position Vector
	Damage   float32
	Time     float32
}

type Observation struct {
	Entities    [NearbyLimit]Entity
	Projectiles [NearbyLimit]Projectile
	Bombs       [NearbyLimit]Bomb
}

type ObservationScale struct {
	Distance float32
	Damage   float32
	Time     float32
}

type ObservationVector [ObservationSize]float32

func DefaultObservationScale() ObservationScale {
	return ObservationScale{
		Distance: 16,
		Damage:   5,
		Time:     2,
	}
}

func Observe(frame Frame, observer Observer) Observation {
	var observation Observation

	entities := nearest(
		frame.Entities,
		observer.Position,
		func(entity EntityState) bool {
			return entity.ID != observer.ID
		},
		func(entity EntityState) uint32 {
			return entity.ID
		},
		func(entity EntityState) Vector {
			return entity.Position
		},
	)
	for index, entity := range entities {
		observation.Entities[index] = Entity{
			Exists:   true,
			Position: entity.Position.relativeTo(observer.Position),
			Friendly: entity.Friendly == observer.Friendly,
		}
	}

	projectiles := nearest(
		frame.Projectiles,
		observer.Position,
		func(projectile ProjectileState) bool {
			return projectile.Friendly != observer.Friendly
		},
		func(projectile ProjectileState) uint32 {
			return projectile.ID
		},
		func(projectile ProjectileState) Vector {
			return projectile.Position
		},
	)
	for index, projectile := range projectiles {
		observation.Projectiles[index] = Projectile{
			Exists:    true,
			Position:  projectile.Position.relativeTo(observer.Position),
			Direction: projectile.Direction.unit(),
			Damage:    projectile.Damage,
		}
	}

	bombs := nearest(
		frame.Bombs,
		observer.Position,
		func(bomb BombState) bool {
			return bomb.Friendly != observer.Friendly
		},
		func(bomb BombState) uint32 {
			return bomb.ID
		},
		func(bomb BombState) Vector {
			return bomb.Position
		},
	)
	for index, bomb := range bombs {
		observation.Bombs[index] = Bomb{
			Exists:   true,
			Position: bomb.Position.relativeTo(observer.Position),
			Damage:   bomb.Damage,
			Time:     bomb.Time,
		}
	}

	return observation
}

func (o Observation) Vector(scale ObservationScale) ObservationVector {
	var vector ObservationVector
	index := 0

	for _, entity := range o.Entities {
		if entity.Exists {
			vector[index] = 1
			vector[index+1] = normalizeSigned(entity.Position.X, scale.Distance)
			vector[index+2] = normalizeSigned(entity.Position.Y, scale.Distance)
			vector[index+3] = truth(entity.Friendly)
		}
		index += EntityVectorSize
	}

	for _, projectile := range o.Projectiles {
		if projectile.Exists {
			direction := projectile.Direction.unit()
			vector[index] = 1
			vector[index+1] = normalizeSigned(projectile.Position.X, scale.Distance)
			vector[index+2] = normalizeSigned(projectile.Position.Y, scale.Distance)
			vector[index+3] = direction.X
			vector[index+4] = direction.Y
			vector[index+5] = normalize(projectile.Damage, scale.Damage)
		}
		index += ProjectileVectorSize
	}

	for _, bomb := range o.Bombs {
		if bomb.Exists {
			vector[index] = 1
			vector[index+1] = normalizeSigned(bomb.Position.X, scale.Distance)
			vector[index+2] = normalizeSigned(bomb.Position.Y, scale.Distance)
			vector[index+3] = normalize(bomb.Damage, scale.Damage)
			vector[index+4] = normalize(bomb.Time, scale.Time)
		}
		index += BombVectorSize
	}

	return vector
}

type candidate[T any] struct {
	id       uint32
	distance float32
	value    T
}

func nearest[T any](
	values []T,
	origin Vector,
	include func(T) bool,
	id func(T) uint32,
	position func(T) Vector,
) []T {
	candidates := make([]candidate[T], 0, len(values))
	for _, value := range values {
		if !include(value) {
			continue
		}
		candidates = append(candidates, candidate[T]{
			id:       id(value),
			distance: position(value).distanceSquared(origin),
			value:    value,
		})
	}

	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].distance == candidates[j].distance {
			return candidates[i].id < candidates[j].id
		}
		return candidates[i].distance < candidates[j].distance
	})

	count := min(len(candidates), NearbyLimit)
	result := make([]T, count)
	for index := range count {
		result[index] = candidates[index].value
	}
	return result
}
