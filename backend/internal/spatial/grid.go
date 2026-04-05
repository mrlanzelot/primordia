package spatial

import (
	"math"

	"github.com/martin/primordia/internal/organism"
)

type Grid struct {
	cellSize float64
	buckets  map[[2]int][]uint64
}

// New creates a uniform grid used to speed up nearby-entity lookups.
func New(cellSize float64) *Grid {
	if cellSize <= 0 {
		cellSize = 1
	}
	return &Grid{
		cellSize: cellSize,
		buckets:  make(map[[2]int][]uint64),
	}
}

// key maps a world position into integer cell coordinates.
func (g *Grid) key(pos organism.Vec2) [2]int {
	return [2]int{int(pos.X / g.cellSize), int(pos.Y / g.cellSize)}
}

// Insert stores an entity id in the bucket that owns its position.
func (g *Grid) Insert(id uint64, pos organism.Vec2) {
	k := g.key(pos)
	g.buckets[k] = append(g.buckets[k], id)
}

// QueryRadius returns ids from all buckets intersecting a radius around pos.
func (g *Grid) QueryRadius(pos organism.Vec2, r float64) []uint64 {
	if r < 0 {
		return nil
	}
	base := g.key(pos)
	steps := int(math.Ceil(r / g.cellSize))
	out := make([]uint64, 0)
	for dx := -steps; dx <= steps; dx++ {
		for dy := -steps; dy <= steps; dy++ {
			k := [2]int{base[0] + dx, base[1] + dy}
			if ids, ok := g.buckets[k]; ok {
				out = append(out, ids...)
			}
		}
	}
	return out
}

// Clear removes all bucket contents so the grid can be rebuilt each tick.
func (g *Grid) Clear() {
	for k := range g.buckets {
		delete(g.buckets, k)
	}
}
