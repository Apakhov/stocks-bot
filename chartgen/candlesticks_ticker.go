package chartgen

import (
	"math"
	"strconv"

	"gonum.org/v1/plot"
)

const (
	// dlamchE is the machine epsilon. For IEEE this is 2^{-53}.
	dlamchE = 1.0 / (1 << 53)

	// dlamchB is the radix of the machine (the base of the number system).
	dlamchB = 2

	// dlamchP is base * eps.
	dlamchP = dlamchB * dlamchE
)

const (
	// free indicates no restriction on label containment.
	free = iota
	// containData specifies that all the data range lies
	// within the interval [label_min, label_max].
	containData
	// withinData specifies that all labels lie within the
	// interval [dMin, dMax].
	withinData
)

// talbotLinHanrahan returns an optimal set of approximately want label values
// for the data range [dMin, dMax], and the step and magnitude of the step between values.
// containment is specifies are guarantees for label and data range containment, valid
// values are free, containData and withinData.
// The optional parameters Q, nice numbers, and w, weights, allow tuning of the
// algorithm but by default (when nil) are set to the parameters described in the
// paper.
// The legibility function allows tuning of the legibility assessment for labels.
// By default, when nil, legbility will set the legibility score for each candidate
// labelling scheme to 1.
// See the paper for an explanation of the function of Q, w and legibility.
func talbotLinHanrahan(dMin, dMax float64, want int, containment int, Q []float64, w *weights, legibility func(lMin, lMax, lStep float64) float64) (values []float64, step, q float64, magnitude int) {
	const eps = dlamchP * 100

	if dMin > dMax {
		panic("labelling: invalid data range: min greater than max")
	}

	if Q == nil {
		Q = []float64{1, 5, 2, 2.5, 4, 3}
	}
	if w == nil {
		w = &weights{
			simplicity: 0.25,
			coverage:   0.2,
			density:    0.5,
			legibility: 0.05,
		}
	}
	if legibility == nil {
		legibility = unitLegibility
	}

	if r := dMax - dMin; r < eps {
		l := make([]float64, want)
		step := r / float64(want-1)
		for i := range l {
			l[i] = dMin + float64(i)*step
		}
		magnitude = minAbsMag(dMin, dMax)
		return l, step, 0, magnitude
	}

	type selection struct {
		// n is the number of labels selected.
		n int
		// lMin and lMax are the selected min
		// and max label values. lq is the q
		// chosen.
		lMin, lMax, lStep, lq float64
		// score is the score for the selection.
		score float64
		// magnitude is the magnitude of the
		// label step distance.
		magnitude int
	}
	best := selection{score: -2}

outer:
	for skip := 1; ; skip++ {
		for _, q := range Q {
			sm := maxSimplicity(q, Q, skip)
			if w.score(sm, 1, 1, 1) < best.score {
				break outer
			}

			for have := 2; ; have++ {
				dm := maxDensity(have, want)
				if w.score(sm, 1, dm, 1) < best.score {
					break
				}

				delta := (dMax - dMin) / float64(have+1) / float64(skip) / q

				const maxExp = 309
				for mag := int(math.Ceil(math.Log10(delta))); mag < maxExp; mag++ {
					step := float64(skip) * q * math.Pow10(mag)

					cm := maxCoverage(dMin, dMax, step*float64(have-1))
					if w.score(sm, cm, dm, 1) < best.score {
						break
					}

					fracStep := step / float64(skip)
					kStep := step * float64(have-1)

					minStart := (math.Floor(dMax/step) - float64(have-1)) * float64(skip)
					maxStart := math.Ceil(dMax/step) * float64(skip)
					for start := minStart; start <= maxStart && start != start-1; start++ {
						lMin := start * fracStep
						lMax := lMin + kStep

						switch containment {
						case containData:
							if dMin < lMin || lMax < dMax {
								continue
							}
						case withinData:
							if lMin < dMin || dMax < lMax {
								continue
							}
						case free:
							// Free choice.
						}

						score := w.score(
							simplicity(q, Q, skip, lMin, lMax, step),
							coverage(dMin, dMax, lMin, lMax),
							density(have, want, dMin, dMax, lMin, lMax),
							legibility(lMin, lMax, step),
						)
						if score > best.score {
							best = selection{
								n:         have,
								lMin:      lMin,
								lMax:      lMax,
								lStep:     float64(skip) * q,
								lq:        q,
								score:     score,
								magnitude: mag,
							}
						}
					}
				}
			}
		}
	}

	if best.score == -2 {
		l := make([]float64, want)
		step := (dMax - dMin) / float64(want-1)
		for i := range l {
			l[i] = dMin + float64(i)*step
		}
		magnitude = minAbsMag(dMin, dMax)
		return l, step, 0, magnitude
	}

	l := make([]float64, best.n)
	step = best.lStep * math.Pow10(best.magnitude)
	for i := range l {
		l[i] = best.lMin + float64(i)*step
	}
	return l, best.lStep, best.lq, best.magnitude
}

// minAbsMag returns the minumum magnitude of the absolute values of a and b.
func minAbsMag(a, b float64) int {
	return int(math.Min(math.Floor(math.Log10(math.Abs(a))), (math.Floor(math.Log10(math.Abs(b))))))
}

// simplicity returns the simplicity score for how will the curent q, lMin, lMax,
// lStep and skip match the given nice numbers, Q.
func simplicity(q float64, Q []float64, skip int, lMin, lMax, lStep float64) float64 {
	const eps = dlamchP * 100

	for i, v := range Q {
		if v == q {
			m := math.Mod(lMin, lStep)
			v = 0
			if (m < eps || lStep-m < eps) && lMin <= 0 && 0 <= lMax {
				v = 1
			}
			return 1 - float64(i)/(float64(len(Q))-1) - float64(skip) + v
		}
	}
	panic("labelling: invalid q for Q")
}

// maxSimplicity returns the maximum simplicity for q, Q and skip.
func maxSimplicity(q float64, Q []float64, skip int) float64 {
	for i, v := range Q {
		if v == q {
			return 1 - float64(i)/(float64(len(Q))-1) - float64(skip) + 1
		}
	}
	panic("labelling: invalid q for Q")
}

// coverage returns the coverage score for based on the average
// squared distance between the extreme labels, lMin and lMax, and
// the extreme data points, dMin and dMax.
func coverage(dMin, dMax, lMin, lMax float64) float64 {
	r := 0.1 * (dMax - dMin)
	max := dMax - lMax
	min := dMin - lMin
	return 1 - 0.5*(max*max+min*min)/(r*r)
}

// maxCoverage returns the maximum coverage achievable for the data
// range.
func maxCoverage(dMin, dMax, span float64) float64 {
	r := dMax - dMin
	if span <= r {
		return 1
	}
	h := 0.5 * (span - r)
	r *= 0.1
	return 1 - (h*h)/(r*r)
}

// density returns the density score which measures the goodness of
// the labelling density compared to the user defined target
// based on the want parameter given to talbotLinHanrahan.
func density(have, want int, dMin, dMax, lMin, lMax float64) float64 {
	rho := float64(have-1) / (lMax - lMin)
	rhot := float64(want-1) / (math.Max(lMax, dMax) - math.Min(dMin, lMin))
	if d := rho / rhot; d >= 1 {
		return 2 - d
	}
	return 2 - rhot/rho
}

// maxDensity returns the maximum density score achievable for have and want.
func maxDensity(have, want int) float64 {
	if have < want {
		return 1
	}
	return 2 - float64(have-1)/float64(want-1)
}

// unitLegibility returns a default legibility score ignoring label
// spacing.
func unitLegibility(_, _, _ float64) float64 {
	return 1
}

// weights is a helper type to calcuate the labelling scheme's total score.
type weights struct {
	simplicity, coverage, density, legibility float64
}

// score returns the score for a labelling scheme with simplicity, s,
// coverage, c, density, d and legibility l.
func (w *weights) score(s, c, d, l float64) float64 {
	return w.simplicity*s + w.coverage*c + w.density*d + w.legibility*l
}

// CandlesticksTicker helps draw candlesticks ticks
type CandlesticksTicker struct {
	WantLables int
}

// Ticks returns Ticks in the specified range.
func (t *CandlesticksTicker) Ticks(min, max float64) []plot.Tick {
	if max <= min {
		panic("illegal range")
	}

	labels, step, q, mag := talbotLinHanrahan(min, max, t.WantLables, free, nil, nil, nil)
	majorDelta := step * math.Pow10(mag)
	if q == 0 {
		// Simple fall back was chosen, so
		// majorDelta is the label distance.
		majorDelta = labels[1] - labels[0]
	}

	// Choose a reasonable, but ad
	// hoc formatting for labels.
	fc := byte('f')
	prec := 2

	ticks := make([]plot.Tick, 0, len(labels)+2)
	if math.Abs(min-labels[0]) > 1e-6 {
		ticks = append(ticks, plot.Tick{Value: min, Label: strconv.FormatFloat(min, fc, prec, 64)})
	}
	for _, v := range labels {
		ticks = append(ticks, plot.Tick{Value: v, Label: strconv.FormatFloat(v, fc, prec, 64)})
	}
	if math.Abs(max-labels[len(labels)-1]) > 1e-6 {
		ticks = append(ticks, plot.Tick{Value: max, Label: strconv.FormatFloat(max, fc, prec, 64)})
	}

	var minorDelta float64
	// See talbotLinHanrahan for the values used here.
	switch step {
	case 1, 2.5:
		minorDelta = majorDelta / 5
	case 2, 3, 4, 5:
		minorDelta = majorDelta / step
	default:
		if majorDelta/2 < dlamchP {
			return ticks
		}
		minorDelta = majorDelta / 2
	}

	// Find the first minor tick not greater
	// than the lowest data value.
	var i float64
	for labels[0]+(i-1)*minorDelta > min {
		i--
	}
	// Add ticks at minorDelta intervals when
	// they are not within minorDelta/2 of a
	// labelled tick.
	for {
		val := labels[0] + i*minorDelta
		if val > max {
			break
		}
		found := false
		for _, t := range ticks {
			if math.Abs(t.Value-val) < minorDelta/2 {
				found = true
			}
		}
		if !found {
			ticks = append(ticks, plot.Tick{Value: val})
		}
		i++
	}

	return ticks
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
