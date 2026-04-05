package systems

const foodEntityMask = uint64(1) << 63

// entityIDForOrganism keeps organism ids unchanged for grid storage.
func entityIDForOrganism(id uint64) uint64 {
	return id
}

// entityIDForFood tags food ids so they can share one grid id namespace.
func entityIDForFood(id uint64) uint64 {
	return foodEntityMask | id
}

// isFoodEntityID checks whether a grid id refers to a food item.
func isFoodEntityID(v uint64) bool {
	return v&foodEntityMask != 0
}

// foodIDFromEntityID strips the food tag bit to recover original id.
func foodIDFromEntityID(v uint64) uint64 {
	return v &^ foodEntityMask
}
