package lexer

import "context"

// Lexer defines the interface for tokenizing DSL source code.
// Implementations must be safe for concurrent use.
type Lexer interface {
	// Tokenize converts a DSL formula string into a slice of tokens.
	Tokenize(ctx context.Context, input string) ([]Token, error)
}
