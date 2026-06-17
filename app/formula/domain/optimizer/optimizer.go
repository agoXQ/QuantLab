package optimizer

import (
	"context"

	"github.com/agoXQ/QuantLab/app/formula/domain/ast"
)

// Optimizer defines the interface for AST optimization.
// Implementations must be safe for concurrent use.
type Optimizer interface {
	// Optimize applies optimizations to an AST node and returns the optimized node.
	Optimize(ctx context.Context, node ast.Node) (ast.Node, error)
}
