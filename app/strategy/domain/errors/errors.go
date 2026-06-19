// Package errors defines domain-level errors for the Strategy Service.
//
// Codes follow the platform-wide allocation in pkg/errors. The Strategy
// service reserves 40030-40039 for resource errors, 10030-10039 for
// validation, 20030-20039 for state-machine refusals, and 50030-50039
// for system errors.
package errors

import "fmt"

// StrategyError represents a domain-level error with a numeric code.
type StrategyError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *StrategyError) Error() string {
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

// New creates a StrategyError.
func New(code int, message string) *StrategyError {
	return &StrategyError{Code: code, Message: message}
}

// 4xxxx — Resource errors.
var (
	ErrStrategyNotFound = New(40030, "strategy not found")
	ErrVersionNotFound  = New(40031, "strategy version not found")
)

// 1xxxx — Validation errors.
var (
	ErrInvalidStrategy   = New(10030, "invalid strategy")
	ErrTitleRequired     = New(10031, "title is required")
	ErrTitleTooLong      = New(10032, "title is too long")
	ErrFormulaRequired   = New(10033, "formula text is required")
	ErrInvalidVisibility = New(10034, "invalid visibility")
	ErrInvalidStatus     = New(10035, "invalid lifecycle status")
	ErrInvalidVersion    = New(10036, "invalid strategy version")
)

// 2xxxx — Conflict / state-machine refusals. The HTTP layer renders
// these as 409 Conflict.
var (
	ErrStrategyArchived       = New(20030, "strategy is archived")
	ErrAlreadyPublished       = New(20031, "strategy is already published in this version")
	ErrPublishWithoutVersion  = New(20032, "strategy has no version to publish")
	ErrInvalidStateTransition = New(20033, "invalid lifecycle state transition")
)

// 3xxxx — Permission errors.
var (
	ErrNotOwner = New(30030, "caller is not the strategy owner")
)

// 5xxxx — System errors.
var (
	ErrPersistence = New(50030, "strategy persistence error")
)
