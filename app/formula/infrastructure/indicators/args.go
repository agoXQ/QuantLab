package indicators

import (
	"errors"
	"fmt"
	"math"

	domainInd "github.com/agoXQ/QuantLab/app/formula/domain/indicators"
	"github.com/agoXQ/QuantLab/app/formula/domain/series"
)

// errArgCount is returned when the caller supplies too few or too many args.
var errArgCount = errors.New("indicator: invalid argument count")

// requireArgs validates the number of arguments passed to an indicator.
// It produces an error that the evaluator translates to ErrInvalidArgCount.
func requireArgs(name string, got, want int) error {
	if got != want {
		return fmt.Errorf("%w: %s expects %d args, got %d", errArgCount, name, want, got)
	}
	return nil
}

// requireArgsBetween validates that got is in the inclusive [min, max] range.
func requireArgsBetween(name string, got, min, max int) error {
	if got < min || got > max {
		return fmt.Errorf("%w: %s expects %d-%d args, got %d", errArgCount, name, min, max, got)
	}
	return nil
}

// asSeries unwraps a Series argument or fails with a typed error. A Number
// argument is broadcast to a constant series aligned with anchor.
func asSeries(name string, idx int, a domainInd.Arg, anchor series.Series) (series.Series, error) {
	switch a.Kind {
	case domainInd.ArgKindSeries:
		return a.Series, nil
	case domainInd.ArgKindNumber:
		return series.Repeat(anchor.Timestamps(), a.Number), nil
	case domainInd.ArgKindBool:
		v := 0.0
		if a.Bool {
			v = 1.0
		}
		return series.Repeat(anchor.Timestamps(), v), nil
	default:
		return series.Series{}, fmt.Errorf("%s arg %d: expected Series, got unknown", name, idx+1)
	}
}

// asInt unwraps a positive integer period argument.
func asInt(name string, idx int, a domainInd.Arg) (int, error) {
	if a.Kind != domainInd.ArgKindNumber {
		return 0, fmt.Errorf("%s arg %d: expected Number period, got non-number", name, idx+1)
	}
	if math.IsNaN(a.Number) || math.IsInf(a.Number, 0) {
		return 0, fmt.Errorf("%s arg %d: period must be finite", name, idx+1)
	}
	n := int(a.Number)
	if float64(n) != a.Number {
		return 0, fmt.Errorf("%s arg %d: period must be an integer, got %v", name, idx+1, a.Number)
	}
	if n <= 0 {
		return 0, fmt.Errorf("%s arg %d: period must be positive, got %d", name, idx+1, n)
	}
	return n, nil
}

// asFloat unwraps a numeric argument (period coefficient, etc.).
func asFloat(name string, idx int, a domainInd.Arg) (float64, error) {
	if a.Kind != domainInd.ArgKindNumber {
		return 0, fmt.Errorf("%s arg %d: expected Number, got non-number", name, idx+1)
	}
	return a.Number, nil
}

// firstSeries finds the first Series argument and returns it. Used to align
// constant Number arguments with the prevailing time index.
func firstSeries(args []domainInd.Arg) (series.Series, bool) {
	for _, a := range args {
		if a.Kind == domainInd.ArgKindSeries {
			return a.Series, true
		}
	}
	return series.Series{}, false
}
