package systems

import "github.com/martin/primordia/internal/food"

// SpawnFood enforces the food cap and appends at most one new food per tick.
func SpawnFood(foods map[uint64]*food.Food, nextID *uint64) {
	if len(foods) >= MaxFoodCount {
		return
	}
	*nextID = *nextID + 1
	foods[*nextID] = food.SpawnRandom(*nextID, WorldWidth, WorldHeight)
}
