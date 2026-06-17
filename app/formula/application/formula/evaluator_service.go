package formula

import (
	"context"
	"fmt"
	"time"

	domainEval "github.com/agoXQ/QuantLab/app/formula/domain/evaluator"
	infraEval "github.com/agoXQ/QuantLab/app/formula/infrastructure/evaluator"
)

// EvaluatorService is the application-level use case that turns a DSL
// formula into a concrete selection / ranking / value map for a stock
// universe at a given cross-section date.
//
// It is a thin orchestrator: compile via the existing Service decorator
// chain, then dispatch the resulting plan into the AST evaluator. Keeping
// it next to Service rather than inside it avoids forcing the decorators
// (cache / log / event / metrics) to evolve every time a new use case
// surfaces.
type EvaluatorService interface {
	// Evaluate compiles the formula end-to-end and runs it on the universe.
	Evaluate(ctx context.Context, req EvaluateRequest) (*EvaluateResult, error)
}

// EvaluateRequest is the application-layer input for an evaluation.
type EvaluateRequest struct {
	Formula      string
	Universe     []string
	AsOfDate     time.Time
	LookbackBars int
	DataVersion  string
	// DataPort is required at request scope because the same EvaluatorService
	// instance may serve live, replayed, or test traffic. Implementations that
	// always use the same port can wrap a default in a thin adapter.
	DataPort domainEval.DataPort
}

// EvaluateResult mirrors the domain Result shape with the formula hash so
// callers can correlate the response with cache entries and audit events.
type EvaluateResult struct {
	FormulaHash string
	Result      *domainEval.Result
}

type evaluatorService struct {
	service   Service
	evaluator domainEval.Evaluator
}

// NewEvaluatorService composes an EvaluatorService from a compiled Service
// (typically the fully-decorated one returned by NewService + decorators)
// and the AST evaluator.
func NewEvaluatorService(svc Service, eval domainEval.Evaluator) EvaluatorService {
	return &evaluatorService{service: svc, evaluator: eval}
}

func (s *evaluatorService) Evaluate(ctx context.Context, req EvaluateRequest) (*EvaluateResult, error) {
	if req.Formula == "" {
		return nil, fmt.Errorf("evaluator_service: formula is required")
	}
	if req.DataPort == nil {
		return nil, fmt.Errorf("evaluator_service: data port is required")
	}
	if req.AsOfDate.IsZero() {
		req.AsOfDate = time.Now()
	}

	compiled, err := s.service.Compile(ctx, req.Formula)
	if err != nil {
		return nil, fmt.Errorf("evaluator_service: compile: %w", err)
	}
	if !compiled.Valid {
		return nil, fmt.Errorf("evaluator_service: invalid formula [%d]: %s", compiled.ErrorCode, compiled.ErrorMsg)
	}
	if compiled.Plan == nil {
		return nil, fmt.Errorf("evaluator_service: compiled plan is nil")
	}

	evalCtx := infraEval.WithDataPort(ctx, req.DataPort)
	result, err := s.evaluator.Evaluate(evalCtx, compiled.Plan, domainEval.Request{
		Universe:     req.Universe,
		AsOfDate:     req.AsOfDate,
		LookbackBars: req.LookbackBars,
		DataVersion:  req.DataVersion,
	})
	if err != nil {
		return nil, err
	}

	return &EvaluateResult{
		FormulaHash: s.service.FormulaHash(req.Formula),
		Result:      result,
	}, nil
}
