package protocol

import (
	"testing"

	"github.com/martin/primordia/internal/food"
	"github.com/martin/primordia/internal/organism"
	"github.com/martin/primordia/internal/world"
)

// TestSnapshotBuildsWorldMsg checks snapshot serialization keeps expected counts and tick.
func TestSnapshotBuildsWorldMsg(t *testing.T) {
	w := &world.World{
		Organisms: map[uint64]*organism.Organism{
			1: {ID: 1, Pos: organism.Vec2{X: 12, Y: 34}, Energy: 55, Angle: 1.2, Age: 123, SenseVec: []float64{0.2, 1}},
		},
		Foods: map[uint64]*food.Food{
			2: {ID: 2, Pos: organism.Vec2{X: 80, Y: 90}},
		},
		TickID: 42,
	}

	msg := Snapshot(w)
	if msg.Tick != 42 {
		t.Fatalf("expected tick 42, got %d", msg.Tick)
	}
	if len(msg.Organisms) != 1 || len(msg.Foods) != 1 {
		t.Fatalf("unexpected counts: org=%d food=%d", len(msg.Organisms), len(msg.Foods))
	}
	if msg.Organisms[0].Energy <= 0 || msg.Organisms[0].Energy >= 1 {
		t.Fatalf("expected normalized energy in (0,1), got %f", msg.Organisms[0].Energy)
	}
	if msg.Organisms[0].Age != 123 {
		t.Fatalf("expected organism age 123, got %d", msg.Organisms[0].Age)
	}
}
