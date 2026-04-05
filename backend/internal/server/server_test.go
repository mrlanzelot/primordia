package server

import (
	"testing"
	"time"

	"github.com/martin/primordia/internal/world"
)

func TestNewServerDefaults(t *testing.T) {
	s := New(world.New(), 40*time.Millisecond)
	if s.World == nil {
		t.Fatalf("expected world to be set")
	}
	if s.BroadcastRate <= 0 {
		t.Fatalf("expected positive broadcast rate")
	}
}
