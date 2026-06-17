package tests

import (
	"context"
	"fmt"
	"testing"

	domainAST "github.com/agoXQ/QuantLab/app/formula/domain/ast"
	infraFunc "github.com/agoXQ/QuantLab/app/formula/infrastructure/function"
	infraValidator "github.com/agoXQ/QuantLab/app/formula/infrastructure/validator"
	infraVar "github.com/agoXQ/QuantLab/app/formula/infrastructure/variable"
)

func validateNode(t *testing.T, node domainAST.Node) (bool, []string) {
	t.Helper()
	funcReg := infraFunc.NewRegistry()
	varReg := infraVar.NewRegistry()
	v := infraValidator.NewValidator(funcReg, varReg)

	result, err := v.Validate(context.Background(), node)
	if err != nil {
		t.Fatalf("Validate failed: %v", err)
	}

	messages := make([]string, 0)
	if !result.Valid {
		for _, e := range result.Errors {
			messages = append(messages, fmt.Sprintf("%d: %s", e.Code, e.Message))
		}
	}
	return result.Valid, messages
}

func TestValidator_ValidComparison(t *testing.T) {
	node := &domainAST.BinaryExpression{
		Left:     &domainAST.Identifier{Name: "ROE"},
		Operator: ">",
		Right:    &domainAST.NumberLiteral{Value: 15},
	}
	valid, msgs := validateNode(t, node)
	if !valid {
		t.Fatalf("expected valid, got errors: %v", msgs)
	}
}

func TestValidator_ValidAndExpression(t *testing.T) {
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
	valid, msgs := validateNode(t, node)
	if !valid {
		t.Fatalf("expected valid, got errors: %v", msgs)
	}
}

func TestValidator_UnknownVariable(t *testing.T) {
	node := &domainAST.Identifier{Name: "UNKNOWN_VAR"}
	valid, msgs := validateNode(t, node)
	if valid {
		t.Fatal("expected invalid for unknown variable")
	}
	if len(msgs) == 0 {
		t.Fatal("expected error message")
	}
}

func TestValidator_UnknownFunction(t *testing.T) {
	node := &domainAST.FunctionCall{
		Name: "UNKNOWN_FUNC",
		Args: []domainAST.Node{&domainAST.NumberLiteral{Value: 1}},
	}
	valid, _ := validateNode(t, node)
	if valid {
		t.Fatal("expected invalid for unknown function")
	}
}

func TestValidator_ValidFunctionCall(t *testing.T) {
	node := &domainAST.FunctionCall{
		Name: "MA",
		Args: []domainAST.Node{
			&domainAST.Identifier{Name: "CLOSE"},
			&domainAST.NumberLiteral{Value: 5},
		},
	}
	valid, msgs := validateNode(t, node)
	if !valid {
		t.Fatalf("expected valid MA call, got errors: %v", msgs)
	}
}

func TestValidator_FunctionArgCountMismatch(t *testing.T) {
	node := &domainAST.FunctionCall{
		Name: "MA",
		Args: []domainAST.Node{
			&domainAST.Identifier{Name: "CLOSE"},
		},
	}
	valid, _ := validateNode(t, node)
	if valid {
		t.Fatal("expected invalid for missing argument")
	}
}

func TestValidator_FutureFunctionDetection(t *testing.T) {
	node := &domainAST.FunctionCall{
		Name: "REF",
		Args: []domainAST.Node{
			&domainAST.Identifier{Name: "CLOSE"},
			&domainAST.NumberLiteral{Value: -1},
		},
	}
	valid, msgs := validateNode(t, node)
	if valid {
		t.Fatal("expected invalid for future function")
	}
	if len(msgs) == 0 {
		t.Fatal("expected future function error")
	}
}

func TestValidator_TypeMismatchInComparison(t *testing.T) {
	node := &domainAST.BinaryExpression{
		Left:     &domainAST.Identifier{Name: "PE"},
		Operator: ">",
		Right:    &domainAST.StringLiteral{Value: "ABC"},
	}
	valid, _ := validateNode(t, node)
	if valid {
		t.Fatal("expected invalid for type mismatch (Number vs String)")
	}
}

func TestValidator_AndWithNonBoolean(t *testing.T) {
	node := &domainAST.BinaryExpression{
		Left:     &domainAST.Identifier{Name: "ROE"},
		Operator: "AND",
		Right:    &domainAST.Identifier{Name: "PE"},
	}
	valid, _ := validateNode(t, node)
	if valid {
		t.Fatal("expected invalid for AND with non-boolean operands")
	}
}

func TestValidator_ValidNotExpression(t *testing.T) {
	node := &domainAST.UnaryExpression{
		Operator: "NOT",
		Operand: &domainAST.BinaryExpression{
			Left:     &domainAST.Identifier{Name: "ROE"},
			Operator: ">",
			Right:    &domainAST.NumberLiteral{Value: 15},
		},
	}
	valid, msgs := validateNode(t, node)
	if !valid {
		t.Fatalf("expected valid NOT expression, got errors: %v", msgs)
	}
}

func TestValidator_ArithmeticExpression(t *testing.T) {
	node := &domainAST.BinaryExpression{
		Left:     &domainAST.Identifier{Name: "ROE"},
		Operator: "*",
		Right:    &domainAST.Identifier{Name: "ProfitGrowth"},
	}
	valid, msgs := validateNode(t, node)
	if !valid {
		t.Fatalf("expected valid arithmetic, got errors: %v", msgs)
	}
}

func TestValidator_CrossFunction(t *testing.T) {
	node := &domainAST.FunctionCall{
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
	valid, msgs := validateNode(t, node)
	if !valid {
		t.Fatalf("expected valid CROSS, got errors: %v", msgs)
	}
}

func TestValidator_TrueFalseLiterals(t *testing.T) {
	node := &domainAST.Identifier{Name: "TRUE"}
	valid, msgs := validateNode(t, node)
	if !valid {
		t.Fatalf("expected TRUE to be valid, got errors: %v", msgs)
	}

	node2 := &domainAST.Identifier{Name: "FALSE"}
	valid, msgs = validateNode(t, node2)
	if !valid {
		t.Fatalf("expected FALSE to be valid, got errors: %v", msgs)
	}
}

func TestValidator_BoolLiteral(t *testing.T) {
	node := &domainAST.BoolLiteral{Value: true}
	valid, msgs := validateNode(t, node)
	if !valid {
		t.Fatalf("expected BoolLiteral to be valid, got errors: %v", msgs)
	}
}

func TestValidator_NumberLiteral(t *testing.T) {
	node := &domainAST.NumberLiteral{Value: 42}
	valid, msgs := validateNode(t, node)
	if !valid {
		t.Fatalf("expected NumberLiteral to be valid, got errors: %v", msgs)
	}
}
