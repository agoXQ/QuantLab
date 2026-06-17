package validator

import (
	"context"

	"github.com/agoXQ/QuantLab/app/formula/domain/ast"
)

// Severity represents the severity level of a validation error.
type Severity string

const (
	SeverityError   Severity = "ERROR"
	SeverityWarning Severity = "WARNING"
)

// ValidationError represents a single validation error or warning.
type ValidationError struct {
	Code     int      `json:"code"`
	CodeStr  string   `json:"code_str"`
	Message  string   `json:"message"`
	Severity Severity `json:"severity"`
	Pos      int      `json:"pos,omitempty"`
}

// ValidationResult contains the result of validating an AST.
type ValidationResult struct {
	Valid  bool              `json:"valid"`
	Errors []ValidationError `json:"errors,omitempty"`
}

// Validator defines the interface for validating AST nodes.
// Implementations must be safe for concurrent use.
type Validator interface {
	// Validate performs semantic validation on an AST node.
	Validate(ctx context.Context, node ast.Node) (*ValidationResult, error)
}
