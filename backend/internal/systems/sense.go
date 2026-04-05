package systems

import (
	"math"

	"github.com/martin/primordia/internal/food"
	"github.com/martin/primordia/internal/organism"
	"github.com/martin/primordia/internal/spatial"
)

const (
	senseRayCount    = 8
	senseArcRadians  = 4 * math.Pi / 3
	senseRayMax      = 150.0
	senseRayStep     = 5.0
	senseHitRadius   = 5.0
	smellRadius      = 200.0
	smellScaleFactor = 0.05
	speedSenseScale  = 1.7
)

// UpdateSense rebuilds each organism's 21-value sense vector for brain input.
func UpdateSense(orgs map[uint64]*organism.Organism, foods map[uint64]*food.Food, grid *spatial.Grid) {
	for _, org := range orgs {
		sv := make([]float64, 21)
		fillRaySense(sv, org, orgs, foods, grid)
		fillSmellSense(sv, org, foods, grid)
		sv[19] = clamp01(org.Energy / MaxEnergy)
		sv[20] = clamp01(org.Vel.Length() / (MaxSpeed * speedSenseScale))
		org.SenseVec = sv
	}
}

// fillRaySense traces forward-arc rays and stores first-hit distance/type pairs.
func fillRaySense(sv []float64, org *organism.Organism, orgs map[uint64]*organism.Organism, foods map[uint64]*food.Food, grid *spatial.Grid) {
	start := org.Angle - (senseArcRadians / 2)
	stepA := senseArcRadians / float64(senseRayCount-1)

	for ray := 0; ray < senseRayCount; ray++ {
		rayAngle := start + float64(ray)*stepA
		dir := organism.Vec2{X: math.Cos(rayAngle), Y: math.Sin(rayAngle)}
		distNorm := 1.0
		typeVal := 0.0

	found:
		for d := senseRayStep; d <= senseRayMax; d += senseRayStep {
			sample := org.Pos.Add(dir.Scale(d))
			ids := grid.QueryRadius(sample, senseHitRadius)
			for _, entityID := range ids {
				if isFoodEntityID(entityID) {
					fid := foodIDFromEntityID(entityID)
					f, ok := foods[fid]
					if !ok {
						continue
					}
					if sample.Sub(f.Pos).Length() <= senseHitRadius {
						distNorm = d / senseRayMax
						typeVal = 1.0
						break found
					}
					continue
				}

				oid := entityID
				if oid == org.ID {
					continue
				}
				o, ok := orgs[oid]
				if !ok {
					continue
				}
				if sample.Sub(o.Pos).Length() <= senseHitRadius {
					distNorm = d / senseRayMax
					typeVal = 2.0
					break found
				}
			}
		}

		base := ray * 2
		sv[base] = clamp01(distNorm)
		sv[base+1] = typeVal
	}
}

// fillSmellSense estimates local food gradient strength and direction.
func fillSmellSense(sv []float64, org *organism.Organism, foods map[uint64]*food.Food, grid *spatial.Grid) {
	var rawX, rawY float64
	ids := grid.QueryRadius(org.Pos, smellRadius)
	for _, entityID := range ids {
		if !isFoodEntityID(entityID) {
			continue
		}
		fid := foodIDFromEntityID(entityID)
		f, ok := foods[fid]
		if !ok {
			continue
		}
		delta := f.Pos.Sub(org.Pos)
		dist := delta.Length()
		if dist == 0 {
			continue
		}
		if dist > smellRadius {
			continue
		}
		norm := delta.Scale(1 / dist)
		weight := 1.0 / dist
		rawX += norm.X * weight
		rawY += norm.Y * weight
	}

	rawMag := math.Hypot(rawX, rawY)
	dirX, dirY := 0.0, 0.0
	if rawMag > 0 {
		dirX = rawX / rawMag
		dirY = rawY / rawMag
	}

	sv[16] = math.Tanh(rawMag / smellScaleFactor)
	sv[17] = dirX
	sv[18] = dirY
}

// clamp01 bounds normalized sensor values into [0, 1].
func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}
