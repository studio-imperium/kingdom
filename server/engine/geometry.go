package engine

import (
	"math"
	"math/rand/v2"
)

func distance(a, b Position) float32 {
	return float32(math.Hypot(float64(a.X-b.X), float64(a.Y-b.Y)))
}

func angle(a, b Position) float32 {
	value := float32(math.Atan2(float64(b.Y-a.Y), float64(b.X-a.X)) * 180 / math.Pi)
	if value < 0 {
		value += 360
	}
	return value
}

func nearby(position Position, radius float32) Position {
	return Position{
		X: position.X + rand.Float32()*radius*2 - radius,
		Y: position.Y + rand.Float32()*radius*2 - radius,
	}
}

func direction(value float32) float32 {
	if value >= 0 {
		return 1
	}
	return -1
}

func multiplier(value float32) float32 {
	if value == 0 {
		return 1
	}
	return value
}

func finite(values ...float32) bool {
	for _, value := range values {
		if math.IsNaN(float64(value)) || math.IsInf(float64(value), 0) {
			return false
		}
	}
	return true
}
