package chartgen

import "gonum.org/v1/plot"

// TimeTicker helps draw time ticks
type TimeTicker struct {
	Delta        int64
	BetweenCount int
}

// Ticks returns Ticks in the specified range.
func (t *TimeTicker) Ticks(min, max float64) []plot.Tick {
	if max <= min {
		panic("illegal range")
	}

	minTS, maxTS := int64(min), int64(max)
	nearestLeftStart := minTS - (minTS % t.Delta)

	ticks := make([]plot.Tick, 0)
	for i := nearestLeftStart; i < maxTS; i += t.Delta {
		ticks = append(ticks, plot.Tick{Value: float64(i), Label: "fill_me"})

		miniDelta := float64(t.Delta) / float64(t.BetweenCount+1)
		for j := 1; j <= t.BetweenCount; j++ {
			ticks = append(ticks, plot.Tick{Value: float64(i) + float64(j)*miniDelta, Label: ""})
		}
	}
	return ticks
}
