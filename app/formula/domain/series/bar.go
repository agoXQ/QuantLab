package series

import "time"

// Bar represents a single OHLCV row consumed by the evaluator.
//
// It is intentionally decoupled from the Market Data domain so that the
// Formula Engine never imports another bounded context. Adapters from the
// market data port populate this struct.
type Bar struct {
	Timestamp time.Time
	Open      float64
	High      float64
	Low       float64
	Close     float64
	Volume    float64
	Amount    float64
}
