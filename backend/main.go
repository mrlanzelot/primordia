package main

import (
	"context"
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
)

// --- CONFIGURATION ---
const (
	WorldWidth        = 1000
	WorldHeight       = 1000
	TickRate          = 30 * time.Millisecond // ~33 FPS simulation
	BroadcastRate     = 40 * time.Millisecond // 25 FPS UI updates
	InitialPopulation = 50
	MaxFoodCount      = 100
	BrownianStep      = 8.0
	EnergyCost        = 0.1
	EnergyFromFood    = 30.0
	EatDistanceSq     = 144.0
	InitialEnergy     = 100.0
)

// --- ECS COMPONENTS ---

type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type Organism struct {
	ID     uint32   `json:"id"`
	Pos    Position `json:"pos"`
	Energy float64  `json:"energy"`
}

type Food struct {
	ID  uint32   `json:"id"`
	Pos Position `json:"pos"`
}

// --- WORLD STATE ---

type World struct {
	Organisms map[uint32]*Organism `json:"orgs"`
	Food      map[uint32]*Food     `json:"food"`
	Mu        sync.RWMutex         `json:"-"` // Hide from JSON
	NextID    uint32               `json:"-"`
}

// --- SYSTEMS (The Logic) ---

func (w *World) Update() {
	w.Mu.Lock()
	defer w.Mu.Unlock()

	w.updateOrganisms()
	w.spawnFood()
	w.applyEating()
}

func (w *World) updateOrganisms() {
	deadIDs := make([]uint32, 0)

	for id, org := range w.Organisms {
		org.Pos.X += (rand.Float64() - 0.5) * BrownianStep
		org.Pos.Y += (rand.Float64() - 0.5) * BrownianStep
		org.Energy -= EnergyCost

		if org.Pos.X < 0 {
			org.Pos.X = 0
		}
		if org.Pos.X > WorldWidth {
			org.Pos.X = WorldWidth
		}
		if org.Pos.Y < 0 {
			org.Pos.Y = 0
		}
		if org.Pos.Y > WorldHeight {
			org.Pos.Y = WorldHeight
		}

		if org.Energy <= 0 {
			deadIDs = append(deadIDs, id)
		}
	}

	for _, id := range deadIDs {
		delete(w.Organisms, id)
	}
}

func (w *World) spawnFood() {
	if len(w.Food) >= MaxFoodCount {
		return
	}

	w.NextID++
	w.Food[w.NextID] = &Food{
		ID:  w.NextID,
		Pos: Position{X: rand.Float64() * WorldWidth, Y: rand.Float64() * WorldHeight},
	}
}

func (w *World) applyEating() {
	eatenFood := make(map[uint32]struct{})

	for _, org := range w.Organisms {
		for fid, f := range w.Food {
			if _, alreadyEaten := eatenFood[fid]; alreadyEaten {
				continue
			}

			dx := org.Pos.X - f.Pos.X
			dy := org.Pos.Y - f.Pos.Y
			if (dx*dx + dy*dy) < EatDistanceSq {
				org.Energy += EnergyFromFood
				eatenFood[fid] = struct{}{}
			}
		}
	}

	for fid := range eatenFood {
		delete(w.Food, fid)
	}
}

// --- NETWORKING ---

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true }, // Allow all for local dev
}

func main() {
	rand.Seed(time.Now().UnixNano())

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	world := &World{
		Organisms: make(map[uint32]*Organism),
		Food:      make(map[uint32]*Food),
		NextID:    0,
	}

	// Initial Population (The Primordial Ooze)
	for i := 0; i < InitialPopulation; i++ {
		world.NextID++
		world.Organisms[world.NextID] = &Organism{
			ID:     world.NextID,
			Pos:    Position{X: rand.Float64() * WorldWidth, Y: rand.Float64() * WorldHeight},
			Energy: InitialEnergy,
		}
	}

	engineDone := make(chan struct{})
	go func() {
		ticker := time.NewTicker(TickRate)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				world.Update()
			case <-engineDone:
				return
			}
		}
	}()

	// WebSocket Handler
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("websocket upgrade error: %v", err)
			return
		}
		defer conn.Close()

		log.Printf("client connected: %s", r.RemoteAddr)

		// Broadcast loop for this specific connection
		ticker := time.NewTicker(BroadcastRate)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				world.Mu.RLock()
				payload, err := json.Marshal(world)
				world.Mu.RUnlock()

				if err != nil {
					log.Printf("world marshal error: %v", err)
					return
				}

				if err := conn.WriteMessage(websocket.TextMessage, payload); err != nil {
					log.Printf("client disconnected: %v", err)
					return
				}
			}
		}
	})

	server := &http.Server{Addr: ":8080"}

	go func() {
		<-ctx.Done()
		close(engineDone)

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("server shutdown error: %v", err)
		}
	}()

	log.Println("Primordia engine running on :8080")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Printf("critical server failure: %v", err)
	}
}
