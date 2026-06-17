package evaluator

import (
	"strings"
	"time"

	"github.com/agoXQ/QuantLab/app/formula/domain/series"
)

// stockContext bundles all per-stock data the evaluator needs to walk an AST
// without going back to the data port mid-traversal.
type stockContext struct {
	stockCode string
	bars      []series.Bar
	// Cached series lazily derived from bars.
	open, high, low, close, vol, amount series.Series
	// Latest financial scalars (PE, PB, ROE, ...). Missing keys fall back to
	// NaN at lookup time.
	financials map[string]float64
	// Local user-defined identifiers produced by Assignment statements.
	locals map[string]value
}

func newStockContext(code string, bars []series.Bar, fins map[string]float64) *stockContext {
	c := &stockContext{
		stockCode:  code,
		bars:       bars,
		financials: fins,
		locals:     make(map[string]value),
	}
	if fins == nil {
		c.financials = map[string]float64{}
	}
	c.open = series.FromBars(bars, func(b series.Bar) float64 { return b.Open })
	c.high = series.FromBars(bars, func(b series.Bar) float64 { return b.High })
	c.low = series.FromBars(bars, func(b series.Bar) float64 { return b.Low })
	c.close = series.FromBars(bars, func(b series.Bar) float64 { return b.Close })
	c.vol = series.FromBars(bars, func(b series.Bar) float64 { return b.Volume })
	c.amount = series.FromBars(bars, func(b series.Bar) float64 { return b.Amount })
	return c
}

// timestamps returns the bar timestamp axis. Used to broadcast scalar
// constants (like financial metrics) into a series-aligned indicator input.
func (c *stockContext) timestamps() []time.Time {
	out := make([]time.Time, len(c.bars))
	for i, b := range c.bars {
		out[i] = b.Timestamp
	}
	return out
}

// resolveBuiltinSeries returns the canonical OHLCV series matching the given
// identifier name (case-insensitive). The boolean reports whether name is a
// known market data identifier.
func (c *stockContext) resolveBuiltinSeries(name string) (series.Series, bool) {
	switch strings.ToUpper(name) {
	case "OPEN", "O":
		return c.open, true
	case "HIGH", "H":
		return c.high, true
	case "LOW", "L":
		return c.low, true
	case "CLOSE", "C":
		return c.close, true
	case "VOL", "VOLUME", "V":
		return c.vol, true
	case "AMOUNT":
		return c.amount, true
	default:
		return series.Series{}, false
	}
}

// resolveFinancial returns a scalar financial metric. NaN is returned when
// the metric is unknown so the evaluator can still produce a result instead
// of failing the whole universe.
func (c *stockContext) resolveFinancial(name string) (float64, bool) {
	if v, ok := c.financials[strings.ToUpper(name)]; ok {
		return v, true
	}
	if v, ok := c.financials[name]; ok {
		return v, true
	}
	return 0, false
}
