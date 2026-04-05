package spatial

import (
	"testing"

	"github.com/martin/primordia/internal/organism"
)

// TestInsertAndQueryRadius verifies grid bucketing finds nearby ids and excludes far ones.
func TestInsertAndQueryRadius(t *testing.T) {
	g := New(10)
	g.Insert(1, organism.Vec2{X: 5, Y: 5})
	g.Insert(2, organism.Vec2{X: 99, Y: 99})

	ids := g.QueryRadius(organism.Vec2{X: 6, Y: 6}, 8)
	seen := map[uint64]bool{}
	for _, id := range ids {
		seen[id] = true
	}
	if !seen[1] {
		t.Fatalf("expected id 1 in query result")
	}
	if seen[2] {
		t.Fatalf("did not expect id 2 in nearby query result")
	}
}
