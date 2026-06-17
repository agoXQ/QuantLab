package indicators

import (
	"fmt"
	"math"
	"time"

	domainInd "github.com/agoXQ/QuantLab/app/formula/domain/indicators"
	"github.com/agoXQ/QuantLab/app/formula/domain/series"
)

// abs implements ABS(x): element-wise absolute value.
func abs(args []domainInd.Arg) (series.Series, error) {
	if err := requireArgs("ABS", len(args), 1); err != nil {
		return series.Series{}, err
	}
	anchor, ok := firstSeries(args)
	if !ok {
		// Allow ABS over a scalar argument by wrapping in a 1-point series.
		anchor = series.FromValues([]float64{0})
	}
	s, err := asSeries("ABS", 0, args[0], anchor)
	if err != nil {
		return series.Series{}, err
	}
	return s.Map(math.Abs), nil
}

// pairwise applies fn to two aligned series and returns a new series.
func pairwise(name string, a, b series.Series, fn func(x, y float64) float64) (series.Series, error) {
	if !a.AlignedWith(b) {
		// Broadcast a constant series produced via Repeat so the user can
		// write MAX(CLOSE, 0) without manually aligning timestamps.
		if a.Len() == b.Len() {
			out := make([]float64, a.Len())
			for i := 0; i < a.Len(); i++ {
				out[i] = applyPair(a.At(i), b.At(i), fn)
			}
			return series.NewSeries(a.Timestamps(), out)
		}
		return series.Series{}, fmtAlignError(name)
	}
	out := make([]float64, a.Len())
	for i := 0; i < a.Len(); i++ {
		out[i] = applyPair(a.At(i), b.At(i), fn)
	}
	return series.NewSeries(a.Timestamps(), out)
}

func applyPair(x, y float64, fn func(x, y float64) float64) float64 {
	if math.IsNaN(x) || math.IsNaN(y) {
		return math.NaN()
	}
	return fn(x, y)
}

// fmtAlignError reports a misaligned-series condition. Indicators surface
// these via the same code path as bad argument counts so callers can rely on
// a single error category.
func fmtAlignError(name string) error {
	return fmt.Errorf("%s: series arguments must be aligned", name)
}

// maxIndicator implements MAX(a, b).
func maxIndicator(args []domainInd.Arg) (series.Series, error) {
	if err := requireArgs("MAX", len(args), 2); err != nil {
		return series.Series{}, err
	}
	anchor, ok := firstSeries(args)
	if !ok {
		anchor = series.MustNewSeries([]time.Time{time.Time{}}, []float64{0})
	}
	a, err := asSeries("MAX", 0, args[0], anchor)
	if err != nil {
		return series.Series{}, err
	}
	b, err := asSeries("MAX", 1, args[1], anchor)
	if err != nil {
		return series.Series{}, err
	}
	return pairwise("MAX", a, b, math.Max)
}

// minIndicator implements MIN(a, b).
func minIndicator(args []domainInd.Arg) (series.Series, error) {
	if err := requireArgs("MIN", len(args), 2); err != nil {
		return series.Series{}, err
	}
	anchor, ok := firstSeries(args)
	if !ok {
		anchor = series.MustNewSeries([]time.Time{time.Time{}}, []float64{0})
	}
	a, err := asSeries("MIN", 0, args[0], anchor)
	if err != nil {
		return series.Series{}, err
	}
	b, err := asSeries("MIN", 1, args[1], anchor)
	if err != nil {
		return series.Series{}, err
	}
	return pairwise("MIN", a, b, math.Min)
}
