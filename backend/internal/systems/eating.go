package systems

import (
	"github.com/martin/primordia/internal/food"
	"github.com/martin/primordia/internal/organism"
	"github.com/martin/primordia/internal/spatial"
)

// ApplyEating transfers energy from nearby food to organisms and removes consumed food.
func ApplyEating(orgs map[uint64]*organism.Organism, foods map[uint64]*food.Food, grid *spatial.Grid) {
	eaten := make(map[uint64]struct{})
	for _, org := range orgs {
		for _, entityID := range grid.QueryRadius(org.Pos, 12) {
			if !isFoodEntityID(entityID) {
				continue
			}
			fid := foodIDFromEntityID(entityID)
			if _, taken := eaten[fid]; taken {
				continue
			}
			f, ok := foods[fid]
			if !ok {
				continue
			}
			dx := org.Pos.X - f.Pos.X
			dy := org.Pos.Y - f.Pos.Y
			if dx*dx+dy*dy < EatDistanceSq {
				org.Energy += EnergyFromFood
				if org.Energy > MaxEnergy {
					org.Energy = MaxEnergy
				}
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
		delete(foods, fid)
	}
}
