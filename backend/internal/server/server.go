package server

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
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
}

func New(w *world.World, rate time.Duration) *Server {
	return &Server{
		World:         w,
		BroadcastRate: rate,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
		clients: make(map[*websocket.Conn]struct{}),
	}
}

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

func (s *Server) addClient(conn *websocket.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.clients[conn] = struct{}{}
}

func (s *Server) removeClient(conn *websocket.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.clients, conn)
}

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
