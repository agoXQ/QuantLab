package tests

import (
	"context"
	"testing"
	"time"

	appFormula "github.com/agoXQ/QuantLab/app/formula/application/formula"
	infraCache "github.com/agoXQ/QuantLab/app/formula/infrastructure/cache"
	infraFunc "github.com/agoXQ/QuantLab/app/formula/infrastructure/function"
	infraLexer "github.com/agoXQ/QuantLab/app/formula/infrastructure/lexer"
	infraOptimizer "github.com/agoXQ/QuantLab/app/formula/infrastructure/optimizer"
	infraParser "github.com/agoXQ/QuantLab/app/formula/infrastructure/parser"
	infraPlanner "github.com/agoXQ/QuantLab/app/formula/infrastructure/planner"
	infraValidator "github.com/agoXQ/QuantLab/app/formula/infrastructure/validator"
	infraVar "github.com/agoXQ/QuantLab/app/formula/infrastructure/variable"
)

func newCachedService() appFormula.Service {
	funcReg := infraFunc.NewRegistry()
	varReg := infraVar.NewRegistry()

	base := appFormula.NewService(
		infraLexer.NewLexer(),
		infraParser.NewParser(funcReg, varReg),
		infraValidator.NewValidator(funcReg, varReg),
		infraOptimizer.NewOptimizer(),
		infraPlanner.NewPlanner(),
		funcReg,
	)

	cache := infraCache.NewMemory(5 * time.Minute)
	return appFormula.NewCachedService(base, cache, 0)
}

func TestCachedService_ValidateCacheHit(t *testing.T) {
	svc := newCachedService()
	ctx := context.Background()

	// First call populates cache
	result1, err := svc.Validate(ctx, "ROE > 15")
	if err != nil {
		t.Fatalf("first Validate failed: %v", err)
	}
	if !result1.Valid {
		t.Fatal("expected valid")
	}

	// Second call should hit cache
	result2, err := svc.Validate(ctx, "ROE > 15")
	if err != nil {
		t.Fatalf("second Validate failed: %v", err)
	}
	if !result2.Valid {
		t.Fatal("expected valid on cache hit")
	}
}

func TestCachedService_CompileCacheHit(t *testing.T) {
	svc := newCachedService()
	ctx := context.Background()

	// First call populates cache
	result1, err := svc.Compile(ctx, "ROE > 15 AND PE < 20")
	if err != nil {
		t.Fatalf("first Compile failed: %v", err)
	}
	if !result1.Valid {
		t.Fatal("expected valid")
	}
	if result1.Plan == nil {
		t.Fatal("expected non-nil plan")
	}

	// Second call should hit cache
	result2, err := svc.Compile(ctx, "ROE > 15 AND PE < 20")
	if err != nil {
		t.Fatalf("second Compile failed: %v", err)
	}
	if !result2.Valid {
		t.Fatal("expected valid on cache hit")
	}
	if result2.Plan == nil {
		t.Fatal("expected non-nil plan on cache hit")
	}
	if result2.Plan.PlanType != result1.Plan.PlanType {
		t.Errorf("expected same plan type %s, got %s", result1.Plan.PlanType, result2.Plan.PlanType)
	}
}

func TestCachedService_GetASTCacheHit(t *testing.T) {
	svc := newCachedService()
	ctx := context.Background()

	// First call populates cache
	node1, err := svc.GetAST(ctx, "ROE > 15")
	if err != nil {
		t.Fatalf("first GetAST failed: %v", err)
	}
	if node1 == nil {
		t.Fatal("expected non-nil AST")
	}

	// Second call should hit cache
	node2, err := svc.GetAST(ctx, "ROE > 15")
	if err != nil {
		t.Fatalf("second GetAST failed: %v", err)
	}
	if node2 == nil {
		t.Fatal("expected non-nil AST on cache hit")
	}
}

func TestCachedService_DifferentFormulasDifferentCache(t *testing.T) {
	svc := newCachedService()
	ctx := context.Background()

	// Compile first formula
	result1, err := svc.Compile(ctx, "ROE > 15")
	if err != nil {
		t.Fatalf("first Compile failed: %v", err)
	}
	if !result1.Valid {
		t.Fatal("expected valid")
	}

	// Compile different formula - should NOT hit cache
	result2, err := svc.Compile(ctx, "PE < 20")
	if err != nil {
		t.Fatalf("second Compile failed: %v", err)
	}
	if !result2.Valid {
		t.Fatal("expected valid")
	}

	// Both should have FILTER plan type
	if result1.Plan.PlanType != result2.Plan.PlanType {
		t.Errorf("expected same plan type, got %s vs %s", result1.Plan.PlanType, result2.Plan.PlanType)
	}
}

func TestCachedService_ListFunctionsPassthrough(t *testing.T) {
	svc := newCachedService()
	ctx := context.Background()

	fns1, err := svc.ListFunctions(ctx)
	if err != nil {
		t.Fatalf("first ListFunctions failed: %v", err)
	}

	fns2, err := svc.ListFunctions(ctx)
	if err != nil {
		t.Fatalf("second ListFunctions failed: %v", err)
	}

	if len(fns1) != len(fns2) {
		t.Errorf("expected same count, got %d vs %d", len(fns1), len(fns2))
	}
}

func TestCachedService_GetFunctionPassthrough(t *testing.T) {
	svc := newCachedService()
	ctx := context.Background()

	def1, err := svc.GetFunction(ctx, "MA")
	if err != nil {
		t.Fatalf("first GetFunction failed: %v", err)
	}
	if def1 == nil {
		t.Fatal("expected non-nil function")
	}

	def2, err := svc.GetFunction(ctx, "MA")
	if err != nil {
		t.Fatalf("second GetFunction failed: %v", err)
	}
	if def2 == nil {
		t.Fatal("expected non-nil function on second call")
	}
	if def1.Name != def2.Name {
		t.Errorf("expected same name, got %s vs %s", def1.Name, def2.Name)
	}
}

func TestCachedService_InvalidFormulaNotCached(t *testing.T) {
	svc := newCachedService()
	ctx := context.Background()

	// Compile invalid formula
	result, err := svc.Compile(ctx, "UNKNOWN_VAR > 15")
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}
	if result.Valid {
		t.Fatal("expected invalid")
	}
}

func TestCachedService_CompileWithFunctionCacheHit(t *testing.T) {
	svc := newCachedService()
	ctx := context.Background()

	// First call
	result1, err := svc.Compile(ctx, "CROSS(MA(CLOSE,5),MA(CLOSE,20))")
	if err != nil {
		t.Fatalf("first Compile failed: %v", err)
	}
	if !result1.Valid {
		t.Fatal("expected valid")
	}

	// Second call - cache hit
	result2, err := svc.Compile(ctx, "CROSS(MA(CLOSE,5),MA(CLOSE,20))")
	if err != nil {
		t.Fatalf("second Compile failed: %v", err)
	}
	if !result2.Valid {
		t.Fatal("expected valid on cache hit")
	}
}

func TestCachedService_FormulaHash(t *testing.T) {
	svc := newCachedService()
	h1 := svc.FormulaHash("ROE > 15")
	h2 := svc.FormulaHash("ROE > 15")
	if h1 != h2 {
		t.Error("same formula should produce same hash")
	}
}

func TestMemoryCache_Basic(t *testing.T) {
	cache := infraCache.NewMemory(1 * time.Hour)
	ctx := context.Background()

	// Get non-existent key
	data, err := cache.Get(ctx, "nonexistent")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if data != nil {
		t.Fatal("expected nil for non-existent key")
	}

	// Set and get
	err = cache.Set(ctx, "testkey", []byte("testvalue"), 0)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	data, err = cache.Get(ctx, "testkey")
	if err != nil {
		t.Fatalf("Get after Set failed: %v", err)
	}
	if string(data) != "testvalue" {
		t.Errorf("expected testvalue, got %s", string(data))
	}

	// Delete
	err = cache.Del(ctx, "testkey")
	if err != nil {
		t.Fatalf("Del failed: %v", err)
	}

	data, err = cache.Get(ctx, "testkey")
	if err != nil {
		t.Fatalf("Get after Del failed: %v", err)
	}
	if data != nil {
		t.Fatal("expected nil after delete")
	}
}

func TestMemoryCache_TTL(t *testing.T) {
	cache := infraCache.NewMemory(50 * time.Millisecond)
	ctx := context.Background()

	err := cache.Set(ctx, "ttlkey", []byte("data"), 0)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Should still be there
	data, err := cache.Get(ctx, "ttlkey")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if data == nil {
		t.Fatal("expected data before TTL expiry")
	}

	// Wait for expiry
	time.Sleep(100 * time.Millisecond)

	data, err = cache.Get(ctx, "ttlkey")
	if err != nil {
		t.Fatalf("Get after TTL failed: %v", err)
	}
	if data != nil {
		t.Fatal("expected nil after TTL expiry")
	}
}
