// Package errors defines domain-level errors for the Market Data service.
//
// These errors map to the platform error catalog in pkg/errors but stay
// inside the domain to keep the layer clean and easily testable.
package errors

import "fmt"

// MarketError represents a domain-level error with a numeric code and message.
type MarketError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *MarketError) Error() string {
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

// New creates a MarketError with the given code and message.
func New(code int, message string) *MarketError {
	return &MarketError{Code: code, Message: message}
}

// Newf creates a MarketError with a formatted message.
func Newf(code int, format string, args ...interface{}) *MarketError {
	return &MarketError{Code: code, Message: fmt.Sprintf(format, args...)}
}

// 4xxxx — Resource errors. These mirror the platform-wide allocations
// declared in pkg/errors, with reservations 40010-40019 for market data.
var (
	ErrSecurityNotFound       = New(40010, "security not found")
	ErrCalendarUnavailable    = New(40011, "trading calendar unavailable")
	ErrDataVersionNotFound    = New(40012, "data version not found")
	ErrFinancialReportMissing = New(40013, "financial report missing")
	ErrIndexNotFound          = New(40014, "index not found")
)

// 1xxxx — Validation errors specific to market data inputs.
var (
	ErrInvalidStockCode  = New(10010, "invalid stock code")
	ErrInvalidPeriod     = New(10011, "invalid period")
	ErrInvalidAdjustment = New(10012, "invalid adjustment")
	ErrInvalidDateRange  = New(10013, "invalid date range")
	ErrInvalidReportType = New(10014, "invalid report type")
)

// 5xxxx — System errors for upstream provider failures.
var (
	ErrProviderUnavailable = New(50010, "data provider unavailable")
	ErrProviderRateLimit   = New(50011, "data provider rate limit exceeded")
	ErrProviderResponse    = New(50012, "data provider returned invalid response")
)
