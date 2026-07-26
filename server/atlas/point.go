package atlas

import (
	"math"
)

type Point struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

func newPoint(x int, y int) Point {
	return Point{float64(x), float64(y)}
}

func (p1 Point) add(p2 Point) Point {
	return Point{
		X: p1.X + p2.X,
		Y: p1.Y + p2.Y,
	}
}

func (p1 Point) subtract(p2 Point) Point {
	return Point{
		X: p1.X - p2.X,
		Y: p1.Y - p2.Y,
	}
}

func (p1 Point) multiply(p2 Point) Point {
	return Point{
		X: p1.X * p2.X,
		Y: p1.Y * p2.Y,
	}
}

func (pt Point) sqr() Point {
	return Point{
		X: math.Pow(pt.X, 2),
		Y: math.Pow(pt.Y, 2),
	}
}

func (p1 Point) divide(p2 Point) Point {
	return Point{
		X: p1.X / p2.X,
		Y: p1.Y / p2.Y,
	}
}

func distance(p1 Point, p2 Point) float64 {
	return math.Hypot(p1.X-p2.X, p1.Y-p2.Y)
}
