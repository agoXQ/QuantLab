package tests

import (
	"context"
	"testing"

	appFormula "github.com/agoXQ/QuantLab/app/formula/application/formula"
	infraFunc "github.com/agoXQ/QuantLab/app/formula/infrastructure/function"
	infraLexer "github.com/agoXQ/QuantLab/app/formula/infrastructure/lexer"
	infraMetrics "github.com/agoXQ/QuantLab/app/formula/infrastructure/metrics"
	infraOptimizer "github.com/agoXQ/QuantLab/app/formula/infrastructure/optimizer"
	infraParser "github.com/agoXQ/QuantLab/app/formula/infrastructure/parser"
	infraPlanner "github.com/agoXQ/QuantLab/app/formula/infrastructure/planner"
	infraValidator "github.com/agoXQ/QuantLab/app/formula/infrastructure/validator"
	infraVar "github.com/agoXQ/QuantLab/app/formula/infrastructure/variable"
)

// collectMetrics implements MetricsCollector for testing.
type collectMetrics struct {
	validateTotal    int
	compileTotal     int
	compileFailTotal int
	cacheHitTotal    int
	cacheMissTotal   int
	validateLatency  []float64
	compileLatency   []float64
}

func (m *collectMetrics) IncValidate()    { m.validateTotal++ }
func (m *collectMetrics) IncCompile()     { m.compileTotal++ }
func (m *collectMetrics) IncCompileFail() { m.compileFailTotal++ }
func (m *collectMetrics) IncCacheHit()    { m.cacheHitTotal++ }
func (m *collectMetrics) IncCacheMiss()   { m.cacheMissTotal++ }
func (m *collectMetrics) ObserveValidateLatency(ms float64) { m.validateLatency = append(m.validateLatency, ms) }
func (m *collectMetrics) ObserveCompileLatency(ms float64)  { m.compileLatency = append(m.compileLatency, ms) }

func newMetricsService() (appFormula.Service, *collectMetrics) {
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

	metrics := &collectMetrics{}
	svc := appFormula.NewMetricsService(base, metrics)
	return svc, metrics
}

func TestMetrics_IncrementsValidate(t *testing.T) {
	svc, m := newMetricsService()
	ctx := context.Background()

	_, _ = svc.Validate(ctx, "ROE > 15")
	if m.validateTotal != 1 {
		t.Errorf("expected validate=1, got %d", m.validateTotal)
	}

	_, _ = svc.Validate(ctx, "PE < 20")
	if m.validateTotal != 2 {
		t.Errorf("expected validate=2, got %d", m.validateTotal)
	}
}

func TestMetrics_IncrementsCompile(t *testing.T) {
	svc, m := newMetricsService()
	ctx := context.Background()

	_, _ = svc.Compile(ctx, "ROE > 15")
	if m.compileTotal != 1 {
		t.Errorf("expected compile=1, got %d", m.compileTotal)
	}
}

func TestMetrics_IncrementsCompileFail(t *testing.T) {
	svc, m := newMetricsService()
	ctx := context.Background()

	_, _ = svc.Compile(ctx, "UNKNOWN_VAR > 15")
	if m.compileTotal != 1 {
		t.Errorf("expected compile=1, got %d", m.compileTotal)
	}
	if m.compileFailTotal != 1 {
		t.Errorf("expected compileFail=1, got %d", m.compileFailTotal)
	}
}

func TestMetrics_DoesNotIncrementCompileFailOnSuccess(t *testing.T) {
	svc, m := newMetricsService()
	ctx := context.Background()

	_, _ = svc.Compile(ctx, "ROE > 15")
	if m.compileFailTotal != 0 {
		t.Errorf("expected compileFail=0, got %d", m.compileFailTotal)
	}
}

func TestMetrics_GetASTDoesNotIncrement(t *testing.T) {
	svc, m := newMetricsService()
	ctx := context.Background()

	_, _ = svc.GetAST(ctx, "ROE > 15")
	if m.validateTotal != 0 {
		t.Errorf("expected validate=0, got %d", m.validateTotal)
	}
	if m.compileTotal != 0 {
		t.Errorf("expected compile=0, got %d", m.compileTotal)
	}
}

func TestMetrics_ListFunctionsDoesNotIncrement(t *testing.T) {
	svc, m := newMetricsService()
	ctx := context.Background()

	_, _ = svc.ListFunctions(ctx)
	if m.validateTotal != 0 {
		t.Errorf("expected validate=0, got %d", m.validateTotal)
	}
	if m.compileTotal != 0 {
		t.Errorf("expected compile=0, got %d", m.compileTotal)
	}
}

func TestNoopCollector_DoesNotPanic(t *testing.T) {
	c := infraMetrics.NewNoopCollector()
	c.IncValidate()
	c.IncCompile()
	c.IncCompileFail()
	c.IncCacheHit()
	c.IncCacheMiss()
	c.ObserveValidateLatency(1.5)
	c.ObserveCompileLatency(2.5)
	// If we got here without panic, it passes
}

func TestMetrics_CacheHitMiss(t *testing.T) {
	m := &collectMetrics{}

	m.IncCacheHit()
	if m.cacheHitTotal != 1 {
		t.Errorf("expected cacheHit=1, got %d", m.cacheHitTotal)
	}

	m.IncCacheMiss()
	if m.cacheMissTotal != 1 {
		t.Errorf("expected cacheMiss=1, got %d", m.cacheMissTotal)
	}
}

func TestMetrics_RecordsValidateLatency(t *testing.T) {
	svc, m := newMetricsService()
	ctx := context.Background()

	_, _ = svc.Validate(ctx, "ROE > 15")
	if len(m.validateLatency) != 1 {
		t.Errorf("expected 1 validate latency observation, got %d", len(m.validateLatency))
	}
	if m.validateLatency[0] < 0 {
		t.Errorf("expected non-negative latency, got %f", m.validateLatency[0])
	}
}

func TestMetrics_RecordsCompileLatency(t *testing.T) {
	svc, m := newMetricsService()
	ctx := context.Background()

	_, _ = svc.Compile(ctx, "ROE > 15")
	if len(m.compileLatency) != 1 {
		t.Errorf("expected 1 compile latency observation, got %d", len(m.compileLatency))
	}
	if m.compileLatency[0] < 0 {
		t.Errorf("expected non-negative latency, got %f", m.compileLatency[0])
	}
}

func TestMetrics_RecordsCompileLatencyOnFailure(t *testing.T) {
	svc, m := newMetricsService()
	ctx := context.Background()

	_, _ = svc.Compile(ctx, "UNKNOWN_VAR > 15")
	if len(m.compileLatency) != 1 {
		t.Errorf("expected 1 compile latency observation on failure, got %d", len(m.compileLatency))
	}
}

func TestPrometheusCollector_DoesNotPanic(t *testing.T) {
	c := infraMetrics.NewPrometheusCollector()
	c.IncValidate()
	c.IncCompile()
	c.IncCompileFail()
	c.IncCacheHit()
	c.IncCacheMiss()
	c.ObserveValidateLatency(1.5)
	c.ObserveCompileLatency(2.5)
	// If we got here without panic, it passes
}
