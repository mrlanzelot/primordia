package systems

import (
	"math"
	"math/rand"

	"github.com/martin/primordia/internal/food"
	"github.com/martin/primordia/internal/organism"
)

const (
	WorldWidth        = 1000
	WorldHeight       = 1000
	EnergyCost        = 0.2
	EnergyFromFood    = 35.0
	EatDistanceSq     = 144.0
	InitialEnergy     = 80.0
	MaxEnergy         = 160.0
	MaxSpeed          = 1.0
	InitialPopulation = 100
	MaxFoodCount      = 135

	SearchWobble  = 0.10
	SearchSpeed   = 1.0
	CircleTurn    = 0.22
	CircleSpeed   = 1.0
	SearchTicks   = 24
	CircleTicks   = 10
	SenseRadiusSq = 500.0

	SearchTurnMax   = 0.09
	PlanTicksMin    = 18
	PlanTicksMax    = 60
	WallAvoidMargin = 50.0
	FoodSeekWeight  = 0.28
	FoodSenseSq     = 6400.0
)

const (
	CellStateSearch        = "search"
	CellStateExploitCircle = "exploit_circle"
	CellStateReorient      = "reorient"
)

// normalize converts a direction vector to unit length with zero fallback.
func normalize(x, y float64) (float64, float64) {
	l := math.Hypot(x, y)
	if l == 0 {
		return 1, 0
	}
	return x / l, y / l
}

// rotate rotates a vector in 2D space by angle a.
func rotate(x, y, a float64) (float64, float64) {
	c, s := math.Cos(a), math.Sin(a)
	return x*c - y*s, x*s + y*c
}

// wrapAngle constrains an angle to [-pi, pi] for stable steering math.
func wrapAngle(a float64) float64 {
	for a > math.Pi {
		a -= 2 * math.Pi
	}
	for a < -math.Pi {
		a += 2 * math.Pi
	}
	return a
}

// steerToward limits heading change toward a target direction by maxTurn.
func steerToward(curX, curY, targetX, targetY, maxTurn float64) (float64, float64) {
	curX, curY = normalize(curX, curY)
	targetX, targetY = normalize(targetX, targetY)
	curA := math.Atan2(curY, curX)
	targetA := math.Atan2(targetY, targetX)
	delta := wrapAngle(targetA - curA)
	if delta > maxTurn {
		delta = maxTurn
	} else if delta < -maxTurn {
		delta = -maxTurn
	}
	return rotate(curX, curY, delta)
}

// randPlanTicks picks how long an organism keeps its current wander plan.
func randPlanTicks() int {
	return PlanTicksMin + rand.Intn(PlanTicksMax-PlanTicksMin+1)
}

// shouldAvoidWall flags organisms that are close enough to boundaries to steer inward.
func shouldAvoidWall(p organism.Vec2) bool {
	return p.X < WallAvoidMargin || p.X > WorldWidth-WallAvoidMargin || p.Y < WallAvoidMargin || p.Y > WorldHeight-WallAvoidMargin
}

// inwardTarget returns the normalized direction toward world center.
func inwardTarget(p organism.Vec2) (float64, float64) {
	return normalize((WorldWidth/2)-p.X, (WorldHeight/2)-p.Y)
}

// randomWanderTarget perturbs the current heading to create exploratory motion.
func randomWanderTarget(dirX, dirY float64) (float64, float64) {
	turn := (rand.Float64()-0.5)*0.8 + (rand.Float64()-0.5)*SearchWobble
	return normalize(rotate(dirX, dirY, turn))
}

// blendDirection mixes two headings with a weighted influence.
func blendDirection(baseX, baseY, addX, addY, weight float64) (float64, float64) {
	bx, by := normalize(baseX, baseY)
	ax, ay := normalize(addX, addY)
	return normalize(bx*(1-weight)+ax*weight, by*(1-weight)+ay*weight)
}

// nearestFood finds the closest food position within a squared-distance limit.
func nearestFood(x, y float64, maxSq float64, foods map[uint64]*food.Food) (float64, float64, bool) {
	best := maxSq
	var bx, by float64
	found := false
	for _, f := range foods {
		dx := x - f.Pos.X
		dy := y - f.Pos.Y
		d := dx*dx + dy*dy
		if d <= best {
			best = d
			bx, by = f.Pos.X, f.Pos.Y
			found = true
		}
	}
	return bx, by, found
}

// NewOrganism initializes one organism with random spawn and search state defaults.
func NewOrganism(id uint64) *organism.Organism {
	ang := rand.Float64() * 2 * math.Pi
	x, y := math.Cos(ang), math.Sin(ang)
	tx, ty := randomWanderTarget(x, y)
	return &organism.Organism{
		ID:       id,
		Pos:      organism.Vec2{X: rand.Float64() * WorldWidth, Y: rand.Float64() * WorldHeight},
		Energy:   InitialEnergy,
		State:    CellStateSearch,
		DirX:     x,
		DirY:     y,
		Timer:    SearchTicks,
		TargetX:  tx,
		TargetY:  ty,
		PlanFor:  randPlanTicks(),
		Angle:    math.Atan2(y, x),
		SenseVec: make([]float64, 21),
	}
}

// UpdateOrganisms preserves the pre-refactor movement and lifecycle behavior.
func UpdateOrganisms(orgs map[uint64]*organism.Organism, foods map[uint64]*food.Food) {
	deadIDs := make([]uint64, 0)
	for id, org := range orgs {
		prev := org.Pos
		switch org.State {
		case CellStateSearch:
			org.Timer--
			org.PlanFor--
			if shouldAvoidWall(org.Pos) {
				org.TargetX, org.TargetY = inwardTarget(org.Pos)
				org.PlanFor = randPlanTicks()
			} else if org.PlanFor <= 0 {
				org.TargetX, org.TargetY = randomWanderTarget(org.DirX, org.DirY)
				org.PlanFor = randPlanTicks()
			}

			desiredX, desiredY := org.TargetX, org.TargetY
			if fx, fy, ok := nearestFood(org.Pos.X, org.Pos.Y, FoodSenseSq, foods); ok {
				toFoodX, toFoodY := normalize(fx-org.Pos.X, fy-org.Pos.Y)
				desiredX, desiredY = blendDirection(desiredX, desiredY, toFoodX, toFoodY, FoodSeekWeight)
			}

			org.DirX, org.DirY = steerToward(org.DirX, org.DirY, desiredX, desiredY, SearchTurnMax)
			org.DirX, org.DirY = normalize(org.DirX, org.DirY)
			org.Pos.X += org.DirX * SearchSpeed
			org.Pos.Y += org.DirY * SearchSpeed
			if fx, fy, ok := nearestFood(org.Pos.X, org.Pos.Y, SenseRadiusSq, foods); ok {
				org.State = CellStateExploitCircle
				org.Timer = 0
				org.Loops = 0
				org.FoodX, org.FoodY = fx, fy
				org.HasFood = true
			}
			if org.Timer <= 0 {
				org.State = CellStateReorient
			}
		case CellStateExploitCircle:
			org.Timer++
			cx, cy := org.FoodX, org.FoodY
			dx := org.Pos.X - cx
			dy := org.Pos.Y - cy
			radius := math.Hypot(dx, dy)
			if radius < 1 {
				dx, dy = org.DirX*12, org.DirY*12
				radius = math.Hypot(dx, dy)
			}
			ux, uy := dx/radius, dy/radius
			tx, ty := -uy, ux

			phase := float64(org.Timer) * 0.35
			desiredRadius := 14 + 8*math.Sin(phase)
			desiredRadius = math.Max(8, desiredRadius)
			radialError := desiredRadius - radius
			desiredDX, desiredDY := normalize(tx+ux*radialError*0.15, ty+uy*radialError*0.15)
			org.DirX, org.DirY = steerToward(org.DirX, org.DirY, desiredDX, desiredDY, CircleTurn)
			org.DirX, org.DirY = normalize(org.DirX, org.DirY)
			org.Pos.X += org.DirX * CircleSpeed
			org.Pos.Y += org.DirY * CircleSpeed
			if org.Timer%12 == 0 {
				org.Loops++
			}
			if fx, fy, ok := nearestFood(org.Pos.X, org.Pos.Y, SenseRadiusSq, foods); ok {
				org.FoodX, org.FoodY = fx, fy
				org.HasFood = true
			} else if org.Loops >= 2 {
				org.State = CellStateReorient
			}
		case CellStateReorient:
			org.TargetX, org.TargetY = randomWanderTarget(org.DirX, org.DirY)
			org.PlanFor = randPlanTicks()
			org.Timer = SearchTicks
			org.State = CellStateSearch
		}

		org.Pos.X = math.Max(0, math.Min(WorldWidth, org.Pos.X))
		org.Pos.Y = math.Max(0, math.Min(WorldHeight, org.Pos.Y))
		org.Energy -= EnergyCost
		if org.Energy <= 0 {
			deadIDs = append(deadIDs, id)
		}
		org.Vel = org.Pos.Sub(prev)
		if org.Vel.Length() > 0 {
			org.Angle = math.Atan2(org.Vel.Y, org.Vel.X)
		}
		org.Age++
	}
	for _, id := range deadIDs {
		delete(orgs, id)
	}
}
