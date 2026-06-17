package indicators

import (
	"math"

	domainInd "github.com/agoXQ/QuantLab/app/formula/domain/indicators"
	"github.com/agoXQ/QuantLab/app/formula/domain/series"
)

// cross implements CROSS(a, b): 1 on the bar where a transitions from
// <= b to > b, 0 otherwise. NaN is preserved on bars where either input is
// NaN, including the very first bar where the previous value is unknown.
func cross(args []domainInd.Arg) (series.Series, error) {
	if err := requireArgs("CROSS", len(args), 2); err != nil {
		return series.Series{}, err
	}
	anchor, _ := firstSeries(args)
	a, err := asSeries("CROSS", 0, args[0], anchor)
	if err != nil {
		return series.Series{}, err
	}
	b, err := asSeries("CROSS", 1, args[1], anchor)
	if err != nil {
		return series.Series{}, err
	}
	if a.Len() != b.Len() {
		return series.Series{}, fmtAlignError("CROSS")
	}

	out := make([]float64, a.Len())
	for i := 0; i < a.Len(); i++ {
		if i == 0 {
			out[i] = math.NaN()
			continue
		}
		av, bv := a.At(i), b.At(i)
		ap, bp := a.At(i-1), b.At(i-1)
		if math.IsNaN(av) || math.IsNaN(bv) || math.IsNaN(ap) || math.IsNaN(bp) {
			out[i] = math.NaN()
			continue
		}
		if ap <= bp && av > bv {
			out[i] = 1
		} else {
			out[i] = 0
		}
	}
	r, _ := series.NewSeries(a.Timestamps(), out)
	return r, nil
}

// longcross implements LONGCROSS(a, b, n): a CROSS where a has stayed strictly
// above b for the n previous bars before the cross point.
func longcross(args []domainInd.Arg) (series.Series, error) {
	if err := requireArgs("LONGCROSS", len(args), 3); err != nil {
		return series.Series{}, err
	}
	anchor, _ := firstSeries(args)
	a, err := asSeries("LONGCROSS", 0, args[0], anchor)
	if err != nil {
		return series.Series{}, err
	}
	b, err := asSeries("LONGCROSS", 1, args[1], anchor)
	if err != nil {
		return series.Series{}, err
	}
	n, err := asInt("LONGCROSS", 2, args[2])
	if err != nil {
		return series.Series{}, err
	}
	if a.Len() != b.Len() {
		return series.Series{}, fmtAlignError("LONGCROSS")
	}

	out := make([]float64, a.Len())
	for i := 0; i < a.Len(); i++ {
		if i < n+1 {
			out[i] = math.NaN()
			continue
		}
		held := true
		for k := 1; k <= n; k++ {
			if !(a.At(i-k) > b.At(i-k)) {
				held = false
				break
			}
		}
		if !held {
			out[i] = 0
			continue
		}
		if a.At(i-n-1) <= b.At(i-n-1) {
			out[i] = 1
		} else {
			out[i] = 0
		}
	}
	r, _ := series.NewSeries(a.Timestamps(), out)
	return r, nil
}

// filterSignal implements FILTER(signal, n): suppress repeated 1s within n
// bars of a previous 1.
func filterSignal(args []domainInd.Arg) (series.Series, error) {
	if err := requireArgs("FILTER", len(args), 2); err != nil {
		return series.Series{}, err
	}
	anchor, _ := firstSeries(args)
	s, err := asSeries("FILTER", 0, args[0], anchor)
	if err != nil {
		return series.Series{}, err
	}
	n, err := asInt("FILTER", 1, args[1])
	if err != nil {
		return series.Series{}, err
	}

	out := make([]float64, s.Len())
	cooldown := 0
	for i := 0; i < s.Len(); i++ {
		v := s.At(i)
		if math.IsNaN(v) {
			out[i] = math.NaN()
			continue
		}
		if cooldown > 0 {
			cooldown--
			out[i] = 0
			continue
		}
		if v != 0 {
			out[i] = 1
			cooldown = n
		} else {
			out[i] = 0
		}
	}
	r, _ := series.NewSeries(s.Timestamps(), out)
	return r, nil
}
