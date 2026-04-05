package food

import (
	"math/rand"

	"github.com/martin/primordia/internal/organism"
)

type Food struct {
	ID  uint64
	Pos organism.Vec2
}

func SpawnRandom(id uint64, width, height float64) *Food {
	return &Food{
		ID: id,
		Pos: organism.Vec2{
			X: rand.Float64() * width,
			Y: rand.Float64() * height,
		},
	}
}
