package world

import "testing"

func TestNewSeedsPopulation(t *testing.T) {
	w := New()
	if len(w.Organisms) != InitialPopulation {
		t.Fatalf("expected %d organisms, got %d", InitialPopulation, len(w.Organisms))
	}
	if w.Grid == nil {
		t.Fatalf("expected non-nil grid")
	}
}
