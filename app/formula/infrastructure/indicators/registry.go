// Package indicators provides built-in technical indicator implementations
// for the Formula Engine evaluator.
//
// Each indicator is a pure function over a Series and returns a Series of
// the same length. Lookback periods that are not yet satisfied are filled
// with NaN, matching the behaviour of TDX/通达信公式 and most Python quant
// stacks.
package indicators

import (
	"strings"
	"sync"

	domainInd "github.com/agoXQ/QuantLab/app/formula/domain/indicators"
)

// library implements the domain Library interface with a static map.
type library struct {
	mu  sync.RWMutex
	fns map[string]domainInd.Indicator // canonical lowercase name -> impl
}

// NewLibrary builds a Library populated with the MVP indicator set:
//
//   Math/series:    ABS, MAX, MIN, SUM, AVG, COUNT
//   Technical:      MA, EMA, SMA, STD, RSI, MACD_DIF, MACD_DEA, MACD,
//                   KDJ_K, KDJ_D, KDJ_J, BOLL_MID, BOLL_UP, BOLL_DOWN, ATR
//   Time-series:    REF, HHV, LLV, BARSLAST
//   Signals:        CROSS, LONGCROSS, FILTER
//
// MACD/KDJ/BOLL also expose a primary alias (MACD, KDJ, BOLL) that returns
// the principal line so callers can write MACD(CLOSE, 12, 26, 9) > 0.
func NewLibrary() domainInd.Library {
	l := &library{fns: make(map[string]domainInd.Indicator, 32)}
	l.registerBuiltins()
	return l
}

func (l *library) Get(name string) (domainInd.Indicator, bool) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	fn, ok := l.fns[strings.ToLower(name)]
	return fn, ok
}

func (l *library) register(name string, fn domainInd.Indicator) {
	l.fns[strings.ToLower(name)] = fn
}

func (l *library) registerBuiltins() {
	// Math / element-wise
	l.register("ABS", abs)
	l.register("MAX", maxIndicator)
	l.register("MIN", minIndicator)

	// Rolling stats
	l.register("SUM", sum)
	l.register("AVG", avg)
	l.register("COUNT", count)

	// Moving averages
	l.register("MA", ma)
	l.register("EMA", ema)
	l.register("SMA", sma)
	l.register("STD", std)

	// Range stats
	l.register("HHV", hhv)
	l.register("LLV", llv)

	// Momentum / oscillators
	l.register("RSI", rsi)
	l.register("MACD", macdDIF)
	l.register("MACD_DIF", macdDIF)
	l.register("MACD_DEA", macdDEA)
	l.register("MACD_HIST", macdHist)
	l.register("KDJ", kdjK)
	l.register("KDJ_K", kdjK)
	l.register("KDJ_D", kdjD)
	l.register("KDJ_J", kdjJ)
	l.register("BOLL", bollMid)
	l.register("BOLL_MID", bollMid)
	l.register("BOLL_UP", bollUp)
	l.register("BOLL_DOWN", bollDown)
	l.register("ATR", atr)

	// Time series
	l.register("REF", ref)
	l.register("BARSLAST", barslast)

	// Signals
	l.register("CROSS", cross)
	l.register("LONGCROSS", longcross)
	l.register("FILTER", filterSignal)
}
