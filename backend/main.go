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
	MaxFoodCount      = 135
	EnergyCost        = 0.2
	EnergyFromFood    = 35.0
	EatDistanceSq     = 144.0
	InitialEnergy     = 80.0

	SearchWobble  = 0.10
	SearchSpeed   = 1.0
	CircleTurn    = 0.22
	CircleSpeed   = 1.0
	SearchTicks   = 24
	CircleTicks   = 10
	SenseRadiusSq = 500.0

	SearchTurnMax   = 0.09
	PlanTicksMin    = 18
	PlanTicksMax    = 60
	WallAvoidMargin = 50.0
	FoodSeekWeight  = 0.28
	FoodSenseSq     = 6400.0
)

type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
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
	DirX    float64   `json:"dirX"`
	DirY    float64   `json:"dirY"`
	Timer   int       `json:"timer"`
	Loops   int       `json:"loops"`
	FoodX   float64   `json:"foodX"`
	FoodY   float64   `json:"foodY"`
	HasFood bool      `json:"hasFood"`
	TargetX float64   `json:"targetX"`
	TargetY float64   `json:"targetY"`
	PlanFor int       `json:"planFor"`
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

func wrapAngle(a float64) float64 {
	for a > math.Pi {
		a -= 2 * math.Pi
	}
	for a < -math.Pi {
		a += 2 * math.Pi
	}
	return a
}

func steerToward(curX, curY, targetX, targetY, maxTurn float64) (float64, float64) {
	curX, curY = normalize(curX, curY)
	targetX, targetY = normalize(targetX, targetY)
	curA := math.Atan2(curY, curX)
	targetA := math.Atan2(targetY, targetX)
	delta := wrapAngle(targetA - curA)
	if delta > maxTurn {
		delta = maxTurn
	} else if delta < -maxTurn {
		delta = -maxTurn
	}
	return rotate(curX, curY, delta)
}

func randPlanTicks() int {
	return PlanTicksMin + rand.Intn(PlanTicksMax-PlanTicksMin+1)
}

func shouldAvoidWall(p Position) bool {
	return p.X < WallAvoidMargin || p.X > WorldWidth-WallAvoidMargin || p.Y < WallAvoidMargin || p.Y > WorldHeight-WallAvoidMargin
}

func inwardTarget(p Position) (float64, float64) {
	return normalize((WorldWidth/2)-p.X, (WorldHeight/2)-p.Y)
}

func randomWanderTarget(dirX, dirY float64) (float64, float64) {
	turn := (rand.Float64()-0.5)*0.8 + (rand.Float64()-0.5)*SearchWobble
	return normalize(rotate(dirX, dirY, turn))
}

func blendDirection(baseX, baseY, addX, addY, weight float64) (float64, float64) {
	bx, by := normalize(baseX, baseY)
	ax, ay := normalize(addX, addY)
	return normalize(bx*(1-weight)+ax*weight, by*(1-weight)+ay*weight)
}

func newOrganism(id uint32) *Organism {
	ang := rand.Float64() * 2 * math.Pi
	x, y := math.Cos(ang), math.Sin(ang)
	tx, ty := randomWanderTarget(x, y)
	return &Organism{
		ID:      id,
		Pos:     Position{X: rand.Float64() * WorldWidth, Y: rand.Float64() * WorldHeight},
		Energy:  InitialEnergy,
		State:   CellStateSearch,
		DirX:    x,
		DirY:    y,
		Timer:   SearchTicks,
		TargetX: tx,
		TargetY: ty,
		PlanFor: randPlanTicks(),
	}
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
		switch org.State {
		case CellStateSearch:
			org.Timer--
			org.PlanFor--
			if shouldAvoidWall(org.Pos) {
				org.TargetX, org.TargetY = inwardTarget(org.Pos)
				org.PlanFor = randPlanTicks()
			} else if org.PlanFor <= 0 {
				org.TargetX, org.TargetY = randomWanderTarget(org.DirX, org.DirY)
				org.PlanFor = randPlanTicks()
			}

			desiredX, desiredY := org.TargetX, org.TargetY
			if fx, fy, ok := w.nearestFood(org.Pos.X, org.Pos.Y, FoodSenseSq); ok {
				toFoodX, toFoodY := normalize(fx-org.Pos.X, fy-org.Pos.Y)
				desiredX, desiredY = blendDirection(desiredX, desiredY, toFoodX, toFoodY, FoodSeekWeight)
			}

			org.DirX, org.DirY = steerToward(org.DirX, org.DirY, desiredX, desiredY, SearchTurnMax)
			org.DirX, org.DirY = normalize(org.DirX, org.DirY)
			org.Pos.X += org.DirX * SearchSpeed
			org.Pos.Y += org.DirY * SearchSpeed
			if fx, fy, ok := w.nearestFood(org.Pos.X, org.Pos.Y, SenseRadiusSq); ok {
				org.State = CellStateExploitCircle
				org.Timer = 0
				org.Loops = 0
				org.FoodX, org.FoodY = fx, fy
				org.HasFood = true
			}
			if org.Timer <= 0 {
				org.State = CellStateReorient
			}
		case CellStateExploitCircle:
			org.Timer++
			cx, cy := org.FoodX, org.FoodY
			dx := org.Pos.X - cx
			dy := org.Pos.Y - cy
			radius := math.Hypot(dx, dy)
			if radius < 1 {
				dx, dy = org.DirX*12, org.DirY*12
				radius = math.Hypot(dx, dy)
			}
			ux, uy := dx/radius, dy/radius
			tx, ty := -uy, ux

			phase := float64(org.Timer) * 0.35
			desiredRadius := 14 + 8*math.Sin(phase)
			desiredRadius = math.Max(8, desiredRadius)
			radialError := desiredRadius - radius
			desiredDX, desiredDY := normalize(tx+ux*radialError*0.15, ty+uy*radialError*0.15)
			org.DirX, org.DirY = steerToward(org.DirX, org.DirY, desiredDX, desiredDY, CircleTurn)
			org.DirX, org.DirY = normalize(org.DirX, org.DirY)
			org.Pos.X += org.DirX * CircleSpeed
			org.Pos.Y += org.DirY * CircleSpeed
			if org.Timer%12 == 0 {
				org.Loops++
			}
			if fx, fy, ok := w.nearestFood(org.Pos.X, org.Pos.Y, SenseRadiusSq); ok {
				org.FoodX, org.FoodY = fx, fy
				org.HasFood = true
			} else if org.Loops >= 2 {
				org.State = CellStateReorient
			}
		case CellStateReorient:
			org.TargetX, org.TargetY = randomWanderTarget(org.DirX, org.DirY)
			org.PlanFor = randPlanTicks()
			org.Timer = SearchTicks
			org.State = CellStateSearch
		}

		org.Pos.X = math.Max(0, math.Min(WorldWidth, org.Pos.X))
		org.Pos.Y = math.Max(0, math.Min(WorldHeight, org.Pos.Y))
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
				org.HasFood = true
				org.FoodX, org.FoodY = f.Pos.X, f.Pos.Y
				org.State = CellStateExploitCircle
				org.Timer = 0
				org.Loops = 0
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
