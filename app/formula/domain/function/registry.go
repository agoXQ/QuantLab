package function

import "context"

// Registry defines the interface for looking up and managing functions.
// Implementations must be safe for concurrent use.
type Registry interface {
	// GetFunction returns the function definition for the given name.
	// The lookup is case-insensitive.
	GetFunction(name string) (FunctionDefinition, bool)

	// ListFunctions returns all registered function definitions.
	ListFunctions() []FunctionDefinition

	// RegisterFunction adds a new function definition to the registry.
	RegisterFunction(def FunctionDefinition) error

	// Exists checks if a function with the given name exists (case-insensitive).
	Exists(name string) bool

	// ResolveName returns the canonical name for a function (case-insensitive lookup).
	ResolveName(name string) (string, bool)
}

// Ensure Registry is usable with context in future extensions.
var _ interface{ RegisterFunction(FunctionDefinition) error } = (Registry)(nil)

// Unused context import placeholder.
var _ = context.Background
