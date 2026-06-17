// Package indicators defines the contract for technical indicator
// implementations consumed by the Formula Engine evaluator.
//
// The contract is intentionally framed as a Library lookup so that the
// evaluator stays decoupled from any concrete indicator implementation.
// Indicators are pure functions over a Series; they neither read state nor
// touch I/O.
package indicators

import (
	"github.com/agoXQ/QuantLab/app/formula/domain/series"
)

// Indicator is the canonical signature for built-in technical indicators.
//
// args carries the positional, already-evaluated arguments passed by the
// formula. Series arguments are wrapped as series.Series; numeric arguments
// remain as float64. Implementations must validate their argument shapes and
// return a typed error on mismatch.
//
// The returned Series must be aligned with the input Series (same timestamps).
// Indicators that produce multiple lines (e.g. MACD, KDJ, BOLL) should return
// the line that callers reference by name; the evaluator dispatches function
// names like "MACD_DIF", "KDJ_K" through the registry.
type Indicator func(args []Arg) (series.Series, error)

// Arg is a tagged union for indicator arguments.
//
// Using an explicit struct rather than interface{} keeps allocations down on
// the hot path and makes type errors obvious.
type Arg struct {
	Kind   ArgKind
	Series series.Series
	Number float64
	Bool   bool
}

// ArgKind enumerates the supported argument shapes.
type ArgKind int

const (
	// ArgKindUnknown is the zero value and signals a misuse.
	ArgKindUnknown ArgKind = iota
	// ArgKindSeries means Series carries the input.
	ArgKindSeries
	// ArgKindNumber means Number carries the input.
	ArgKindNumber
	// ArgKindBool means Bool carries the input.
	ArgKindBool
)

// SeriesArg is a convenience constructor.
func SeriesArg(s series.Series) Arg { return Arg{Kind: ArgKindSeries, Series: s} }

// NumberArg is a convenience constructor.
func NumberArg(v float64) Arg { return Arg{Kind: ArgKindNumber, Number: v} }

// BoolArg is a convenience constructor.
func BoolArg(v bool) Arg { return Arg{Kind: ArgKindBool, Bool: v} }

// Library is the lookup surface for indicator implementations.
//
// The evaluator resolves built-in function calls through Library.Get; if the
// canonical name is not registered the evaluator returns ErrUnknownFunction
// (mapped to the formula error code 1002).
type Library interface {
	Get(name string) (Indicator, bool)
}
