package evaluator

import (
	"fmt"
	"math"

	"github.com/agoXQ/QuantLab/app/formula/domain/series"
)

// valueKind enumerates the three runtime shapes the evaluator works with.
type valueKind int

const (
	valueKindUnknown valueKind = iota
	valueKindNumber
	valueKindBool
	valueKindSeries
)

// value is the per-stock intermediate result produced by traversing the AST.
//
// Booleans and Numbers are kept as scalars to avoid allocating constant
// series for every compare/arithmetic step. Series is allocated only when an
// indicator or time-series identifier (CLOSE, HIGH, ...) is referenced.
type value struct {
	kind  valueKind
	num   float64
	bul   bool
	ser   series.Series
}

func numberValue(n float64) value { return value{kind: valueKindNumber, num: n} }
func boolValue(b bool) value      { return value{kind: valueKindBool, bul: b} }
func seriesValue(s series.Series) value {
	return value{kind: valueKindSeries, ser: s}
}

// asNumber materialises the final scalar reading of v, looking at the last
// sample of any series. NaN propagates.
func (v value) asNumber() float64 {
	switch v.kind {
	case valueKindNumber:
		return v.num
	case valueKindBool:
		if v.bul {
			return 1
		}
		return 0
	case valueKindSeries:
		return v.ser.Last()
	default:
		return math.NaN()
	}
}

// asBool returns the boolean projection of v. For series, the last bar's
// value is interpreted as truthy iff it is not NaN and not zero.
func (v value) asBool() bool {
	switch v.kind {
	case valueKindBool:
		return v.bul
	case valueKindNumber:
		return !math.IsNaN(v.num) && v.num != 0
	case valueKindSeries:
		last := v.ser.Last()
		return !math.IsNaN(last) && last != 0
	default:
		return false
	}
}

// String aids debugging.
func (v value) String() string {
	switch v.kind {
	case valueKindNumber:
		return fmt.Sprintf("Number(%v)", v.num)
	case valueKindBool:
		return fmt.Sprintf("Bool(%v)", v.bul)
	case valueKindSeries:
		return fmt.Sprintf("Series(len=%d, last=%v)", v.ser.Len(), v.ser.Last())
	default:
		return "Unknown"
	}
}
