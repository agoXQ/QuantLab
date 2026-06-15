// Package errors defines the unified error code system for all QuantLab services.
//
// Error code ranges:
//
//	1xxxx — Validation errors
//	2xxxx — Business errors
//	3xxxx — Permission errors
//	4xxxx — Resource errors
//	5xxxx — System errors
package errors

import "fmt"

// AppError is the unified application error.
type AppError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *AppError) Error() string {
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

// New creates a new AppError.
func New(code int, message string) *AppError {
	return &AppError{Code: code, Message: message}
}

// Newf creates a new AppError with a formatted message.
func Newf(code int, format string, args ...interface{}) *AppError {
	return &AppError{Code: code, Message: fmt.Sprintf(format, args...)}
}

// ============================================================
// 1xxxx — Validation
// ============================================================

var (
	ErrInvalidParam     = New(10001, "invalid parameter")
	ErrMissingRequired  = New(10002, "missing required field")
	ErrInvalidFormat    = New(10003, "invalid format")
	ErrValueOutOfRange  = New(10004, "value out of range")
)

// ============================================================
// 2xxxx — Business
// ============================================================

var (
	ErrEmailAlreadyExists    = New(20001, "email already exists")
	ErrUsernameAlreadyExists = New(20002, "username already exists")
	ErrSubscriptionExpired   = New(20003, "subscription expired")
	ErrQuotaExceeded         = New(20004, "quota exceeded")
	ErrInvalidCredentials    = New(20005, "invalid credentials")
	ErrStrategyNotBacktested = New(20006, "strategy has not been backtested")
	ErrDuplicateOperation    = New(20007, "duplicate operation")
)

// ============================================================
// 3xxxx — Permission
// ============================================================

var (
	ErrUnauthorized     = New(30001, "unauthorized")
	ErrForbidden        = New(30002, "forbidden")
	ErrAccountBanned    = New(30003, "account banned")
	ErrMembershipExpired = New(30004, "membership expired")
)

// ============================================================
// 4xxxx — Resource
// ============================================================

var (
	ErrNotFound         = New(40001, "resource not found")
	ErrUserNotFound     = New(40002, "user not found")
	ErrStrategyNotFound = New(40003, "strategy not found")
	ErrPortfolioNotFound = New(40004, "portfolio not found")
	ErrBacktestNotFound  = New(40005, "backtest not found")
	ErrOrderNotFound     = New(40006, "order not found")
)

// ============================================================
// 5xxxx — System
// ============================================================

var (
	ErrInternal      = New(50001, "internal server error")
	ErrDatabaseError = New(50002, "database error")
	ErrCacheError    = New(50003, "cache error")
	ErrKafkaError    = New(50004, "message queue error")
	ErrTimeout       = New(50005, "request timeout")
)
