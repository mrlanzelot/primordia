package protocol

import "github.com/martin/primordia/internal/world"

type OrganismMsg struct {
	ID       uint64    `json:"id"`
	X        float32   `json:"x"`
	Y        float32   `json:"y"`
	Angle    float32   `json:"a"`
	Energy   float32   `json:"e"`
	SenseVec []float32 `json:"sv,omitempty"`
	Selected bool      `json:"sel,omitempty"`
}

type FoodMsg struct {
	X float32 `json:"x"`
	Y float32 `json:"y"`
}

type WorldMsg struct {
	Tick      uint64        `json:"tick"`
	Organisms []OrganismMsg `json:"organisms"`
	Foods     []FoodMsg     `json:"foods"`
}

func Snapshot(w *world.World) WorldMsg {
	w.Mu.RLock()
	defer w.Mu.RUnlock()

	msg := WorldMsg{
		Tick:      w.TickID,
		Organisms: make([]OrganismMsg, 0, len(w.Organisms)),
		Foods:     make([]FoodMsg, 0, len(w.Foods)),
	}
	for _, org := range w.Organisms {
		o := OrganismMsg{
			ID:     org.ID,
			X:      float32(org.Pos.X),
			Y:      float32(org.Pos.Y),
			Angle:  float32(org.Angle),
			Energy: float32(org.Energy),
		}
		if len(org.SenseVec) > 0 {
			o.SenseVec = make([]float32, len(org.SenseVec))
			for i, v := range org.SenseVec {
				o.SenseVec[i] = float32(v)
			}
		}
		msg.Organisms = append(msg.Organisms, o)
	}
	for _, f := range w.Foods {
		msg.Foods = append(msg.Foods, FoodMsg{X: float32(f.Pos.X), Y: float32(f.Pos.Y)})
	}
	return msg
}
