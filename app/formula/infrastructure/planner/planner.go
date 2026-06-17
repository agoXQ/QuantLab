package planner

import (
	"context"
	"strings"

	domainAST "github.com/agoXQ/QuantLab/app/formula/domain/ast"
	domainCompiler "github.com/agoXQ/QuantLab/app/formula/domain/compiler"
)

type planner struct{}

// NewPlanner creates a new execution plan generator.
func NewPlanner() domainCompiler.Planner {
	return &planner{}
}

func (p *planner) Plan(_ context.Context, node domainAST.Node) (*domainCompiler.ExecutionPlan, error) {
	planType := p.inferPlanType(node)
	return &domainCompiler.ExecutionPlan{
		PlanType:  planType,
		Root:      node,
		Optimized: false,
	}, nil
}

func (p *planner) inferPlanType(node domainAST.Node) domainCompiler.PlanType {
	switch n := node.(type) {
	case *domainAST.BinaryExpression:
		op := strings.ToUpper(n.Operator)
		switch op {
		case "AND", "OR":
			if p.isSignalExpression(n) {
				return domainCompiler.PlanTypeSignal
			}
			return domainCompiler.PlanTypeFilter
		case ">", "<", ">=", "<=", "==", "!=":
			return domainCompiler.PlanTypeFilter
		default:
			return domainCompiler.PlanTypeSort
		}
	case *domainAST.FunctionCall:
		upper := strings.ToUpper(n.Name)
		if upper == "CROSS" || upper == "LONGCROSS" || upper == "FILTER" {
			return domainCompiler.PlanTypeSignal
		}
		if def, ok := p.getReturnType(n); ok && def == "Boolean" {
			return domainCompiler.PlanTypeFilter
		}
		return domainCompiler.PlanTypeValue
	case *domainAST.Identifier:
		return domainCompiler.PlanTypeValue
	case *domainAST.NumberLiteral:
		return domainCompiler.PlanTypeValue
	case *domainAST.BoolLiteral:
		return domainCompiler.PlanTypeFilter
	case *domainAST.Assignment:
		return p.inferPlanType(n.Value)
	case *domainAST.Program:
		if len(n.Statements) > 0 {
			return p.inferPlanType(n.Statements[len(n.Statements)-1])
		}
		return domainCompiler.PlanTypeValue
	default:
		return domainCompiler.PlanTypeValue
	}
}

func (p *planner) isSignalExpression(node domainAST.Node) bool {
	switch n := node.(type) {
	case *domainAST.FunctionCall:
		upper := strings.ToUpper(n.Name)
		return upper == "CROSS" || upper == "LONGCROSS" || upper == "FILTER"
	case *domainAST.BinaryExpression:
		return p.isSignalExpression(n.Left) || p.isSignalExpression(n.Right)
	case *domainAST.Assignment:
		return p.isSignalExpression(n.Value)
	case *domainAST.Program:
		if len(n.Statements) > 0 {
			return p.isSignalExpression(n.Statements[len(n.Statements)-1])
		}
		return false
	default:
		return false
	}
}

func (p *planner) getReturnType(call *domainAST.FunctionCall) (string, bool) {
	switch strings.ToUpper(call.Name) {
	case "CROSS", "LONGCROSS", "FILTER":
		return "Signal", true
	case "MA", "EMA", "SMA", "STD", "MACD", "RSI", "KDJ", "BOLL", "ATR":
		return "Series", true
	case "SUM", "AVG", "ABS", "MAX", "MIN":
		return "Series", true
	case "COUNT", "BARSLAST":
		return "Number", true
	case "REF", "HHV", "LLV":
		return "Series", true
	default:
		return "", false
	}
}
