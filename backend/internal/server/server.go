package server

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/martin/primordia/internal/protocol"
	"github.com/martin/primordia/internal/world"
)

type Server struct {
	World         *world.World
	BroadcastRate time.Duration

	upgrader websocket.Upgrader
	mu       sync.Mutex
	clients  map[*websocket.Conn]struct{}
	speedMu  sync.RWMutex
	speed    float64
	onSpeed  func(float64)
	ctrlMu   sync.RWMutex
	onAction func(string) bool
}

// ControlHandler applies start/stop/restart control actions from the HUD.
func (s *Server) ControlHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	action := r.URL.Query().Get("action")
	if action != "start" && action != "stop" && action != "restart" {
		http.Error(w, "invalid action", http.StatusBadRequest)
		return
	}

	s.ctrlMu.RLock()
	onAction := s.onAction
	s.ctrlMu.RUnlock()
	if onAction == nil {
		http.Error(w, "control handler unavailable", http.StatusServiceUnavailable)
		return
	}

	if !onAction(action) {
		http.Error(w, "control action failed", http.StatusInternalServerError)
		return
	}

	log.Printf("simulation control action=%s", action)
	w.WriteHeader(http.StatusNoContent)
}

// SpeedHandler validates speed changes from the HUD and applies them to the running engine.
func (s *Server) SpeedHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rateStr := r.URL.Query().Get("rate")
	if rateStr == "" {
		http.Error(w, "missing rate query parameter", http.StatusBadRequest)
		return
	}

	rate, err := strconv.ParseFloat(rateStr, 64)
	if err != nil || rate <= 0 {
		http.Error(w, "invalid rate", http.StatusBadRequest)
		return
	}

	s.speedMu.Lock()
	s.speed = rate
	onSpeed := s.onSpeed
	s.speedMu.Unlock()

	if onSpeed != nil {
		onSpeed(rate)
	}

	log.Printf("speed control updated rate=%0.2f", rate)
	w.WriteHeader(http.StatusNoContent)
}

// New constructs a websocket server bound to a world instance and broadcast interval.
func New(w *world.World, rate time.Duration) *Server {
	return &Server{
		World:         w,
		BroadcastRate: rate,
		speed:         1,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
		clients: make(map[*websocket.Conn]struct{}),
	}
}

// SetSpeedController registers a callback that applies speed changes to the simulation loop.
func (s *Server) SetSpeedController(initial float64, onSpeed func(float64)) {
	s.speedMu.Lock()
	defer s.speedMu.Unlock()
	if initial > 0 {
		s.speed = initial
	}
	s.onSpeed = onSpeed
}

// SetControlHandler registers control action callback for start/stop/restart operations.
func (s *Server) SetControlHandler(onAction func(string) bool) {
	s.ctrlMu.Lock()
	defer s.ctrlMu.Unlock()
	s.onAction = onAction
}

// Speed returns the last speed multiplier accepted by the server.
func (s *Server) Speed() float64 {
	s.speedMu.RLock()
	defer s.speedMu.RUnlock()
	return s.speed
}

// WSHandler upgrades HTTP clients and registers them in the broadcast hub.
func (s *Server) WSHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("websocket upgrade error: %v", err)
		return
	}
	s.addClient(conn)

	go func() {
		defer s.removeClient(conn)
		defer conn.Close()
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				return
			}
		}
	}()
}

// BroadcastLoop emits serialized world snapshots to all active websocket clients.
func (s *Server) BroadcastLoop(ctx context.Context) {
	ticker := time.NewTicker(s.BroadcastRate)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			msg := protocol.Snapshot(s.World)
			payload, err := json.Marshal(msg)
			if err != nil {
				continue
			}
			s.broadcast(payload)
		}
	}
}

// addClient tracks a newly connected websocket client.
func (s *Server) addClient(conn *websocket.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.clients[conn] = struct{}{}
}

// removeClient evicts a disconnected websocket client from the hub.
func (s *Server) removeClient(conn *websocket.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.clients, conn)
}

// broadcast writes one payload to every client and prunes broken connections.
func (s *Server) broadcast(payload []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for conn := range s.clients {
		if err := conn.WriteMessage(websocket.TextMessage, payload); err != nil {
			_ = conn.Close()
			delete(s.clients, conn)
		}
	}
}
