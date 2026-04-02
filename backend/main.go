package main

import (
	"context"
	"encoding/json"
	"log"
	"math"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
)

const (
	WorldWidth        = 1000
	WorldHeight       = 1000
	TickRate          = 30 * time.Millisecond
	BroadcastRate     = 40 * time.Millisecond
	InitialPopulation = 100
	MaxFoodCount      = 200
	EnergyCost        = 0.1
	EnergyFromFood    = 35.0
	EatDistanceSq     = 144.0
	InitialEnergy     = 100.0

	SearchSpeed        = 1.5
	SenseRadiusSq      = 500.0
	CircleRadiusJitter = 0.22
)

type Position struct {
	X, Y float64 `json:"x"`
}
type CellState string

const (
	CellStateSearch        CellState = "search"
	CellStateExploitCircle CellState = "exploit_circle"
	CellStateReorient      CellState = "reorient"
)

type Organism struct {
	ID      uint32    `json:"id"`
	Pos     Position  `json:"pos"`
	Energy  float64   `json:"energy"`
	State   CellState `json:"state"`
	VelX    float64   `json:"velX"`
	VelY    float64   `json:"velY"`
	Turn    float64   `json:"turn"`
	Wobble  float64   `json:"wobble"`
	Radius  float64   `json:"radius"`
	Timer   int       `json:"timer"`
	Loops   int       `json:"loops"`
	GoalX   float64   `json:"goalX"`
	GoalY   float64   `json:"goalY"`
	HasGoal bool      `json:"hasGoal"`
}

type Food struct {
	ID  uint32   `json:"id"`
	Pos Position `json:"pos"`
}

type World struct {
	Organisms map[uint32]*Organism `json:"orgs"`
	Food      map[uint32]*Food     `json:"food"`
	Mu        sync.RWMutex         `json:"-"`
	NextID    uint32               `json:"-"`
}

func normalize(x, y float64) (float64, float64) {
	l := math.Hypot(x, y)
	if l == 0 {
		return 1, 0
	}
	return x / l, y / l
}
func rotate(x, y, a float64) (float64, float64) {
	c, s := math.Cos(a), math.Sin(a)
	return x*c - y*s, x*s + y*c
}
func clamp(v, lo, hi float64) float64 { return math.Max(lo, math.Min(hi, v)) }
func pickCircleRadius(base float64) float64 {
	r := 10 + rand.Float64()*30
	if rand.Float64() < 0.35 {
		r = base * (1 + (rand.Float64()*2-1)*CircleRadiusJitter)
	}
	return clamp(r, 8, 64)
}

func newOrganism(id uint32) *Organism {
	ang := rand.Float64() * 2 * math.Pi
	x, y := math.Cos(ang), math.Sin(ang)
	return &Organism{ID: id, Energy: InitialEnergy, State: CellStateSearch, VelX: x * SearchSpeed, VelY: y * SearchSpeed, Turn: 0.06 + rand.Float64()*0.10, Wobble: 0.03 + rand.Float64()*0.08, Radius: 18 + rand.Float64()*28, Timer: 12 + rand.Intn(20)}
}

func lerp(a, b, t float64) float64 { return a + (b-a)*t }
func speed(x, y float64) float64   { return math.Hypot(x, y) }
func steerTowards(org *Organism, tx, ty float64, strength float64) {
	dx, dy := normalize(tx, ty)
	s := speed(org.VelX, org.VelY)
	if s == 0 {
		s = SearchSpeed
	}
	org.VelX = lerp(org.VelX, dx*s, strength)
	org.VelY = lerp(org.VelY, dy*s, strength)
}

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
		if org.Radius == 0 {
			org.Radius = 18 + rand.Float64()*28
		}
		switch org.State {
		case CellStateSearch:
			org.Timer--
			org.VelX, org.VelY = rotate(org.VelX, org.VelY, (rand.Float64()-0.5)*org.Wobble)
			if fx, fy, ok := w.nearestFood(org.Pos.X, org.Pos.Y, SenseRadiusSq*rand.Float64()); ok {
				org.State = CellStateExploitCircle
				org.Loops = 0
				org.Timer = 0
				org.Radius = pickCircleRadius(org.Radius)
				org.GoalX, org.GoalY = fx, fy
				org.HasGoal = true
			}
			if org.Timer <= 0 {
				org.State = CellStateReorient
			}
			org.VelX, org.VelY = normalize(org.VelX, org.VelY)
			org.VelX *= SearchSpeed
			org.VelY *= SearchSpeed
		case CellStateExploitCircle:
			org.Timer++
			cx, cy := org.GoalX, org.GoalY
			dx := org.Pos.X - cx
			dy := org.Pos.Y - cy
			org.Radius += (rand.Float64() - 0.5) * 0.8
			org.Radius = clamp(org.Radius, 8, 64)
			targetAngle := org.Turn + (rand.Float64()-0.5)*0.10
			px, py := rotate(dx, dy, targetAngle)
			px, py = normalize(px, py)
			desiredX := px * org.Radius * 0.10
			desiredY := py * org.Radius * 0.10
			steerTowards(org, desiredX, desiredY, 0.12+org.Wobble*0.5)
			if org.Timer%8 == 0 {
				org.Loops++
			}
			if fx, fy, ok := w.nearestFood(org.Pos.X, org.Pos.Y, SenseRadiusSq*0.75+rand.Float64()*SenseRadiusSq*0.75); ok {
				org.GoalX, org.GoalY = fx, fy
				org.HasGoal = true
				org.Radius = pickCircleRadius(org.Radius)
			} else if org.Loops >= 2 {
				org.State = CellStateReorient
			}
		case CellStateReorient:
			if !org.HasGoal {
				org.GoalX = rand.Float64() * WorldWidth
				org.GoalY = rand.Float64() * WorldHeight
			}
			dx := org.GoalX - org.Pos.X
			dy := org.GoalY - org.Pos.Y
			steerTowards(org, dx, dy, 0.05+org.Wobble*0.35)
			org.Timer -= 2
			if org.Timer <= 0 {
				org.Timer = 12 + rand.Intn(20)
				org.State = CellStateSearch
				org.Radius = pickCircleRadius(org.Radius)
			}
		}
		org.VelX, org.VelY = normalize(org.VelX, org.VelY)
		org.Pos.X += org.VelX * speed(org.VelX, org.VelY)
		org.Pos.Y += org.VelY * speed(org.VelX, org.VelY)
		org.Pos.X = clamp(org.Pos.X, 0, WorldWidth)
		org.Pos.Y = clamp(org.Pos.Y, 0, WorldHeight)
		org.Energy -= EnergyCost
		if org.Energy <= 0 {
			deadIDs = append(deadIDs, id)
		}
	}
	for _, id := range deadIDs {
		delete(w.Organisms, id)
	}
}

func (w *World) nearestFood(x, y float64, maxSq float64) (float64, float64, bool) {
	best := maxSq
	var bx, by float64
	found := false
	for _, f := range w.Food {
		dx := x - f.Pos.X
		dy := y - f.Pos.Y
		d := dx*dx + dy*dy
		if d <= best {
			best = d
			bx, by = f.Pos.X, f.Pos.Y
			found = true
		}
	}
	return bx, by, found
}
func (w *World) spawnFood() {
	if len(w.Food) >= MaxFoodCount {
		return
	}
	w.NextID++
	w.Food[w.NextID] = &Food{ID: w.NextID, Pos: Position{X: rand.Float64() * WorldWidth, Y: rand.Float64() * WorldHeight}}
}
func (w *World) applyEating() {
	eaten := make(map[uint32]struct{})
	for _, org := range w.Organisms {
		for fid, f := range w.Food {
			if _, ok := eaten[fid]; ok {
				continue
			}
			dx := org.Pos.X - f.Pos.X
			dy := org.Pos.Y - f.Pos.Y
			if dx*dx+dy*dy < EatDistanceSq {
				org.Energy += EnergyFromFood
				org.HasGoal = true
				org.GoalX, org.GoalY = f.Pos.X, f.Pos.Y
				org.State = CellStateExploitCircle
				org.Timer = 0
				org.Loops = 0
				org.Radius = pickCircleRadius(org.Radius)
				eaten[fid] = struct{}{}
			}
		}
	}
	for fid := range eaten {
		delete(w.Food, fid)
	}
}

var upgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}

func main() {
	rand.Seed(time.Now().UnixNano())
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	world := &World{Organisms: make(map[uint32]*Organism), Food: make(map[uint32]*Food)}
	for i := 0; i < InitialPopulation; i++ {
		world.NextID++
		world.Organisms[world.NextID] = newOrganism(world.NextID)
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
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("websocket upgrade error: %v", err)
			return
		}
		defer conn.Close()
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
					return
				}
				if err := conn.WriteMessage(websocket.TextMessage, payload); err != nil {
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
		_ = server.Shutdown(shutdownCtx)
	}()
	log.Println("Primordia engine running on :8080")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Printf("critical server failure: %v", err)
	}
}
