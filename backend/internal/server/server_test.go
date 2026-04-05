package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/martin/primordia/internal/world"
)

// TestNewServerDefaults ensures constructor stores world reference and broadcast rate.
func TestNewServerDefaults(t *testing.T) {
	s := New(world.New(), 40*time.Millisecond)
	if s.World == nil {
		t.Fatalf("expected world to be set")
	}
	if s.BroadcastRate <= 0 {
		t.Fatalf("expected positive broadcast rate")
	}
}

// TestSpeedHandlerStub verifies the speed control stub accepts valid POST requests.
func TestSpeedHandlerStub(t *testing.T) {
	s := New(world.New(), 40*time.Millisecond)
	called := 0
	s.SetSpeedController(1, func(rate float64) {
		called++
		if rate != 2 {
			t.Fatalf("expected rate 2, got %f", rate)
		}
	})

	req := httptest.NewRequest(http.MethodPost, "/speed?rate=2", nil)
	rr := httptest.NewRecorder()
	s.SpeedHandler(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rr.Code)
	}
	if called != 1 {
		t.Fatalf("expected speed callback to be called once, got %d", called)
	}
	if s.Speed() != 2 {
		t.Fatalf("expected server speed 2, got %f", s.Speed())
	}
}

// TestControlHandlerDispatch verifies valid control actions are forwarded to runtime callback.
func TestControlHandlerDispatch(t *testing.T) {
	s := New(world.New(), 40*time.Millisecond)
	seen := ""
	s.SetControlHandler(func(action string) bool {
		seen = action
		return true
	})

	req := httptest.NewRequest(http.MethodPost, "/control?action=stop", nil)
	rr := httptest.NewRecorder()
	s.ControlHandler(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rr.Code)
	}
	if seen != "stop" {
		t.Fatalf("expected action stop, got %q", seen)
	}
}

// TestControlHandlerRejectsInvalidAction ensures malformed action values are rejected.
func TestControlHandlerRejectsInvalidAction(t *testing.T) {
	s := New(world.New(), 40*time.Millisecond)
	s.SetControlHandler(func(action string) bool {
		return true
	})

	req := httptest.NewRequest(http.MethodPost, "/control?action=unknown", nil)
	rr := httptest.NewRecorder()
	s.ControlHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}
