// Package series defines the Series and Bar value objects used by the
// Formula Engine evaluator and indicator library.
//
// A Series is an ordered, equal-length pair of timestamps and float64 values.
// NaN is used to mark missing or undefined samples, matching the convention
// adopted by most quant DSLs (TDX, MyQuant, AkShare, etc.).
package series

import (
	"errors"
	"math"
	"time"
)

// Series is an immutable, time-aligned float64 sequence.
//
// Implementations must not be mutated after construction; helper methods
// always return new Series instances.
type Series struct {
	timestamps []time.Time
	values     []float64
}

// NewSeries constructs a Series from parallel timestamps and values slices.
// The two slices must have equal length.
func NewSeries(timestamps []time.Time, values []float64) (Series, error) {
	if len(timestamps) != len(values) {
		return Series{}, errors.New("series: timestamps and values length mismatch")
	}
	// Defensive copies preserve immutability semantics.
	ts := make([]time.Time, len(timestamps))
	copy(ts, timestamps)
	vs := make([]float64, len(values))
	copy(vs, values)
	return Series{timestamps: ts, values: vs}, nil
}

// MustNewSeries panics on length mismatch. Useful for tests and constants.
func MustNewSeries(timestamps []time.Time, values []float64) Series {
	s, err := NewSeries(timestamps, values)
	if err != nil {
		panic(err)
	}
	return s
}

// FromValues creates a Series with synthetic, evenly-spaced timestamps. It is
// intended for indicator unit tests where the actual dates do not matter.
func FromValues(values []float64) Series {
	ts := make([]time.Time, len(values))
	base := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	for i := range values {
		ts[i] = base.AddDate(0, 0, i)
	}
	return Series{timestamps: ts, values: append([]float64(nil), values...)}
}

// FromBars projects a slice of bars onto a single field. The provided fn maps
// each bar to a float64 value.
func FromBars(bars []Bar, fn func(Bar) float64) Series {
	ts := make([]time.Time, len(bars))
	vs := make([]float64, len(bars))
	for i, b := range bars {
		ts[i] = b.Timestamp
		vs[i] = fn(b)
	}
	return Series{timestamps: ts, values: vs}
}

// Repeat returns a constant Series aligned with the given timestamps.
func Repeat(timestamps []time.Time, value float64) Series {
	ts := make([]time.Time, len(timestamps))
	copy(ts, timestamps)
	vs := make([]float64, len(timestamps))
	for i := range vs {
		vs[i] = value
	}
	return Series{timestamps: ts, values: vs}
}

// Len returns the number of samples.
func (s Series) Len() int { return len(s.values) }

// Empty reports whether the series has no samples.
func (s Series) Empty() bool { return len(s.values) == 0 }

// Values returns a copy of the underlying values slice.
func (s Series) Values() []float64 {
	out := make([]float64, len(s.values))
	copy(out, s.values)
	return out
}

// Timestamps returns a copy of the underlying timestamps slice.
func (s Series) Timestamps() []time.Time {
	out := make([]time.Time, len(s.timestamps))
	copy(out, s.timestamps)
	return out
}

// At returns the value at index i.
func (s Series) At(i int) float64 { return s.values[i] }

// TimeAt returns the timestamp at index i.
func (s Series) TimeAt(i int) time.Time { return s.timestamps[i] }

// Last returns the value at the final index. Returns NaN for empty series.
func (s Series) Last() float64 {
	if len(s.values) == 0 {
		return math.NaN()
	}
	return s.values[len(s.values)-1]
}

// IsNaN reports whether the value at i is NaN.
func (s Series) IsNaN(i int) bool { return math.IsNaN(s.values[i]) }

// Map returns a new Series whose values are produced element-wise by fn.
// NaN inputs are preserved unless fn explicitly handles them.
func (s Series) Map(fn func(float64) float64) Series {
	vs := make([]float64, len(s.values))
	for i, v := range s.values {
		if math.IsNaN(v) {
			vs[i] = math.NaN()
			continue
		}
		vs[i] = fn(v)
	}
	ts := make([]time.Time, len(s.timestamps))
	copy(ts, s.timestamps)
	return Series{timestamps: ts, values: vs}
}

// AlignedWith reports whether two series share the same timestamps.
//
// Used by the evaluator to fail fast when arithmetic combines series produced
// from different universes or ranges.
func (s Series) AlignedWith(other Series) bool {
	if len(s.timestamps) != len(other.timestamps) {
		return false
	}
	for i := range s.timestamps {
		if !s.timestamps[i].Equal(other.timestamps[i]) {
			return false
		}
	}
	return true
}

// NaN is the canonical missing-value marker used by indicators.
func NaN() float64 { return math.NaN() }
