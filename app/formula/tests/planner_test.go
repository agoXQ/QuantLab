package tests

import (
	"context"
	"testing"

	domainAST "github.com/agoXQ/QuantLab/app/formula/domain/ast"
	domainCompiler "github.com/agoXQ/QuantLab/app/formula/domain/compiler"
	infraPlanner "github.com/agoXQ/QuantLab/app/formula/infrastructure/planner"
)

func planNode(t *testing.T, node domainAST.Node) *domainCompiler.ExecutionPlan {
	t.Helper()
	p := infraPlanner.NewPlanner()
	plan, err := p.Plan(context.Background(), node)
	if err != nil {
		t.Fatalf("Plan failed: %v", err)
	}
	return plan
}

func TestPlanner_ComparisonFilter(t *testing.T) {
	node := &domainAST.BinaryExpression{
		Left:     &domainAST.Identifier{Name: "ROE"},
		Operator: ">",
		Right:    &domainAST.NumberLiteral{Value: 15},
	}
	plan := planNode(t, node)
	if plan.PlanType != domainCompiler.PlanTypeFilter {
		t.Errorf("expected FILTER, got %s", plan.PlanType)
	}
}

func TestPlanner_AndFilter(t *testing.T) {
	node := &domainAST.BinaryExpression{
		Left: &domainAST.BinaryExpression{
			Left:     &domainAST.Identifier{Name: "ROE"},
			Operator: ">",
			Right:    &domainAST.NumberLiteral{Value: 15},
		},
		Operator: "AND",
		Right: &domainAST.BinaryExpression{
			Left:     &domainAST.Identifier{Name: "PE"},
			Operator: "<",
			Right:    &domainAST.NumberLiteral{Value: 20},
		},
	}
	plan := planNode(t, node)
	if plan.PlanType != domainCompiler.PlanTypeFilter {
		t.Errorf("expected FILTER, got %s", plan.PlanType)
	}
}

func TestPlanner_SignalCross(t *testing.T) {
	node := &domainAST.FunctionCall{
		Name: "CROSS",
		Args: []domainAST.Node{
			&domainAST.Identifier{Name: "MA"},
			&domainAST.Identifier{Name: "MA"},
		},
	}
	plan := planNode(t, node)
	if plan.PlanType != domainCompiler.PlanTypeSignal {
		t.Errorf("expected SIGNAL, got %s", plan.PlanType)
	}
}

func TestPlanner_SignalLongCross(t *testing.T) {
	node := &domainAST.FunctionCall{
		Name: "LONGCROSS",
		Args: []domainAST.Node{
			&domainAST.Identifier{Name: "MA"},
			&domainAST.Identifier{Name: "MA"},
			&domainAST.NumberLiteral{Value: 3},
		},
	}
	plan := planNode(t, node)
	if plan.PlanType != domainCompiler.PlanTypeSignal {
		t.Errorf("expected SIGNAL, got %s", plan.PlanType)
	}
}

func TestPlanner_ValueFunction(t *testing.T) {
	node := &domainAST.FunctionCall{
		Name: "MA",
		Args: []domainAST.Node{
			&domainAST.Identifier{Name: "CLOSE"},
			&domainAST.NumberLiteral{Value: 5},
		},
	}
	plan := planNode(t, node)
	if plan.PlanType != domainCompiler.PlanTypeValue {
		t.Errorf("expected VALUE, got %s", plan.PlanType)
	}
}

func TestPlanner_SortArithmetic(t *testing.T) {
	node := &domainAST.BinaryExpression{
		Left:     &domainAST.Identifier{Name: "ROE"},
		Operator: "*",
		Right:    &domainAST.Identifier{Name: "ProfitGrowth"},
	}
	plan := planNode(t, node)
	if plan.PlanType != domainCompiler.PlanTypeSort {
		t.Errorf("expected SORT, got %s", plan.PlanType)
	}
}

func TestPlanner_IdentifierValue(t *testing.T) {
	node := &domainAST.Identifier{Name: "ROE"}
	plan := planNode(t, node)
	if plan.PlanType != domainCompiler.PlanTypeValue {
		t.Errorf("expected VALUE, got %s", plan.PlanType)
	}
}

func TestPlanner_NumberLiteralValue(t *testing.T) {
	node := &domainAST.NumberLiteral{Value: 42}
	plan := planNode(t, node)
	if plan.PlanType != domainCompiler.PlanTypeValue {
		t.Errorf("expected VALUE, got %s", plan.PlanType)
	}
}

func TestPlanner_BoolLiteralFilter(t *testing.T) {
	node := &domainAST.BoolLiteral{Value: true}
	plan := planNode(t, node)
	if plan.PlanType != domainCompiler.PlanTypeFilter {
		t.Errorf("expected FILTER, got %s", plan.PlanType)
	}
}

func TestPlanner_SignalInAndExpression(t *testing.T) {
	// CROSS(MA(CLOSE,5),MA(CLOSE,20)) AND CLOSE > MA(CLOSE,20)
	cross := &domainAST.FunctionCall{
		Name: "CROSS",
		Args: []domainAST.Node{
			&domainAST.FunctionCall{
				Name: "MA",
				Args: []domainAST.Node{
					&domainAST.Identifier{Name: "CLOSE"},
					&domainAST.NumberLiteral{Value: 5},
				},
			},
			&domainAST.FunctionCall{
				Name: "MA",
				Args: []domainAST.Node{
					&domainAST.Identifier{Name: "CLOSE"},
					&domainAST.NumberLiteral{Value: 20},
				},
			},
		},
	}
	closeCheck := &domainAST.BinaryExpression{
		Left:     &domainAST.Identifier{Name: "CLOSE"},
		Operator: ">",
		Right: &domainAST.FunctionCall{
			Name: "MA",
			Args: []domainAST.Node{
				&domainAST.Identifier{Name: "CLOSE"},
				&domainAST.NumberLiteral{Value: 20},
			},
		},
	}
	node := &domainAST.BinaryExpression{
		Left:     cross,
		Operator: "AND",
		Right:    closeCheck,
	}
	plan := planNode(t, node)
	if plan.PlanType != domainCompiler.PlanTypeSignal {
		t.Errorf("expected SIGNAL (contains CROSS), got %s", plan.PlanType)
	}
}

func TestPlanner_FilterFunction(t *testing.T) {
	node := &domainAST.FunctionCall{
		Name: "FILTER",
		Args: []domainAST.Node{
			&domainAST.Identifier{Name: "CROSS"},
			&domainAST.NumberLiteral{Value: 3},
		},
	}
	plan := planNode(t, node)
	if plan.PlanType != domainCompiler.PlanTypeSignal {
		t.Errorf("expected SIGNAL, got %s", plan.PlanType)
	}
}

func TestPlanner_NotOptimizedByDefault(t *testing.T) {
	node := &domainAST.Identifier{Name: "ROE"}
	plan := planNode(t, node)
	if plan.Optimized {
		t.Error("expected Optimized to be false by default")
	}
}
