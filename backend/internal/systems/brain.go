package systems

import "math/rand"

// ActionVec represents the output of the brain.
type ActionVec struct {
	TurnDelta float64
	Thrust    float64
	EatFlag   float64
}

// Think takes a sense vector and returns an action vector.
// TODO: replace with neural network forward pass in Phase 2.
func Think(sv []float64) ActionVec {
	_ = sv
	return ActionVec{
		TurnDelta: (rand.Float64()*2 - 1) * 0.1,
		Thrust:    0.5 + rand.Float64()*0.5,
	}
}
