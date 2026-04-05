package organism

import "math"

// Vec2 is a simple 2D vector used across simulation systems.
type Vec2 struct {
	X float64
	Y float64
}

func (v Vec2) Add(o Vec2) Vec2 {
	return Vec2{X: v.X + o.X, Y: v.Y + o.Y}
}

func (v Vec2) Sub(o Vec2) Vec2 {
	return Vec2{X: v.X - o.X, Y: v.Y - o.Y}
}

func (v Vec2) Scale(s float64) Vec2 {
	return Vec2{X: v.X * s, Y: v.Y * s}
}

func (v Vec2) Length() float64 {
	return math.Hypot(v.X, v.Y)
}

func (v Vec2) Normalize() Vec2 {
	l := v.Length()
	if l == 0 {
		return Vec2{X: 1, Y: 0}
	}
	return Vec2{X: v.X / l, Y: v.Y / l}
}

func Rotate(v Vec2, a float64) Vec2 {
	c, s := math.Cos(a), math.Sin(a)
	return Vec2{X: v.X*c - v.Y*s, Y: v.X*s + v.Y*c}
}

// Organism includes current behavior state plus Phase 2 placeholders.
type Organism struct {
	ID       uint64
	Pos      Vec2
	Vel      Vec2
	Angle    float64
	Energy   float64
	Age      int
	Genome   []float64
	SenseVec []float64

	State   string
	DirX    float64
	DirY    float64
	Timer   int
	Loops   int
	FoodX   float64
	FoodY   float64
	HasFood bool
	TargetX float64
	TargetY float64
	PlanFor int
}
