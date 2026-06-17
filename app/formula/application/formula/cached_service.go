package formula

import (
	"context"
	"encoding/json"
	"time"

	domainAST "github.com/agoXQ/QuantLab/app/formula/domain/ast"
	domainCache "github.com/agoXQ/QuantLab/app/formula/domain/cache"
	domainCompiler "github.com/agoXQ/QuantLab/app/formula/domain/compiler"
	domainFunc "github.com/agoXQ/QuantLab/app/formula/domain/function"
	domainValidator "github.com/agoXQ/QuantLab/app/formula/domain/validator"
)

const (
	cacheKeyPrefixAST  = "formula:ast:"
	cacheKeyPrefixPlan = "formula:plan:"
	cacheKeyPrefixVal  = "formula:validate:"

	defaultCacheTTL = 24 * time.Hour
)

// CachedService wraps a Service with a Cache layer.
type CachedService struct {
	inner   Service
	cache   domainCache.Cache
	ttl     time.Duration
	metrics MetricsCollector
}

// NewCachedService creates a new cached service decorator.
func NewCachedService(inner Service, cache domainCache.Cache, ttl time.Duration) Service {
	if ttl == 0 {
		ttl = defaultCacheTTL
	}
	return &CachedService{
		inner:   inner,
		cache:   cache,
		ttl:     ttl,
		metrics: &noopMetricsCollector{},
	}
}

// WithMetrics sets the metrics collector on the cached service.
func (s *CachedService) WithMetrics(m MetricsCollector) *CachedService {
	s.metrics = m
	return s
}

func (s *CachedService) Validate(ctx context.Context, formula string) (*domainValidator.ValidationResult, error) {
	hash := s.inner.FormulaHash(formula)
	key := cacheKeyPrefixVal + hash

	if cached, err := s.cache.Get(ctx, key); err == nil && cached != nil {
		var result domainValidator.ValidationResult
		if err := json.Unmarshal(cached, &result); err == nil {
			s.metrics.IncCacheHit()
			return &result, nil
		}
	}

	s.metrics.IncCacheMiss()
	result, err := s.inner.Validate(ctx, formula)
	if err != nil {
		return result, err
	}

	if data, marshalErr := json.Marshal(result); marshalErr == nil {
		_ = s.cache.Set(ctx, key, data, s.ttl)
	}

	return result, nil
}

func (s *CachedService) Compile(ctx context.Context, formula string) (*CompileResult, error) {
	hash := s.inner.FormulaHash(formula)
	planKey := cacheKeyPrefixPlan + hash
	astKey := cacheKeyPrefixAST + hash

	if cached, err := s.cache.Get(ctx, planKey); err == nil && cached != nil {
		var plan domainCompiler.ExecutionPlan
		if err := json.Unmarshal(cached, &plan); err == nil {
			result := &CompileResult{
				Plan:  &plan,
				Valid: true,
			}
			if astCached, astErr := s.cache.Get(ctx, astKey); astErr == nil && astCached != nil {
				var astNode domainAST.Node
				astNode = &domainAST.BinaryExpression{}
				if err := json.Unmarshal(astCached, astNode); err == nil {
					result.AST = astNode
				}
			}
			s.metrics.IncCacheHit()
			return result, nil
		}
	}

	s.metrics.IncCacheMiss()
	result, err := s.inner.Compile(ctx, formula)
	if err != nil {
		return result, err
	}
	if result == nil {
		return result, nil
	}

	if result.Plan != nil {
		if data, marshalErr := json.Marshal(result.Plan); marshalErr == nil {
			_ = s.cache.Set(ctx, planKey, data, s.ttl)
		}
	}
	if result.AST != nil {
		if data, marshalErr := json.Marshal(result.AST); marshalErr == nil {
			_ = s.cache.Set(ctx, astKey, data, s.ttl)
		}
	}

	return result, nil
}

func (s *CachedService) GetAST(ctx context.Context, formula string) (domainAST.Node, error) {
	hash := s.inner.FormulaHash(formula)
	key := cacheKeyPrefixAST + hash

	if cached, err := s.cache.Get(ctx, key); err == nil && cached != nil {
		var typeOnly struct {
			Type string `json:"type"`
		}
		if err := json.Unmarshal(cached, &typeOnly); err == nil {
			if node := unmarshalNodeByType(typeOnly.Type, cached); node != nil {
				s.metrics.IncCacheHit()
				return node, nil
			}
		}
	}

	s.metrics.IncCacheMiss()
	node, err := s.inner.GetAST(ctx, formula)
	if err != nil {
		return node, err
	}

	if data, marshalErr := json.Marshal(node); marshalErr == nil {
		_ = s.cache.Set(ctx, key, data, s.ttl)
	}

	return node, nil
}

func (s *CachedService) ListFunctions(ctx context.Context) ([]domainFunc.FunctionDefinition, error) {
	return s.inner.ListFunctions(ctx)
}

func (s *CachedService) GetFunction(ctx context.Context, name string) (*domainFunc.FunctionDefinition, error) {
	return s.inner.GetFunction(ctx, name)
}

func (s *CachedService) FormulaHash(formula string) string {
	return s.inner.FormulaHash(formula)
}

// unmarshalNodeByType unmarshals JSON into the correct AST node type.
func unmarshalNodeByType(typeName string, data []byte) domainAST.Node {
	var node domainAST.Node
	switch typeName {
	case "BinaryExpression":
		node = &domainAST.BinaryExpression{}
	case "UnaryExpression":
		node = &domainAST.UnaryExpression{}
	case "FunctionCall":
		node = &domainAST.FunctionCall{}
	case "Identifier":
		node = &domainAST.Identifier{}
	case "NumberLiteral":
		node = &domainAST.NumberLiteral{}
	case "StringLiteral":
		node = &domainAST.StringLiteral{}
	case "BoolLiteral":
		node = &domainAST.BoolLiteral{}
	default:
		return nil
	}
	if err := json.Unmarshal(data, node); err != nil {
		return nil
	}
	return node
}

// noopMetricsCollector is the default no-op metrics collector.
type noopMetricsCollector struct{}

func (m *noopMetricsCollector) IncValidate()              {}
func (m *noopMetricsCollector) IncCompile()               {}
func (m *noopMetricsCollector) IncCompileFail()            {}
func (m *noopMetricsCollector) IncCacheHit()               {}
func (m *noopMetricsCollector) IncCacheMiss()              {}
func (m *noopMetricsCollector) ObserveValidateLatency(float64) {}
func (m *noopMetricsCollector) ObserveCompileLatency(float64)  {}

var _ MetricsCollector = (*noopMetricsCollector)(nil)
var _ Service = (*CachedService)(nil)
