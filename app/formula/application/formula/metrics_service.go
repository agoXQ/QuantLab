package formula

import (
	"context"
	"time"

	domainAST "github.com/agoXQ/QuantLab/app/formula/domain/ast"
	domainFunc "github.com/agoXQ/QuantLab/app/formula/domain/function"
	domainValidator "github.com/agoXQ/QuantLab/app/formula/domain/validator"
)

// MetricsCollector defines the interface for collecting metrics.
// Implementations should be safe for concurrent use.
type MetricsCollector interface {
	// IncValidate increments the validate counter.
	IncValidate()
	// IncCompile increments the compile counter.
	IncCompile()
	// IncCompileFail increments the compile failure counter.
	IncCompileFail()
	// IncCacheHit increments the cache hit counter.
	IncCacheHit()
	// IncCacheMiss increments the cache miss counter.
	IncCacheMiss()
	// ObserveValidateLatency records the latency of a validate operation in milliseconds.
	ObserveValidateLatency(ms float64)
	// ObserveCompileLatency records the latency of a compile operation in milliseconds.
	ObserveCompileLatency(ms float64)
}

// MetricsService wraps a Service with metrics collection.
type MetricsService struct {
	inner   Service
	metrics MetricsCollector
}

// NewMetricsService creates a new metrics-collecting service decorator.
func NewMetricsService(inner Service, metrics MetricsCollector) Service {
	return &MetricsService{
		inner:   inner,
		metrics: metrics,
	}
}

func (s *MetricsService) Validate(ctx context.Context, formula string) (*domainValidator.ValidationResult, error) {
	start := time.Now()
	s.metrics.IncValidate()
	result, err := s.inner.Validate(ctx, formula)
	s.metrics.ObserveValidateLatency(float64(time.Since(start).Milliseconds()))
	return result, err
}

func (s *MetricsService) Compile(ctx context.Context, formula string) (*CompileResult, error) {
	start := time.Now()
	s.metrics.IncCompile()
	result, err := s.inner.Compile(ctx, formula)
	if err == nil && result != nil && !result.Valid {
		s.metrics.IncCompileFail()
	}
	if err != nil {
		return result, err
	}
	s.metrics.ObserveCompileLatency(float64(time.Since(start).Milliseconds()))
	return result, err
}

func (s *MetricsService) GetAST(ctx context.Context, formula string) (domainAST.Node, error) {
	return s.inner.GetAST(ctx, formula)
}

func (s *MetricsService) ListFunctions(ctx context.Context) ([]domainFunc.FunctionDefinition, error) {
	return s.inner.ListFunctions(ctx)
}

func (s *MetricsService) GetFunction(ctx context.Context, name string) (*domainFunc.FunctionDefinition, error) {
	return s.inner.GetFunction(ctx, name)
}

func (s *MetricsService) FormulaHash(formula string) string {
	return s.inner.FormulaHash(formula)
}

// Ensure MetricsService implements Service.
var _ Service = (*MetricsService)(nil)
