package organism

import (
	"math"
	"testing"
)

// TestVec2NormalizeZero ensures zero vectors normalize to deterministic unit-x fallback.
func TestVec2NormalizeZero(t *testing.T) {
	v := (Vec2{}).Normalize()
	if math.Abs(v.X-1) > 1e-9 || math.Abs(v.Y) > 1e-9 {
		t.Fatalf("expected unit x for zero normalize, got %+v", v)
	}
}
