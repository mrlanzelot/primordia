package food

import "testing"

// TestSpawnRandomWithinBounds validates random food spawn respects map bounds.
func TestSpawnRandomWithinBounds(t *testing.T) {
	f := SpawnRandom(7, 100, 200)
	if f.ID != 7 {
		t.Fatalf("expected id 7, got %d", f.ID)
	}
	if f.Pos.X < 0 || f.Pos.X > 100 || f.Pos.Y < 0 || f.Pos.Y > 200 {
		t.Fatalf("spawned out of bounds: %+v", f.Pos)
	}
}
