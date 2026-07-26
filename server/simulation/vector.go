package simulation

import "math"

func (v Vector) relativeTo(origin Vector) Vector {
	return Vector{X: v.X - origin.X, Y: v.Y - origin.Y}
}

func (v Vector) distanceSquared(other Vector) float32 {
	x := v.X - other.X
	y := v.Y - other.Y
	return x*x + y*y
}

func (v Vector) unit() Vector {
	length := float32(math.Hypot(float64(v.X), float64(v.Y)))
	if length == 0 || !finite(length) {
		return Vector{}
	}
	return Vector{X: v.X / length, Y: v.Y / length}
}

func (v Vector) limit(maximum float32) Vector {
	length := float32(math.Hypot(float64(v.X), float64(v.Y)))
	if length <= maximum {
		return v
	}
	return v.scale(maximum / length)
}

func finite(value float32) bool {
	return !math.IsNaN(float64(value)) && !math.IsInf(float64(value), 0)
}

func clamp(value, minimum, maximum float32) float32 {
	if !finite(value) {
		return minimum
	}
	return min(max(value, minimum), maximum)
}

func normalize(value, scale float32) float32 {
	if scale <= 0 || !finite(scale) {
		return 0
	}
	return clamp(value/scale, 0, 1)
}

func normalizeSigned(value, scale float32) float32 {
	if scale <= 0 || !finite(scale) {
		return 0
	}
	return clamp(value/scale, -1, 1)
}

func truth(value bool) float32 {
	if value {
		return 1
	}
	return 0
}
