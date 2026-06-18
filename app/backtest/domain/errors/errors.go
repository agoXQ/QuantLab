// Package errors defines domain-level errors for the Backtest Engine.
//
// Codes follow the platform-wide allocation in pkg/errors. The Backtest
// service reserves 40020-40029 for resource errors and 10020-10029 for
// validation errors so we never collide with Market Data (40010-) or the
// Formula Engine.
package errors

import "fmt"

// BacktestError represents a domain-level error with a numeric code.
type BacktestError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *BacktestError) Error() string {
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

// New creates a BacktestError.
func New(code int, message string) *BacktestError {
	return &BacktestError{Code: code, Message: message}
}

// Newf creates a BacktestError with a formatted message.
func Newf(code int, format string, args ...interface{}) *BacktestError {
	return &BacktestError{Code: code, Message: fmt.Sprintf(format, args...)}
}

// 4xxxx — Resource errors.
var (
	ErrJobNotFound    = New(40020, "backtest job not found")
	ErrReportNotFound = New(40021, "backtest report not found")
)

// 1xxxx — Validation errors.
var (
	ErrInvalidJob              = New(10020, "invalid backtest job")
	ErrInvalidConfig           = New(10021, "invalid backtest config")
	ErrInvalidUniverse         = New(10022, "universe is required")
	ErrInvalidFormula          = New(10023, "formula is required")
	ErrInvalidDateRange        = New(10024, "invalid date range")
	ErrInvalidInitialCapital   = New(10025, "initial capital must be positive")
	ErrInvalidRebalanceFreq    = New(10026, "invalid rebalance frequency")
	ErrInvalidStateTransition  = New(10027, "invalid job state transition")
)

// 5xxxx — System errors.
var (
	ErrEvaluatorUnavailable = New(50020, "formula evaluator unavailable")
	ErrMarketDataMissing    = New(50021, "market data missing for backtest window")
)
