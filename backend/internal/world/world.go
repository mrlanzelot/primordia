package world

import (
	"sync"

	"github.com/martin/primordia/internal/food"
	"github.com/martin/primordia/internal/organism"
	"github.com/martin/primordia/internal/spatial"
	"github.com/martin/primordia/internal/systems"
)

const (
	WorldWidth        = systems.WorldWidth
	WorldHeight       = systems.WorldHeight
	InitialPopulation = systems.InitialPopulation
)

const foodEntityMask = uint64(1) << 63

type World struct {
	Organisms map[uint64]*organism.Organism
	Foods     map[uint64]*food.Food
	Grid      *spatial.Grid
	Mu        sync.RWMutex
	NextID    uint64
	TickID    uint64
}

// seedInitialPopulation fills the organism map with the configured initial population.
func (w *World) seedInitialPopulation() {
	for i := 0; i < InitialPopulation; i++ {
		w.NextID++
		w.Organisms[w.NextID] = systems.NewOrganism(w.NextID)
	}
}

// New builds initial world state and seeds the starting organism population.
func New() *World {
	w := &World{
		Organisms: make(map[uint64]*organism.Organism),
		Foods:     make(map[uint64]*food.Food),
		Grid:      spatial.New(20),
	}
	w.seedInitialPopulation()
	return w
}

// Reset clears dynamic state and reseeds the world to its initial starting conditions.
func (w *World) Reset() {
	w.Mu.Lock()
	defer w.Mu.Unlock()
	w.Organisms = make(map[uint64]*organism.Organism)
	w.Foods = make(map[uint64]*food.Food)
	w.NextID = 0
	w.TickID = 0
	w.Grid.Clear()
	w.seedInitialPopulation()
}

// Tick runs one simulation step in deterministic system order under a write lock.
func (w *World) Tick() {
	w.Mu.Lock()
	defer w.Mu.Unlock()

	w.rebuildGrid()
	systems.UpdateSense(w.Organisms, w.Foods, w.Grid)
	systems.UpdateOrganisms(w.Organisms, w.Foods)
	w.rebuildGrid()
	systems.ApplyEating(w.Organisms, w.Foods, w.Grid)
	systems.SpawnFood(w.Foods, &w.NextID)
	w.TickID++
}

// OrganismCount returns the current population size under a read lock.
func (w *World) OrganismCount() int {
	w.Mu.RLock()
	defer w.Mu.RUnlock()
	return len(w.Organisms)
}

// rebuildGrid repopulates spatial buckets with current organism and food positions.
func (w *World) rebuildGrid() {
	w.Grid.Clear()
	for id, o := range w.Organisms {
		w.Grid.Insert(id, o.Pos)
	}
	for id, f := range w.Foods {
		w.Grid.Insert(foodEntityMask|id, f.Pos)
	}
}
