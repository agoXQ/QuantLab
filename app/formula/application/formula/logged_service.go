package formula

import (
	"context"
	"time"

	domainAST "github.com/agoXQ/QuantLab/app/formula/domain/ast"
	domainLog "github.com/agoXQ/QuantLab/app/formula/domain/compilelog"
	domainFunc "github.com/agoXQ/QuantLab/app/formula/domain/function"
	domainValidator "github.com/agoXQ/QuantLab/app/formula/domain/validator"
)

// LoggedService wraps a Service with compile log persistence.
// It implements the same Service interface, so it can be used as a drop-in replacement.
type LoggedService struct {
	inner  Service
	logRepo domainLog.Repository
}

// NewLoggedService creates a new logged service decorator.
func NewLoggedService(inner Service, logRepo domainLog.Repository) Service {
	return &LoggedService{
		inner:   inner,
		logRepo: logRepo,
	}
}

func (s *LoggedService) Validate(ctx context.Context, formula string) (*domainValidator.ValidationResult, error) {
	return s.inner.Validate(ctx, formula)
}

func (s *LoggedService) Compile(ctx context.Context, formula string) (*CompileResult, error) {
	start := time.Now()

	result, err := s.inner.Compile(ctx, formula)

	elapsedMs := int(time.Since(start).Milliseconds())

	// Record compile log (best-effort, non-blocking)
	if result != nil {
		hash := s.inner.FormulaHash(formula)
		record := &domainLog.CompileLogRecord{
			FormulaHash:   hash,
			Formula:       formula,
			Success:       result.Valid,
			ErrorCode:     result.ErrorCode,
			CompileTimeMs: elapsedMs,
			CreatedAt:     time.Now(),
		}
		_ = s.logRepo.Save(ctx, record)
	}

	return result, err
}

func (s *LoggedService) GetAST(ctx context.Context, formula string) (domainAST.Node, error) {
	return s.inner.GetAST(ctx, formula)
}

func (s *LoggedService) ListFunctions(ctx context.Context) ([]domainFunc.FunctionDefinition, error) {
	return s.inner.ListFunctions(ctx)
}

func (s *LoggedService) GetFunction(ctx context.Context, name string) (*domainFunc.FunctionDefinition, error) {
	return s.inner.GetFunction(ctx, name)
}

func (s *LoggedService) FormulaHash(formula string) string {
	return s.inner.FormulaHash(formula)
}

// Ensure LoggedService implements Service.
var _ Service = (*LoggedService)(nil)
