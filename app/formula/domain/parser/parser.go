package parser

import (
	"context"

	"github.com/agoXQ/QuantLab/app/formula/domain/ast"
	"github.com/agoXQ/QuantLab/app/formula/domain/lexer"
)

// Parser defines the interface for parsing tokens into an AST.
// Implementations must be safe for concurrent use.
type Parser interface {
	// Parse converts a slice of tokens into an AST node.
	Parse(ctx context.Context, tokens []lexer.Token) (ast.Node, error)
}
