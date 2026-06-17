package tests

import (
	"context"
	"testing"

	appFormula "github.com/agoXQ/QuantLab/app/formula/application/formula"
	infraFunc "github.com/agoXQ/QuantLab/app/formula/infrastructure/function"
	infraLexer "github.com/agoXQ/QuantLab/app/formula/infrastructure/lexer"
	infraOptimizer "github.com/agoXQ/QuantLab/app/formula/infrastructure/optimizer"
	infraParser "github.com/agoXQ/QuantLab/app/formula/infrastructure/parser"
	infraPlanner "github.com/agoXQ/QuantLab/app/formula/infrastructure/planner"
	infraValidator "github.com/agoXQ/QuantLab/app/formula/infrastructure/validator"
	infraVar "github.com/agoXQ/QuantLab/app/formula/infrastructure/variable"
)

func newService() appFormula.Service {
	funcReg := infraFunc.NewRegistry()
	varReg := infraVar.NewRegistry()

	return appFormula.NewService(
		infraLexer.NewLexer(),
		infraParser.NewParser(funcReg, varReg),
		infraValidator.NewValidator(funcReg, varReg),
		infraOptimizer.NewOptimizer(),
		infraPlanner.NewPlanner(),
		funcReg,
	)
}

func TestService_ValidateValid(t *testing.T) {
	svc := newService()
	result, err := svc.Validate(context.Background(), "ROE > 15")
	if err != nil {
		t.Fatalf("Validate failed: %v", err)
	}
	if !result.Valid {
		t.Fatalf("expected valid, got errors: %v", result.Errors)
	}
}

func TestService_ValidateInvalid(t *testing.T) {
	svc := newService()
	result, err := svc.Validate(context.Background(), "UNKNOWN_VAR > 15")
	if err != nil {
		t.Fatalf("Validate failed: %v", err)
	}
	if result.Valid {
		t.Fatal("expected invalid for unknown variable")
	}
}

func TestService_ValidateSyntaxError(t *testing.T) {
	svc := newService()
	result, err := svc.Validate(context.Background(), "ROE >")
	if err != nil {
		t.Fatalf("Validate failed: %v", err)
	}
	if result.Valid {
		t.Fatal("expected invalid for syntax error")
	}
}

func TestService_CompileValid(t *testing.T) {
	svc := newService()
	result, err := svc.Compile(context.Background(), "ROE > 15")
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}
	if !result.Valid {
		t.Fatalf("expected valid compile")
	}
	if result.Plan == nil {
		t.Fatal("expected non-nil plan")
	}
	if result.AST == nil {
		t.Fatal("expected non-nil AST")
	}
}

func TestService_CompileInvalid(t *testing.T) {
	svc := newService()
	result, err := svc.Compile(context.Background(), "UNKNOWN_VAR > 15")
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}
	if result.Valid {
		t.Fatal("expected invalid compile for unknown variable")
	}
}

func TestService_CompileAndExpression(t *testing.T) {
	svc := newService()
	result, err := svc.Compile(context.Background(), "ROE > 15 AND PE < 20")
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}
	if !result.Valid {
		t.Fatalf("expected valid compile")
	}
	if result.Plan == nil {
		t.Fatal("expected non-nil plan")
	}
}

func TestService_CompileFunctionCall(t *testing.T) {
	svc := newService()
	result, err := svc.Compile(context.Background(), "MA(CLOSE,5)")
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}
	if !result.Valid {
		t.Fatalf("expected valid compile, got errors: %v", result)
	}
}

func TestService_CompileCrossFunction(t *testing.T) {
	svc := newService()
	result, err := svc.Compile(context.Background(), "CROSS(MA(CLOSE,5),MA(CLOSE,20))")
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}
	if !result.Valid {
		t.Fatalf("expected valid compile")
	}
}

func TestService_GetAST(t *testing.T) {
	svc := newService()
	node, err := svc.GetAST(context.Background(), "ROE > 15")
	if err != nil {
		t.Fatalf("GetAST failed: %v", err)
	}
	if node == nil {
		t.Fatal("expected non-nil AST node")
	}
}

func TestService_ListFunctions(t *testing.T) {
	svc := newService()
	fns, err := svc.ListFunctions(context.Background())
	if err != nil {
		t.Fatalf("ListFunctions failed: %v", err)
	}
	if len(fns) == 0 {
		t.Fatal("expected at least one function")
	}
}

func TestService_GetFunction(t *testing.T) {
	svc := newService()
	def, err := svc.GetFunction(context.Background(), "MA")
	if err != nil {
		t.Fatalf("GetFunction failed: %v", err)
	}
	if def == nil {
		t.Fatal("expected non-nil function definition")
	}
	if def.Name != "MA" {
		t.Errorf("expected MA, got %s", def.Name)
	}
}

func TestService_GetFunctionNotFound(t *testing.T) {
	svc := newService()
	def, err := svc.GetFunction(context.Background(), "NONEXISTENT")
	if err != nil {
		t.Fatalf("GetFunction failed: %v", err)
	}
	if def != nil {
		t.Fatal("expected nil for unknown function")
	}
}

func TestService_FormulaHash(t *testing.T) {
	svc := newService()
	h1 := svc.FormulaHash("ROE > 15")
	h2 := svc.FormulaHash("ROE > 15")
	h3 := svc.FormulaHash("PE < 20")

	if h1 != h2 {
		t.Error("same formula should produce same hash")
	}
	if h1 == h3 {
		t.Error("different formulas should produce different hashes")
	}
	if len(h1) != 64 {
		t.Errorf("expected SHA256 hex (64 chars), got %d", len(h1))
	}
}

func TestService_CompileSortFormula(t *testing.T) {
	svc := newService()
	result, err := svc.Compile(context.Background(), "ROE * ProfitGrowth")
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}
	if !result.Valid {
		t.Fatalf("expected valid compile")
	}
}

func TestService_ValidateComplexExpression(t *testing.T) {
	svc := newService()
	formula := "ROE > 15 AND PE < 20 AND MA(CLOSE,5) > MA(CLOSE,20)"
	result, err := svc.Validate(context.Background(), formula)
	if err != nil {
		t.Fatalf("Validate failed: %v", err)
	}
	if !result.Valid {
		t.Fatalf("expected valid, got errors: %v", result.Errors)
	}
}

func TestService_CompileWithOptimization(t *testing.T) {
	svc := newService()
	// 1 + 2 should be constant-folded to 3
	result, err := svc.Compile(context.Background(), "ROE > 1 + 2")
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}
	if !result.Valid {
		t.Fatalf("expected valid compile")
	}
}

func TestService_EmptyFormula(t *testing.T) {
	svc := newService()
	_, err := svc.Compile(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty formula")
	}
}
