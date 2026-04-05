package systems

import "testing"

func TestThinkOutputRange(t *testing.T) {
	out := Think(make([]float64, 21))
	if out.Thrust < 0.5 || out.Thrust > 1.0 {
		t.Fatalf("thrust out of range: %f", out.Thrust)
	}
	if out.TurnDelta < -0.1 || out.TurnDelta > 0.1 {
		t.Fatalf("turn delta out of range: %f", out.TurnDelta)
	}
}
