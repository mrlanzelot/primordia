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

func New() *World {
	w := &World{
		Organisms: make(map[uint64]*organism.Organism),
		Foods:     make(map[uint64]*food.Food),
		Grid:      spatial.New(20),
	}
	for i := 0; i < InitialPopulation; i++ {
		w.NextID++
		w.Organisms[w.NextID] = systems.NewOrganism(w.NextID)
	}
	return w
}

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

func (w *World) rebuildGrid() {
	w.Grid.Clear()
	for id, o := range w.Organisms {
		w.Grid.Insert(id, o.Pos)
	}
	for id, f := range w.Foods {
		w.Grid.Insert(foodEntityMask|id, f.Pos)
	}
}
