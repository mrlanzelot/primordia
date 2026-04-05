package systems

import (
	"math"
	"testing"

	"github.com/martin/primordia/internal/organism"
)

// TestNearestOrganismSkipsSelf verifies nearest lookup ignores the current organism id.
func TestNearestOrganismSkipsSelf(t *testing.T) {
	orgs := map[uint64]*organism.Organism{
		1: {ID: 1, Pos: organism.Vec2{X: 100, Y: 100}},
		2: {ID: 2, Pos: organism.Vec2{X: 110, Y: 100}},
	}

	x, y, distSq, ok := nearestOrganism(1, 100, 100, OrganismAvoidRadiusSq, orgs)
	if !ok {
		t.Fatalf("expected nearby organism, got none")
	}
	if x != 110 || y != 100 {
		t.Fatalf("unexpected nearest organism position: (%f,%f)", x, y)
	}
	if distSq >= OrganismAvoidRadiusSq {
		t.Fatalf("expected nearest distance within avoid radius, got %f", distSq)
	}
}

// TestCrowdingAvoidanceTurnsAway confirms search steering bends away from close organisms.
func TestCrowdingAvoidanceTurnsAway(t *testing.T) {
	org := &organism.Organism{
		ID:      1,
		Pos:     organism.Vec2{X: 200, Y: 200},
		Energy:  InitialEnergy,
		State:   CellStateSearch,
		DirX:    1,
		DirY:    0,
		TargetX: 1,
		TargetY: 0,
		Timer:   20,
		PlanFor: 20,
	}
	other := &organism.Organism{ID: 2, Pos: organism.Vec2{X: 208, Y: 200}, Energy: InitialEnergy}
	orgs := map[uint64]*organism.Organism{1: org, 2: other}

	UpdateOrganisms(orgs, nil)

	if math.Abs(org.DirY) < 0.05 {
		t.Fatalf("expected steering response from crowding, dirY stayed too small: %f", org.DirY)
	}
	if math.Abs(org.Pos.X-200) > SearchSpeed {
		t.Fatalf("unexpected search displacement: %f", org.Pos.X-200)
	}
}
