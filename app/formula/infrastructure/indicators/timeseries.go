package indicators

import (
	"math"

	domainInd "github.com/agoXQ/QuantLab/app/formula/domain/indicators"
	"github.com/agoXQ/QuantLab/app/formula/domain/series"
)

// ref implements REF(series, n): the value of `series` n bars earlier.
// REF(s, 0) returns s. Negative n is rejected by the validator.
func ref(args []domainInd.Arg) (series.Series, error) {
	if err := requireArgs("REF", len(args), 2); err != nil {
		return series.Series{}, err
	}
	anchor, _ := firstSeries(args)
	s, err := asSeries("REF", 0, args[0], anchor)
	if err != nil {
		return series.Series{}, err
	}
	n, err := asInt("REF", 1, args[1])
	if err != nil {
		// REF(x, 0) is a legal usage so allow zero through asInt's positive
		// guard by upgrading the message instead of bailing here.
		// Implementation note: asInt returns an error for n=0 because most
		// indicators require a positive period. REF treats 0 as identity.
		if n0, ok := tryZeroPeriod(args[1]); ok {
			n = n0
		} else {
			return series.Series{}, err
		}
	}

	out := make([]float64, s.Len())
	for i := 0; i < s.Len(); i++ {
		if i-n < 0 || i-n >= s.Len() {
			out[i] = math.NaN()
			continue
		}
		out[i] = s.At(i - n)
	}
	r, _ := series.NewSeries(s.Timestamps(), out)
	return r, nil
}

func tryZeroPeriod(a domainInd.Arg) (int, bool) {
	if a.Kind != domainInd.ArgKindNumber {
		return 0, false
	}
	if a.Number == 0 {
		return 0, true
	}
	return 0, false
}

// barslast implements BARSLAST(condition): number of bars since condition was
// last true. Output is NaN until at least one true sample is observed.
func barslast(args []domainInd.Arg) (series.Series, error) {
	if err := requireArgs("BARSLAST", len(args), 1); err != nil {
		return series.Series{}, err
	}
	anchor, _ := firstSeries(args)
	cond, err := asSeries("BARSLAST", 0, args[0], anchor)
	if err != nil {
		return series.Series{}, err
	}

	out := make([]float64, cond.Len())
	last := -1
	for i := 0; i < cond.Len(); i++ {
		v := cond.At(i)
		if !math.IsNaN(v) && v != 0 {
			last = i
		}
		if last < 0 {
			out[i] = math.NaN()
		} else {
			out[i] = float64(i - last)
		}
	}
	r, _ := series.NewSeries(cond.Timestamps(), out)
	return r, nil
}
