package systems

const foodEntityMask = uint64(1) << 63

func entityIDForOrganism(id uint64) uint64 {
	return id
}

func entityIDForFood(id uint64) uint64 {
	return foodEntityMask | id
}

func isFoodEntityID(v uint64) bool {
	return v&foodEntityMask != 0
}

func foodIDFromEntityID(v uint64) uint64 {
	return v &^ foodEntityMask
}
