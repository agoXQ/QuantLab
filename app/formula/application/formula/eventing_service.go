package formula

import (
	"context"
	"time"

	"github.com/google/uuid"

	domainAST "github.com/agoXQ/QuantLab/app/formula/domain/ast"
	domainEvent "github.com/agoXQ/QuantLab/app/formula/domain/event"
	domainFunc "github.com/agoXQ/QuantLab/app/formula/domain/function"
	domainValidator "github.com/agoXQ/QuantLab/app/formula/domain/validator"
)

// EventingService wraps a Service with domain event publishing.
// It implements the same Service interface, so it can be used as a drop-in replacement.
type EventingService struct {
	inner     Service
	publisher domainEvent.Publisher
}

// NewEventingService creates a new eventing service decorator.
func NewEventingService(inner Service, publisher domainEvent.Publisher) Service {
	return &EventingService{
		inner:     inner,
		publisher: publisher,
	}
}

func (s *EventingService) Validate(ctx context.Context, formula string) (*domainValidator.ValidationResult, error) {
	result, err := s.inner.Validate(ctx, formula)
	if err != nil {
		return result, err
	}

	// Publish FormulaValidated event (best-effort)
	hash := s.inner.FormulaHash(formula)
	evt := domainEvent.Event{
		EventID:       uuid.New().String(),
		EventType:     domainEvent.EventFormulaValidated,
		EventVersion:  "1.0",
		OccurredAt:    time.Now(),
		AggregateType: domainEvent.AggregateTypeStrategy,
		Producer:      domainEvent.ProducerFormulaEngine,
		Payload: domainEvent.FormulaValidatedPayload{
			FormulaHash: hash,
			Valid:       result.Valid,
		},
	}
	_ = s.publisher.Publish(evt)

	return result, nil
}

func (s *EventingService) Compile(ctx context.Context, formula string) (*CompileResult, error) {
	result, err := s.inner.Compile(ctx, formula)
	if err != nil {
		return result, err
	}

	if result != nil && !result.Valid {
		// Publish FormulaCompileFailed event (best-effort)
		hash := s.inner.FormulaHash(formula)
		evt := domainEvent.Event{
			EventID:       uuid.New().String(),
			EventType:     domainEvent.EventFormulaCompileFailed,
			EventVersion:  "1.0",
			OccurredAt:    time.Now(),
			AggregateType: domainEvent.AggregateTypeStrategy,
			Producer:      domainEvent.ProducerFormulaEngine,
			Payload: domainEvent.FormulaCompileFailedPayload{
				FormulaHash: hash,
				Error:       result.ErrorMsg,
				ErrorCode:   result.ErrorCode,
			},
		}
		_ = s.publisher.Publish(evt)
	}

	return result, err
}

func (s *EventingService) GetAST(ctx context.Context, formula string) (domainAST.Node, error) {
	return s.inner.GetAST(ctx, formula)
}

func (s *EventingService) ListFunctions(ctx context.Context) ([]domainFunc.FunctionDefinition, error) {
	return s.inner.ListFunctions(ctx)
}

func (s *EventingService) GetFunction(ctx context.Context, name string) (*domainFunc.FunctionDefinition, error) {
	return s.inner.GetFunction(ctx, name)
}

func (s *EventingService) FormulaHash(formula string) string {
	return s.inner.FormulaHash(formula)
}

// Ensure EventingService implements Service.
var _ Service = (*EventingService)(nil)
