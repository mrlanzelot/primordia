package main

import "testing"

func TestRatesArePositive(t *testing.T) {
	if TickRate <= 0 || BroadcastRate <= 0 {
		t.Fatalf("tick and broadcast rates must be positive")
	}
}
