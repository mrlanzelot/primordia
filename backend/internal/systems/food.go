package systems

import "github.com/martin/primordia/internal/food"

func SpawnFood(foods map[uint64]*food.Food, nextID *uint64) {
	if len(foods) >= MaxFoodCount {
		return
	}
	*nextID = *nextID + 1
	foods[*nextID] = food.SpawnRandom(*nextID, WorldWidth, WorldHeight)
}
