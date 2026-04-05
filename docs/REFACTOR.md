# Primordia — Backend Refactor + Sense System

## Context

You are working on **Primordia**, an evolution simulation.

- **Backend**: Go (single file at `backend/main.go`)
- **Frontend**: React + PixiJS, consumes WebSocket world state
- **Goal**: Refactor the backend into clean packages, then implement a sense system for organisms

The backend currently runs these systems each tick:
1. `updateOrganisms` — movement, boundary clamp, energy decay, death
2. `spawnFood` — food cap enforcement, random spawn
3. `applyEating` — proximity check, energy gain

## Task 1 — Refactor `backend/main.go` into packages

Split the single file into the following package structure. Do not change any behavior — this is a pure structural refactor.

```
backend/
├── cmd/primordia/
│   └── main.go              # flags, wiring, signal context only
└── internal/
    ├── world/
    │   └── world.go         # World struct, Tick(), RWMutex
    ├── organism/
    │   └── organism.go      # Organism struct
    ├── spatial/
    │   └── grid.go          # UniformGrid: Insert, QueryRadius, Clear
    ├── food/
    │   └── food.go          # Food struct, spawner
    ├── systems/
    │   ├── movement.go      # updateOrganisms logic
    │   ├── eating.go        # applyEating logic
    │   ├── sense.go         # NEW — fill SenseVec each tick
    │   └── brain.go         # STUB — placeholder for Phase 2
    ├── protocol/
    │   └── protocol.go      # WorldMsg, OrganismMsg, FoodMsg
    └── server/
        └── server.go        # WebSocket hub, broadcast loop, /ws handler
```

### Organism struct

Add these fields even if not yet used:

```go
type Organism struct {
    ID       uint64
    Pos      Vec2
    Vel      Vec2
    Angle    float64
    Energy   float64
    Age      int
    Genome   []float64   // placeholder for NN weights (Phase 2)
    SenseVec []float64   // filled by systems/sense each tick
}
```

### Protocol structs

```go
type OrganismMsg struct {
    ID       uint64    `json:"id"`
    X        float32   `json:"x"`
    Y        float32   `json:"y"`
    Angle    float32   `json:"a"`
    Energy   float32   `json:"e"`
    SenseVec []float32 `json:"sv,omitempty"`
    Selected bool      `json:"sel,omitempty"`
}

type FoodMsg struct {
    X float32 `json:"x"`
    Y float32 `json:"y"`
}

type WorldMsg struct {
    Tick      uint64        `json:"tick"`
    Organisms []OrganismMsg `json:"organisms"`
    Foods     []FoodMsg     `json:"foods"`
}
```

### Spatial grid

Implement a simple uniform grid in `internal/spatial/grid.go`:

```go
type Grid struct {
    cellSize float64
    buckets  map[[2]int][]uint64
}

func New(cellSize float64) *Grid
func (g *Grid) Insert(id uint64, pos Vec2)
func (g *Grid) QueryRadius(pos Vec2, r float64) []uint64
func (g *Grid) Clear()
```

Bucket key = `[2]int{int(pos.X / cellSize), int(pos.Y / cellSize)}`.
`QueryRadius` checks all neighboring buckets within ceil(r/cellSize) steps.

### Tick order after refactor

```
grid.Clear()
grid.Insert(all organisms + food)
systems/sense    — fill SenseVec via grid queries
systems/movement — apply velocity, decay energy, clamp bounds, death
systems/eating   — grid query, energy transfer, remove food
systems/food     — enforce cap, spawn new food
broadcast snapshot
```

---

## Task 2 — Implement `systems/sense.go`

Each tick, fill every organism's `SenseVec` with a fixed-length input vector.

### Sense vector layout

| Index | Description |
|-------|-------------|
| 0–15  | 8 rays × 2 values: (normalized distance 0–1, type 0=empty 1=food 2=organism) |
| 16    | Smell strength (food gradient magnitude at organism pos) |
| 17    | Smell direction X (normalized) |
| 18    | Smell direction Y (normalized) |
| 19    | Self energy (normalized 0–1, capped at max energy) |
| 20    | Self speed (normalized 0–1) |

Total: 21 floats.

### Raycasting

- Cast 8 rays evenly spaced across a **240° forward arc** centered on `Organism.Angle`
- Max ray length: **150 world units**
- For each ray, step in increments of **5 world units**
- At each step, call `grid.QueryRadius(stepPos, 5)` to check for hits
- First hit wins; record normalized distance and entity type
- If no hit: distance = 1.0, type = 0.0

### Smell gradient

- Sample food positions within radius **200** using `grid.QueryRadius`
- Sum inverse-distance-weighted vectors toward each food
- Normalize the result; magnitude = tanh(raw magnitude / 50)

### Self sensors

- Energy: `organism.Energy / MaxEnergy` clamped to [0, 1]
- Speed: `Vec2.Length(organism.Vel) / MaxSpeed` clamped to [0, 1]

---

## Task 3 — Stub `systems/brain.go`

Add a no-op brain that returns zero actions. This will be replaced in Phase 2.

```go
// ActionVec represents the output of the brain
type ActionVec struct {
    TurnDelta float64  // radians/tick
    Thrust    float64  // 0–1
    EatFlag   float64  // > 0.5 = attempt eat
}

// Think takes a sense vector and returns an action vector.
// For now returns random walk behaviour so organisms still move.
func Think(sv []float64) ActionVec {
    // TODO: replace with neural network forward pass in Phase 2
    return ActionVec{
        TurnDelta: (rand.Float64()*2 - 1) * 0.1,
        Thrust:    0.5 + rand.Float64()*0.5,
    }
}
```

---

## Constraints

- Preserve all existing simulation behaviour exactly
- All new packages must have a `_test.go` with at least one unit test
- `spatial.Grid` must have tests for `Insert` + `QueryRadius`
- `systems/sense` must have a test with a synthetic world (1 organism, 1 food in front)
- No external dependencies — stdlib only
- `go vet ./...` and `go build ./...` must pass with zero errors