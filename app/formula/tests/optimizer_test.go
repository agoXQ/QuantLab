package tests

import (
	"context"
	"testing"

	domainAST "github.com/agoXQ/QuantLab/app/formula/domain/ast"
	infraOptimizer "github.com/agoXQ/QuantLab/app/formula/infrastructure/optimizer"
)

func optimizeNode(t *testing.T, node domainAST.Node) domainAST.Node {
	t.Helper()
	opt := infraOptimizer.NewOptimizer()
	result, err := opt.Optimize(context.Background(), node)
	if err != nil {
		t.Fatalf("Optimize failed: %v", err)
	}
	return result
}

func TestOptimizer_ConstantFoldingAddition(t *testing.T) {
	node := &domainAST.BinaryExpression{
		Left:     &domainAST.NumberLiteral{Value: 1},
		Operator: "+",
		Right:    &domainAST.NumberLiteral{Value: 2},
	}
	result := optimizeNode(t, node)
	num, ok := result.(*domainAST.NumberLiteral)
	if !ok {
		t.Fatalf("expected *NumberLiteral after folding, got %T", result)
	}
	if num.Value != 3 {
		t.Errorf("expected 3, got %f", num.Value)
	}
}

func TestOptimizer_ConstantFoldingComplex(t *testing.T) {
	node := &domainAST.BinaryExpression{
		Left:     &domainAST.NumberLiteral{Value: 10},
		Operator: "*",
		Right:    &domainAST.NumberLiteral{Value: 5},
	}
	result := optimizeNode(t, node)
	num, ok := result.(*domainAST.NumberLiteral)
	if !ok {
		t.Fatalf("expected *NumberLiteral, got %T", result)
	}
	if num.Value != 50 {
		t.Errorf("expected 50, got %f", num.Value)
	}
}

func TestOptimizer_ConstantFoldingDivision(t *testing.T) {
	node := &domainAST.BinaryExpression{
		Left:     &domainAST.NumberLiteral{Value: 10},
		Operator: "/",
		Right:    &domainAST.NumberLiteral{Value: 3},
	}
	result := optimizeNode(t, node)
	num, ok := result.(*domainAST.NumberLiteral)
	if !ok {
		t.Fatalf("expected *NumberLiteral, got %T", result)
	}
	if num.Value != 10.0/3.0 {
		t.Errorf("expected %f, got %f", 10.0/3.0, num.Value)
	}
}

func TestOptimizer_NoDivisionByZeroFold(t *testing.T) {
	node := &domainAST.BinaryExpression{
		Left:     &domainAST.NumberLiteral{Value: 10},
		Operator: "/",
		Right:    &domainAST.NumberLiteral{Value: 0},
	}
	result := optimizeNode(t, node)
	// Should NOT fold division by zero - keep as BinaryExpression
	_, ok := result.(*domainAST.BinaryExpression)
	if !ok {
		t.Fatalf("expected *BinaryExpression (no fold for div by zero), got %T", result)
	}
}

func TestOptimizer_BooleanSimplifyAndTrue(t *testing.T) {
	node := &domainAST.BinaryExpression{
		Left:     &domainAST.Identifier{Name: "ROE"},
		Operator: "AND",
		Right:    &domainAST.BoolLiteral{Value: true},
	}
	result := optimizeNode(t, node)
	// A AND TRUE -> A
	_, ok := result.(*domainAST.Identifier)
	if !ok {
		t.Fatalf("expected *Identifier (A AND TRUE -> A), got %T", result)
	}
}

func TestOptimizer_BooleanSimplifyAndFalse(t *testing.T) {
	node := &domainAST.BinaryExpression{
		Left:     &domainAST.Identifier{Name: "ROE"},
		Operator: "AND",
		Right:    &domainAST.BoolLiteral{Value: false},
	}
	result := optimizeNode(t, node)
	// A AND FALSE -> FALSE
	b, ok := result.(*domainAST.BoolLiteral)
	if !ok {
		t.Fatalf("expected *BoolLiteral (A AND FALSE -> FALSE), got %T", result)
	}
	if b.Value {
		t.Error("expected false")
	}
}

func TestOptimizer_BooleanSimplifyOrTrue(t *testing.T) {
	node := &domainAST.BinaryExpression{
		Left:     &domainAST.Identifier{Name: "ROE"},
		Operator: "OR",
		Right:    &domainAST.BoolLiteral{Value: true},
	}
	result := optimizeNode(t, node)
	// A OR TRUE -> TRUE
	b, ok := result.(*domainAST.BoolLiteral)
	if !ok {
		t.Fatalf("expected *BoolLiteral (A OR TRUE -> TRUE), got %T", result)
	}
	if !b.Value {
		t.Error("expected true")
	}
}

func TestOptimizer_BooleanSimplifyOrFalse(t *testing.T) {
	node := &domainAST.BinaryExpression{
		Left:     &domainAST.Identifier{Name: "ROE"},
		Operator: "OR",
		Right:    &domainAST.BoolLiteral{Value: false},
	}
	result := optimizeNode(t, node)
	// A OR FALSE -> A
	_, ok := result.(*domainAST.Identifier)
	if !ok {
		t.Fatalf("expected *Identifier (A OR FALSE -> A), got %T", result)
	}
}

func TestOptimizer_DoubleNegation(t *testing.T) {
	node := &domainAST.UnaryExpression{
		Operator: "NOT",
		Operand: &domainAST.UnaryExpression{
			Operator: "NOT",
			Operand:  &domainAST.Identifier{Name: "ROE"},
		},
	}
	result := optimizeNode(t, node)
	// NOT NOT A -> A
	_, ok := result.(*domainAST.Identifier)
	if !ok {
		t.Fatalf("expected *Identifier (NOT NOT A -> A), got %T", result)
	}
}

func TestOptimizer_NotTrue(t *testing.T) {
	node := &domainAST.UnaryExpression{
		Operator: "NOT",
		Operand:  &domainAST.BoolLiteral{Value: true},
	}
	result := optimizeNode(t, node)
	b, ok := result.(*domainAST.BoolLiteral)
	if !ok {
		t.Fatalf("expected *BoolLiteral, got %T", result)
	}
	if b.Value {
		t.Error("expected false")
	}
}

func TestOptimizer_NotFalse(t *testing.T) {
	node := &domainAST.UnaryExpression{
		Operator: "NOT",
		Operand:  &domainAST.BoolLiteral{Value: false},
	}
	result := optimizeNode(t, node)
	b, ok := result.(*domainAST.BoolLiteral)
	if !ok {
		t.Fatalf("expected *BoolLiteral, got %T", result)
	}
	if !b.Value {
		t.Error("expected true")
	}
}

func TestOptimizer_ConstantFoldBothSides(t *testing.T) {
	// (1 + 2) * (3 + 4) -> 3 * 7 -> 21
	node := &domainAST.BinaryExpression{
		Left: &domainAST.BinaryExpression{
			Left:     &domainAST.NumberLiteral{Value: 1},
			Operator: "+",
			Right:    &domainAST.NumberLiteral{Value: 2},
		},
		Operator: "*",
		Right: &domainAST.BinaryExpression{
			Left:     &domainAST.NumberLiteral{Value: 3},
			Operator: "+",
			Right:    &domainAST.NumberLiteral{Value: 4},
		},
	}
	result := optimizeNode(t, node)
	num, ok := result.(*domainAST.NumberLiteral)
	if !ok {
		t.Fatalf("expected *NumberLiteral, got %T", result)
	}
	if num.Value != 21 {
		t.Errorf("expected 21, got %f", num.Value)
	}
}

func TestOptimizer_PreservesFunctionCalls(t *testing.T) {
	node := &domainAST.FunctionCall{
		Name: "MA",
		Args: []domainAST.Node{
			&domainAST.Identifier{Name: "CLOSE"},
			&domainAST.NumberLiteral{Value: 5},
		},
	}
	result := optimizeNode(t, node)
	fn, ok := result.(*domainAST.FunctionCall)
	if !ok {
		t.Fatalf("expected *FunctionCall, got %T", result)
	}
	if fn.Name != "MA" {
		t.Errorf("expected MA, got %s", fn.Name)
	}
}

func TestOptimizer_ModuloFolding(t *testing.T) {
	node := &domainAST.BinaryExpression{
		Left:     &domainAST.NumberLiteral{Value: 10},
		Operator: "%",
		Right:    &domainAST.NumberLiteral{Value: 3},
	}
	result := optimizeNode(t, node)
	num, ok := result.(*domainAST.NumberLiteral)
	if !ok {
		t.Fatalf("expected *NumberLiteral, got %T", result)
	}
	if num.Value != 1 {
		t.Errorf("expected 1, got %f", num.Value)
	}
}

func TestOptimizer_FoldBooleanBothSides(t *testing.T) {
	node := &domainAST.BinaryExpression{
		Left:     &domainAST.BoolLiteral{Value: true},
		Operator: "AND",
		Right:    &domainAST.BoolLiteral{Value: false},
	}
	result := optimizeNode(t, node)
	b, ok := result.(*domainAST.BoolLiteral)
	if !ok {
		t.Fatalf("expected *BoolLiteral, got %T", result)
	}
	if b.Value {
		t.Error("expected false")
	}
}
