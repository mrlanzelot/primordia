package systems

import (
	"math"
	"testing"

	"github.com/martin/primordia/internal/food"
	"github.com/martin/primordia/internal/organism"
	"github.com/martin/primordia/internal/spatial"
)

func TestSenseDetectsFoodInFront(t *testing.T) {
	org := &organism.Organism{
		ID:     1,
		Pos:    organism.Vec2{X: 100, Y: 100},
		Angle:  0,
		Vel:    organism.Vec2{X: 1, Y: 0},
		Energy: 80,
	}
	f := &food.Food{ID: 10, Pos: organism.Vec2{X: 128.6, Y: 108.8}}

	orgs := map[uint64]*organism.Organism{1: org}
	foods := map[uint64]*food.Food{10: f}
	grid := spatial.New(20)
	grid.Insert(entityIDForOrganism(org.ID), org.Pos)
	grid.Insert(entityIDForFood(f.ID), f.Pos)

	UpdateSense(orgs, foods, grid)
	if len(org.SenseVec) != 21 {
		t.Fatalf("expected 21 sense values, got %d", len(org.SenseVec))
	}

	foundFood := false
	for i := 0; i < 16; i += 2 {
		if org.SenseVec[i+1] == 1 {
			foundFood = true
			if org.SenseVec[i] >= 1 {
				t.Fatalf("expected normalized hit distance < 1, got %f", org.SenseVec[i])
			}
		}
	}
	if !foundFood {
		t.Fatalf("expected at least one ray to detect food")
	}
	if org.SenseVec[16] <= 0 || math.Abs(org.SenseVec[17]) < 0.1 {
		t.Fatalf("expected smell vector to point toward food, got %+v", org.SenseVec[16:19])
	}
}
