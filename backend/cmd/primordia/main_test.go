package main

import "testing"

// TestRatesArePositive guards against accidental non-positive runtime intervals.
func TestRatesArePositive(t *testing.T) {
	if TickRate <= 0 || BroadcastRate <= 0 {
		t.Fatalf("tick and broadcast rates must be positive")
	}
}
