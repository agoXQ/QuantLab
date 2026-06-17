package tests

import (
	"testing"

	domainFunc "github.com/agoXQ/QuantLab/app/formula/domain/function"
	infraFunc "github.com/agoXQ/QuantLab/app/formula/infrastructure/function"
	infraVar "github.com/agoXQ/QuantLab/app/formula/infrastructure/variable"
)

func TestFunctionRegistry_BuiltinsCount(t *testing.T) {
	r := infraFunc.NewRegistry()
	fns := r.ListFunctions()
	// We expect 22 built-in functions
	if len(fns) != 22 {
		t.Errorf("expected 22 built-in functions, got %d", len(fns))
	}
}

func TestFunctionRegistry_GetFunction(t *testing.T) {
	r := infraFunc.NewRegistry()

	def, ok := r.GetFunction("MA")
	if !ok {
		t.Fatal("expected to find MA function")
	}
	if def.Name != "MA" {
		t.Errorf("expected name MA, got %s", def.Name)
	}
	if def.Category != domainFunc.CategoryTechnical {
		t.Errorf("expected category Technical, got %s", def.Category)
	}
	if def.ReturnType != domainFunc.TypeSeries {
		t.Errorf("expected return type Series, got %s", def.ReturnType)
	}
	if len(def.Args) != 2 {
		t.Errorf("expected 2 args, got %d", len(def.Args))
	}
}

func TestFunctionRegistry_CaseInsensitive(t *testing.T) {
	r := infraFunc.NewRegistry()

	tests := []string{"ma", "Ma", "MA", "mA"}
	for _, name := range tests {
		_, ok := r.GetFunction(name)
		if !ok {
			t.Errorf("expected to find function %q case-insensitively", name)
		}
	}
}

func TestFunctionRegistry_ResolveName(t *testing.T) {
	r := infraFunc.NewRegistry()

	canonical, ok := r.ResolveName("ma")
	if !ok {
		t.Fatal("expected to resolve ma")
	}
	if canonical != "MA" {
		t.Errorf("expected canonical MA, got %s", canonical)
	}
}

func TestFunctionRegistry_Exists(t *testing.T) {
	r := infraFunc.NewRegistry()

	if !r.Exists("RSI") {
		t.Error("expected RSI to exist")
	}
	if !r.Exists("rsi") {
		t.Error("expected rsi to exist (case-insensitive)")
	}
	if r.Exists("NONEXISTENT") {
		t.Error("expected NONEXISTENT to not exist")
	}
}

func TestFunctionRegistry_RegisterFunction(t *testing.T) {
	r := infraFunc.NewRegistry()

	err := r.RegisterFunction(domainFunc.FunctionDefinition{
		Name:       "TEST_FUNC",
		Category:   domainFunc.CategoryMath,
		ReturnType: domainFunc.TypeNumber,
		Args: []domainFunc.ArgDef{
			{Name: "x", ArgType: "Number", Required: true},
		},
	})
	if err != nil {
		t.Fatalf("RegisterFunction failed: %v", err)
	}

	def, ok := r.GetFunction("TEST_FUNC")
	if !ok {
		t.Fatal("expected to find TEST_FUNC after registration")
	}
	if def.Name != "TEST_FUNC" {
		t.Errorf("expected TEST_FUNC, got %s", def.Name)
	}
}

func TestFunctionRegistry_DuplicateRegistration(t *testing.T) {
	r := infraFunc.NewRegistry()

	err := r.RegisterFunction(domainFunc.FunctionDefinition{
		Name:       "MA",
		Category:   domainFunc.CategoryMath,
		ReturnType: domainFunc.TypeNumber,
	})
	if err == nil {
		t.Fatal("expected error for duplicate registration")
	}
}

func TestFunctionRegistry_AllFunctionsHaveArgs(t *testing.T) {
	r := infraFunc.NewRegistry()
	fns := r.ListFunctions()

	for _, fn := range fns {
		if fn.Name == "" {
			t.Error("found function with empty name")
		}
		if fn.Category == "" {
			t.Errorf("function %s has empty category", fn.Name)
		}
		if fn.ReturnType == "" {
			t.Errorf("function %s has empty return type", fn.Name)
		}
	}
}

func TestVariableRegistry_BuiltinsCount(t *testing.T) {
	r := infraVar.NewRegistry()
	vars := r.ListVariables()
	if len(vars) != 16 {
		t.Errorf("expected 16 built-in variables, got %d", len(vars))
	}
}

func TestVariableRegistry_GetVariable(t *testing.T) {
	r := infraVar.NewRegistry()

	def, ok := r.GetVariable("CLOSE")
	if !ok {
		t.Fatal("expected to find CLOSE variable")
	}
	if def.Name != "CLOSE" {
		t.Errorf("expected CLOSE, got %s", def.Name)
	}
	if string(def.VarType) != "Series" {
		t.Errorf("expected Series type, got %s", def.VarType)
	}
}

func TestVariableRegistry_CaseInsensitive(t *testing.T) {
	r := infraVar.NewRegistry()

	tests := []string{"close", "Close", "CLOSE"}
	for _, name := range tests {
		_, ok := r.GetVariable(name)
		if !ok {
			t.Errorf("expected to find %q case-insensitively", name)
		}
	}
}

func TestVariableRegistry_Exists(t *testing.T) {
	r := infraVar.NewRegistry()

	if !r.Exists("ROE") {
		t.Error("expected ROE to exist")
	}
	if !r.Exists("roe") {
		t.Error("expected roe to exist (case-insensitive)")
	}
	if r.Exists("NONEXISTENT") {
		t.Error("expected NONEXISTENT to not exist")
	}
}

func TestVariableRegistry_AllCategories(t *testing.T) {
	r := infraVar.NewRegistry()
	vars := r.ListVariables()

	categories := make(map[string]int)
	for _, v := range vars {
		categories[v.Category]++
	}

	expectedCategories := []string{"MarketData", "Financial", "Growth", "MarketCap"}
	for _, cat := range expectedCategories {
		if categories[cat] == 0 {
			t.Errorf("expected at least one variable in category %s", cat)
		}
	}
}

func TestVariableRegistry_MarketDataVariables(t *testing.T) {
	r := infraVar.NewRegistry()
	expected := []string{"OPEN", "HIGH", "LOW", "CLOSE", "VOL", "AMOUNT"}
	for _, name := range expected {
		def, ok := r.GetVariable(name)
		if !ok {
			t.Errorf("expected market data variable %s", name)
			continue
		}
		if string(def.VarType) != "Series" {
			t.Errorf("expected %s to be Series type, got %s", name, def.VarType)
		}
	}
}

func TestVariableRegistry_FinancialVariables(t *testing.T) {
	r := infraVar.NewRegistry()
	expected := []string{"PE", "PB", "PS", "ROE", "ROA", "EPS"}
	for _, name := range expected {
		def, ok := r.GetVariable(name)
		if !ok {
			t.Errorf("expected financial variable %s", name)
			continue
		}
		if string(def.VarType) != "Number" {
			t.Errorf("expected %s to be Number type, got %s", name, def.VarType)
		}
	}
}
