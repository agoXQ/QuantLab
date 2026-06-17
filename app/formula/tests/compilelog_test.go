package tests

import (
	"context"
	"testing"
	"time"

	appFormula "github.com/agoXQ/QuantLab/app/formula/application/formula"
	domainLog "github.com/agoXQ/QuantLab/app/formula/domain/compilelog"
	infraLog "github.com/agoXQ/QuantLab/app/formula/infrastructure/compilelog"
	infraFunc "github.com/agoXQ/QuantLab/app/formula/infrastructure/function"
	infraLexer "github.com/agoXQ/QuantLab/app/formula/infrastructure/lexer"
	infraOptimizer "github.com/agoXQ/QuantLab/app/formula/infrastructure/optimizer"
	infraParser "github.com/agoXQ/QuantLab/app/formula/infrastructure/parser"
	infraPlanner "github.com/agoXQ/QuantLab/app/formula/infrastructure/planner"
	infraValidator "github.com/agoXQ/QuantLab/app/formula/infrastructure/validator"
	infraVar "github.com/agoXQ/QuantLab/app/formula/infrastructure/variable"
)

func newLoggedService() (appFormula.Service, domainLog.Repository) {
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

	logRepo := infraLog.NewMemoryRepository()
	svc := appFormula.NewLoggedService(base, logRepo)
	return svc, logRepo
}

func TestCompileLog_RecordsSuccessfulCompile(t *testing.T) {
	svc, repo := newLoggedService()
	ctx := context.Background()

	result, err := svc.Compile(ctx, "ROE > 15")
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}
	if !result.Valid {
		t.Fatal("expected valid compile")
	}

	hash := svc.FormulaHash("ROE > 15")
	records, err := repo.ListByHash(ctx, hash, 10, 0)
	if err != nil {
		t.Fatalf("ListByHash failed: %v", err)
	}

	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}

	rec := records[0]
	if !rec.Success {
		t.Error("expected success=true")
	}
	if rec.FormulaHash != hash {
		t.Errorf("expected hash %s, got %s", hash, rec.FormulaHash)
	}
	if rec.Formula != "ROE > 15" {
		t.Errorf("expected formula 'ROE > 15', got %q", rec.Formula)
	}
	if rec.CompileTimeMs < 0 {
		t.Error("expected non-negative compile time")
	}
	if rec.CreatedAt.IsZero() {
		t.Error("expected non-zero created_at")
	}
}

func TestCompileLog_RecordsFailedCompile(t *testing.T) {
	svc, repo := newLoggedService()
	ctx := context.Background()

	result, err := svc.Compile(ctx, "UNKNOWN_VAR > 15")
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}
	if result.Valid {
		t.Fatal("expected invalid compile")
	}

	hash := svc.FormulaHash("UNKNOWN_VAR > 15")
	records, err := repo.ListByHash(ctx, hash, 10, 0)
	if err != nil {
		t.Fatalf("ListByHash failed: %v", err)
	}

	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}

	rec := records[0]
	if rec.Success {
		t.Error("expected success=false")
	}
	if rec.ErrorCode == 0 {
		t.Error("expected non-zero error_code")
	}
}

func TestCompileLog_MultipleCompilesSameFormula(t *testing.T) {
	svc, repo := newLoggedService()
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		_, err := svc.Compile(ctx, "ROE > 15")
		if err != nil {
			t.Fatalf("Compile %d failed: %v", i, err)
		}
	}

	hash := svc.FormulaHash("ROE > 15")
	records, err := repo.ListByHash(ctx, hash, 10, 0)
	if err != nil {
		t.Fatalf("ListByHash failed: %v", err)
	}

	if len(records) != 3 {
		t.Fatalf("expected 3 records, got %d", len(records))
	}

	// Records should be in DESC order
	if records[0].ID < records[1].ID {
		t.Error("expected records in DESC order (newest first)")
	}
}

func TestCompileLog_DifferentFormulas(t *testing.T) {
	svc, repo := newLoggedService()
	ctx := context.Background()

	formulas := []string{"ROE > 15", "PE < 20", "MA(CLOSE,5) > MA(CLOSE,20)"}
	for _, f := range formulas {
		_, err := svc.Compile(ctx, f)
		if err != nil {
			t.Fatalf("Compile %q failed: %v", f, err)
		}
	}

	for _, f := range formulas {
		hash := svc.FormulaHash(f)
		records, err := repo.ListByHash(ctx, hash, 10, 0)
		if err != nil {
			t.Fatalf("ListByHash for %q failed: %v", f, err)
		}
		if len(records) != 1 {
			t.Errorf("expected 1 record for %q, got %d", f, len(records))
		}
	}
}

func TestCompileLog_RecordsErrorCode(t *testing.T) {
	svc, repo := newLoggedService()
	ctx := context.Background()

	// Compile with unknown variable should record error code 1001
	result, err := svc.Compile(ctx, "UNKNOWN_VAR > 15")
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}
	if result.Valid {
		t.Fatal("expected invalid")
	}

	hash := svc.FormulaHash("UNKNOWN_VAR > 15")
	records, err := repo.ListByHash(ctx, hash, 10, 0)
	if err != nil {
		t.Fatalf("ListByHash failed: %v", err)
	}

	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}

	if records[0].ErrorCode != 1001 {
		t.Errorf("expected error_code 1001 (UNKNOWN_VARIABLE), got %d", records[0].ErrorCode)
	}
}

func TestCompileLog_Pagination(t *testing.T) {
	svc, repo := newLoggedService()
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		_, err := svc.Compile(ctx, "ROE > 15")
		if err != nil {
			t.Fatalf("Compile %d failed: %v", i, err)
		}
	}

	hash := svc.FormulaHash("ROE > 15")

	// First page: 2 records
	page1, err := repo.ListByHash(ctx, hash, 2, 0)
	if err != nil {
		t.Fatalf("ListByHash page1 failed: %v", err)
	}
	if len(page1) != 2 {
		t.Errorf("expected 2 records on page 1, got %d", len(page1))
	}

	// Second page: 2 records
	page2, err := repo.ListByHash(ctx, hash, 2, 2)
	if err != nil {
		t.Fatalf("ListByHash page2 failed: %v", err)
	}
	if len(page2) != 2 {
		t.Errorf("expected 2 records on page 2, got %d", len(page2))
	}

	// Third page: 1 record
	page3, err := repo.ListByHash(ctx, hash, 2, 4)
	if err != nil {
		t.Fatalf("ListByHash page3 failed: %v", err)
	}
	if len(page3) != 1 {
		t.Errorf("expected 1 record on page 3, got %d", len(page3))
	}

	// Pages should not overlap
	if page1[0].ID == page2[0].ID {
		t.Error("pages should not overlap")
	}
}

func TestCompileLog_ValidateNotLogged(t *testing.T) {
	svc, repo := newLoggedService()
	ctx := context.Background()

	_, err := svc.Validate(ctx, "ROE > 15")
	if err != nil {
		t.Fatalf("Validate failed: %v", err)
	}

	// Validate should not create compile log records
	hash := svc.FormulaHash("ROE > 15")
	records, err := repo.ListByHash(ctx, hash, 10, 0)
	if err != nil {
		t.Fatalf("ListByHash failed: %v", err)
	}
	if len(records) != 0 {
		t.Errorf("expected 0 records from Validate, got %d", len(records))
	}
}

func TestCompileLog_RecordsCompileTime(t *testing.T) {
	svc, repo := newLoggedService()
	ctx := context.Background()

	_, err := svc.Compile(ctx, "CROSS(MA(CLOSE,5),MA(CLOSE,20))")
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	hash := svc.FormulaHash("CROSS(MA(CLOSE,5),MA(CLOSE,20))")
	records, err := repo.ListByHash(ctx, hash, 10, 0)
	if err != nil {
		t.Fatalf("ListByHash failed: %v", err)
	}

	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}

	if records[0].CompileTimeMs < 0 {
		t.Error("expected non-negative compile time")
	}
}

func TestMemoryRepository_EnsureTableNoop(t *testing.T) {
	repo := infraLog.NewMemoryRepository()
	err := repo.EnsureTable(context.Background())
	if err != nil {
		t.Fatalf("EnsureTable should be no-op for memory repo: %v", err)
	}
}

func TestCompileLog_SaveSetsID(t *testing.T) {
	repo := infraLog.NewMemoryRepository()
	ctx := context.Background()

	rec := &domainLog.CompileLogRecord{
		FormulaHash:   "hash123",
		Formula:       "ROE > 15",
		Success:       true,
		CompileTimeMs: 5,
		CreatedAt:     time.Now(),
	}

	err := repo.Save(ctx, rec)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	if rec.ID == 0 {
		t.Error("expected non-zero ID after save")
	}
}
