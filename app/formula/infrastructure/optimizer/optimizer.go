package optimizer

import (
	"context"
	"strings"

	domainAST "github.com/agoXQ/QuantLab/app/formula/domain/ast"
	domainOptimizer "github.com/agoXQ/QuantLab/app/formula/domain/optimizer"
)

type optimizer struct{}

// NewOptimizer creates a new AST optimizer instance.
func NewOptimizer() domainOptimizer.Optimizer {
	return &optimizer{}
}

func (o *optimizer) Optimize(_ context.Context, node domainAST.Node) (domainAST.Node, error) {
	return o.optimize(node), nil
}

func (o *optimizer) optimize(node domainAST.Node) domainAST.Node {
	switch n := node.(type) {
	case *domainAST.BinaryExpression:
		return o.optimizeBinary(n)
	case *domainAST.UnaryExpression:
		return o.optimizeUnary(n)
	case *domainAST.FunctionCall:
		return o.optimizeFunctionCall(n)
	case *domainAST.Assignment:
		return &domainAST.Assignment{Name: n.Name, Value: o.optimize(n.Value)}
	case *domainAST.Program:
		stmts := make([]domainAST.Node, len(n.Statements))
		for i, stmt := range n.Statements {
			stmts[i] = o.optimize(stmt)
		}
		return &domainAST.Program{Statements: stmts}
	default:
		return node
	}
}

func (o *optimizer) optimizeBinary(expr *domainAST.BinaryExpression) domainAST.Node {
	left := o.optimize(expr.Left)
	right := o.optimize(expr.Right)
	op := strings.ToUpper(expr.Operator)

	if numLeft, ok := left.(*domainAST.NumberLiteral); ok {
		if numRight, ok := right.(*domainAST.NumberLiteral); ok {
			if folded := o.foldArithmetic(numLeft, numRight, expr.Operator); folded != nil {
				return folded
			}
		}
	}

	if boolLeft, ok := left.(*domainAST.BoolLiteral); ok {
		if boolRight, ok := right.(*domainAST.BoolLiteral); ok {
			if folded := o.foldBoolean(boolLeft, boolRight, op); folded != nil {
				return folded
			}
		}
		if op == "AND" {
			if boolLeft.Value {
				return right
			}
			return &domainAST.BoolLiteral{Value: false}
		}
		if op == "OR" {
			if boolLeft.Value {
				return &domainAST.BoolLiteral{Value: true}
			}
			return right
		}
	}

	if boolRight, ok := right.(*domainAST.BoolLiteral); ok {
		if op == "AND" {
			if boolRight.Value {
				return left
			}
			return &domainAST.BoolLiteral{Value: false}
		}
		if op == "OR" {
			if boolRight.Value {
				return &domainAST.BoolLiteral{Value: true}
			}
			return left
		}
	}

	return &domainAST.BinaryExpression{Left: left, Operator: expr.Operator, Right: right}
}

func (o *optimizer) optimizeUnary(expr *domainAST.UnaryExpression) domainAST.Node {
	operand := o.optimize(expr.Operand)

	if strings.ToUpper(expr.Operator) == "NOT" {
		if unary, ok := operand.(*domainAST.UnaryExpression); ok && strings.ToUpper(unary.Operator) == "NOT" {
			return unary.Operand
		}
	}

	if boolLit, ok := operand.(*domainAST.BoolLiteral); ok {
		return &domainAST.BoolLiteral{Value: !boolLit.Value}
	}

	return &domainAST.UnaryExpression{Operator: expr.Operator, Operand: operand}
}

func (o *optimizer) optimizeFunctionCall(call *domainAST.FunctionCall) domainAST.Node {
	args := make([]domainAST.Node, len(call.Args))
	for i, arg := range call.Args {
		args[i] = o.optimize(arg)
	}
	return &domainAST.FunctionCall{Name: call.Name, Args: args}
}

func (o *optimizer) foldArithmetic(left, right *domainAST.NumberLiteral, op string) domainAST.Node {
	switch op {
	case "+":
		return &domainAST.NumberLiteral{Value: left.Value + right.Value}
	case "-":
		return &domainAST.NumberLiteral{Value: left.Value - right.Value}
	case "*":
		return &domainAST.NumberLiteral{Value: left.Value * right.Value}
	case "/":
		if right.Value != 0 {
			return &domainAST.NumberLiteral{Value: left.Value / right.Value}
		}
		return nil
	case "%":
		if right.Value != 0 {
			return &domainAST.NumberLiteral{Value: float64(int64(left.Value) % int64(right.Value))}
		}
		return nil
	default:
		return nil
	}
}

func (o *optimizer) foldBoolean(left, right *domainAST.BoolLiteral, op string) domainAST.Node {
	switch op {
	case "AND":
		return &domainAST.BoolLiteral{Value: left.Value && right.Value}
	case "OR":
		return &domainAST.BoolLiteral{Value: left.Value || right.Value}
	default:
		return nil
	}
}
