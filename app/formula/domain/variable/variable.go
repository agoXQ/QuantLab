package variable

import "context"

// VariableType represents the data type of a variable.
type VariableType string

const (
	TypeSeries VariableType = "Series"
	TypeNumber VariableType = "Number"
)

// VariableDefinition defines a known variable in the DSL.
type VariableDefinition struct {
	Name        string       `json:"name"`
	VarType     VariableType `json:"var_type"`
	Category    string       `json:"category"`
	Description string       `json:"description"`
}

// Registry defines the interface for looking up known variables.
// Implementations must be safe for concurrent use.
type Registry interface {
	// GetVariable returns the variable definition for the given name (case-insensitive).
	GetVariable(name string) (VariableDefinition, bool)

	// ListVariables returns all registered variable definitions.
	ListVariables() []VariableDefinition

	// Exists checks if a variable with the given name exists (case-insensitive).
	Exists(name string) bool
}

// Unused context import placeholder.
var _ = context.Background
