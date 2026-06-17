package errors

import "fmt"

// FormulaError represents a domain-level error with a numeric code and message.
type FormulaError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Pos     int    `json:"pos,omitempty"`
}

func (e *FormulaError) Error() string {
	if e.Pos > 0 {
		return fmt.Sprintf("[%d] %s (pos:%d)", e.Code, e.Message, e.Pos)
	}
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

// DSL-spec error codes (1000-1005).
const (
	ErrSyntaxError     = 1000
	ErrUnknownVariable = 1001
	ErrUnknownFunction = 1002
	ErrInvalidArgCount = 1003
	ErrTypeError       = 1004
	ErrDivisionByZero  = 1005
)

// Extended error codes (1006+).
const (
	ErrFutureFunction = 1006
	ErrLexerError     = 1007
	ErrParseError     = 1008
)

// ErrorCodeMessage returns a human-readable message for a given error code.
func ErrorCodeMessage(code int) string {
	switch code {
	case ErrSyntaxError:
		return "syntax_error"
	case ErrUnknownVariable:
		return "unknown_variable"
	case ErrUnknownFunction:
		return "unknown_function"
	case ErrInvalidArgCount:
		return "invalid_argument_count"
	case ErrTypeError:
		return "type_error"
	case ErrDivisionByZero:
		return "division_by_zero"
	case ErrFutureFunction:
		return "future_function"
	case ErrLexerError:
		return "lexer_error"
	case ErrParseError:
		return "parse_error"
	default:
		return "unknown_error"
	}
}

// NewError creates a new FormulaError.
func NewError(code int, message string) *FormulaError {
	return &FormulaError{Code: code, Message: message}
}

// NewErrorWithPos creates a new FormulaError with position information.
func NewErrorWithPos(code int, message string, pos int) *FormulaError {
	return &FormulaError{Code: code, Message: message, Pos: pos}
}
