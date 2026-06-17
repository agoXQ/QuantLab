package evaluator

import (
	"context"

	domainAST "github.com/agoXQ/QuantLab/app/formula/domain/ast"
	domainEval "github.com/agoXQ/QuantLab/app/formula/domain/evaluator"
	domainVar "github.com/agoXQ/QuantLab/app/formula/domain/variable"
)

// dataPortKey is the unexported key used to plumb the DataPort through the
// request context. Pinning it to the package keeps callers from accidentally
// shadowing the value with a same-named string key.
type dataPortKey struct{}

// WithDataPort returns a derived context that carries the given port. The
// evaluator looks the port up via dataPortFromContext.
//
// The port is request-scoped so a single Service can switch between live
// market data and a test fake without reconfiguration.
func WithDataPort(ctx context.Context, port domainEval.DataPort) context.Context {
	return context.WithValue(ctx, dataPortKey{}, port)
}

func dataPortFromContext(ctx context.Context) (domainEval.DataPort, bool) {
	v, ok := ctx.Value(dataPortKey{}).(domainEval.DataPort)
	return v, ok
}

// collectFinancialIdentifiers walks the AST and returns the canonical names
// of every Number-typed (financial) variable referenced. The evaluator uses
// the result to issue a single batched LoadFinancialsLatest call rather than
// making one round trip per identifier.
func collectFinancialIdentifiers(node domainAST.Node, vars domainVar.Registry) []string {
	if node == nil {
		return nil
	}
	seen := map[string]struct{}{}
	walkAST(node, func(n domainAST.Node) {
		ident, ok := n.(*domainAST.Identifier)
		if !ok {
			return
		}
		def, found := vars.GetVariable(ident.Name)
		if !found {
			return
		}
		if def.VarType != domainVar.TypeNumber {
			return
		}
		seen[def.Name] = struct{}{}
	})
	if len(seen) == 0 {
		return nil
	}
	out := make([]string, 0, len(seen))
	for k := range seen {
		out = append(out, k)
	}
	return out
}

func walkAST(node domainAST.Node, visit func(domainAST.Node)) {
	if node == nil {
		return
	}
	visit(node)
	switch n := node.(type) {
	case *domainAST.BinaryExpression:
		walkAST(n.Left, visit)
		walkAST(n.Right, visit)
	case *domainAST.UnaryExpression:
		walkAST(n.Operand, visit)
	case *domainAST.FunctionCall:
		for _, a := range n.Args {
			walkAST(a, visit)
		}
	case *domainAST.Assignment:
		walkAST(n.Value, visit)
	case *domainAST.Program:
		for _, s := range n.Statements {
			walkAST(s, visit)
		}
	}
}
