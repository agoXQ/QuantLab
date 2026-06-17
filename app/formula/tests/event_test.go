package tests

import (
	"context"
	"testing"

	appFormula "github.com/agoXQ/QuantLab/app/formula/application/formula"
	domainEvent "github.com/agoXQ/QuantLab/app/formula/domain/event"
	infraEvent "github.com/agoXQ/QuantLab/app/formula/infrastructure/event"
	infraFunc "github.com/agoXQ/QuantLab/app/formula/infrastructure/function"
	infraLexer "github.com/agoXQ/QuantLab/app/formula/infrastructure/lexer"
	infraOptimizer "github.com/agoXQ/QuantLab/app/formula/infrastructure/optimizer"
	infraParser "github.com/agoXQ/QuantLab/app/formula/infrastructure/parser"
	infraPlanner "github.com/agoXQ/QuantLab/app/formula/infrastructure/planner"
	infraValidator "github.com/agoXQ/QuantLab/app/formula/infrastructure/validator"
	infraVar "github.com/agoXQ/QuantLab/app/formula/infrastructure/variable"
)

func newEventingService() (appFormula.Service, *infraEvent.MemoryPublisher) {
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

	publisher := infraEvent.NewMemoryPublisher()
	svc := appFormula.NewEventingService(base, publisher)
	return svc, publisher
}

func TestEventing_PublishesFormulaValidatedOnValidate(t *testing.T) {
	svc, pub := newEventingService()
	ctx := context.Background()

	result, err := svc.Validate(ctx, "ROE > 15")
	if err != nil {
		t.Fatalf("Validate failed: %v", err)
	}
	if !result.Valid {
		t.Fatal("expected valid")
	}

	events := pub.Published()
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	evt := events[0]
	if evt.EventType != domainEvent.EventFormulaValidated {
		t.Errorf("expected FormulaValidated, got %s", evt.EventType)
	}
	if evt.Producer != domainEvent.ProducerFormulaEngine {
		t.Errorf("expected producer %s, got %s", domainEvent.ProducerFormulaEngine, evt.Producer)
	}
	if evt.EventVersion != "1.0" {
		t.Errorf("expected event version 1.0, got %s", evt.EventVersion)
	}
	if evt.AggregateType != domainEvent.AggregateTypeStrategy {
		t.Errorf("expected aggregate type STRATEGY, got %s", evt.AggregateType)
	}
	if evt.EventID == "" {
		t.Error("expected non-empty event_id")
	}

	payload, ok := evt.Payload.(domainEvent.FormulaValidatedPayload)
	if !ok {
		t.Fatalf("expected FormulaValidatedPayload, got %T", evt.Payload)
	}
	if !payload.Valid {
		t.Error("expected payload.valid = true")
	}
	if payload.FormulaHash == "" {
		t.Error("expected non-empty formula_hash")
	}
}

func TestEventing_PublishesFormulaValidatedOnInvalidValidate(t *testing.T) {
	svc, pub := newEventingService()
	ctx := context.Background()

	result, err := svc.Validate(ctx, "UNKNOWN_VAR > 15")
	if err != nil {
		t.Fatalf("Validate failed: %v", err)
	}
	if result.Valid {
		t.Fatal("expected invalid")
	}

	events := pub.Published()
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	evt := events[0]
	if evt.EventType != domainEvent.EventFormulaValidated {
		t.Errorf("expected FormulaValidated, got %s", evt.EventType)
	}

	payload, ok := evt.Payload.(domainEvent.FormulaValidatedPayload)
	if !ok {
		t.Fatalf("expected FormulaValidatedPayload, got %T", evt.Payload)
	}
	if payload.Valid {
		t.Error("expected payload.valid = false")
	}
}

func TestEventing_PublishesFormulaCompileFailedOnInvalidCompile(t *testing.T) {
	svc, pub := newEventingService()
	ctx := context.Background()

	result, err := svc.Compile(ctx, "UNKNOWN_VAR > 15")
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}
	if result.Valid {
		t.Fatal("expected invalid")
	}

	events := pub.Published()
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	evt := events[0]
	if evt.EventType != domainEvent.EventFormulaCompileFailed {
		t.Errorf("expected FormulaCompileFailed, got %s", evt.EventType)
	}

	payload, ok := evt.Payload.(domainEvent.FormulaCompileFailedPayload)
	if !ok {
		t.Fatalf("expected FormulaCompileFailedPayload, got %T", evt.Payload)
	}
	if payload.ErrorCode == 0 {
		t.Error("expected non-zero error_code")
	}
	if payload.Error == "" {
		t.Error("expected non-empty error message")
	}
}

func TestEventing_DoesNotPublishOnSuccessfulCompile(t *testing.T) {
	svc, pub := newEventingService()
	ctx := context.Background()

	result, err := svc.Compile(ctx, "ROE > 15")
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}
	if !result.Valid {
		t.Fatal("expected valid")
	}

	events := pub.Published()
	if len(events) != 0 {
		t.Errorf("expected 0 events for successful compile, got %d", len(events))
	}
}

func TestEventing_PublishesEventWithCorrectHash(t *testing.T) {
	svc, pub := newEventingService()
	ctx := context.Background()

	formula := "ROE > 15"
	expectedHash := svc.FormulaHash(formula)

	_, err := svc.Validate(ctx, formula)
	if err != nil {
		t.Fatalf("Validate failed: %v", err)
	}

	events := pub.Published()
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	payload, ok := events[0].Payload.(domainEvent.FormulaValidatedPayload)
	if !ok {
		t.Fatalf("expected FormulaValidatedPayload, got %T", events[0].Payload)
	}
	if payload.FormulaHash != expectedHash {
		t.Errorf("expected hash %s, got %s", expectedHash, payload.FormulaHash)
	}
}

func TestEventing_EventHasAllRequiredFields(t *testing.T) {
	svc, pub := newEventingService()
	ctx := context.Background()

	_, err := svc.Validate(ctx, "ROE > 15")
	if err != nil {
		t.Fatalf("Validate failed: %v", err)
	}

	events := pub.Published()
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	evt := events[0]
	if evt.EventID == "" {
		t.Error("missing event_id")
	}
	if evt.EventType == "" {
		t.Error("missing event_type")
	}
	if evt.EventVersion == "" {
		t.Error("missing event_version")
	}
	if evt.OccurredAt.IsZero() {
		t.Error("missing occurred_at")
	}
	if evt.AggregateType == "" {
		t.Error("missing aggregate_type")
	}
	if evt.Producer == "" {
		t.Error("missing producer")
	}
	if evt.Payload == nil {
		t.Error("missing payload")
	}
}

func TestEventing_MultipleEvents(t *testing.T) {
	svc, pub := newEventingService()
	ctx := context.Background()

	_, _ = svc.Validate(ctx, "ROE > 15")
	_, _ = svc.Validate(ctx, "PE < 20")
	_, _ = svc.Compile(ctx, "UNKNOWN_VAR > 15")

	events := pub.Published()
	if len(events) != 3 {
		t.Fatalf("expected 3 events, got %d", len(events))
	}

	if events[0].EventType != domainEvent.EventFormulaValidated {
		t.Errorf("expected event 0 to be FormulaValidated, got %s", events[0].EventType)
	}
	if events[1].EventType != domainEvent.EventFormulaValidated {
		t.Errorf("expected event 1 to be FormulaValidated, got %s", events[1].EventType)
	}
	if events[2].EventType != domainEvent.EventFormulaCompileFailed {
		t.Errorf("expected event 2 to be FormulaCompileFailed, got %s", events[2].EventType)
	}
}

func TestMemoryPublisher_Reset(t *testing.T) {
	pub := infraEvent.NewMemoryPublisher()

	_ = pub.Publish(domainEvent.Event{EventType: domainEvent.EventFormulaValidated})
	if len(pub.Published()) != 1 {
		t.Fatal("expected 1 event before reset")
	}

	pub.Reset()
	if len(pub.Published()) != 0 {
		t.Fatal("expected 0 events after reset")
	}
}

func TestEventing_GetASTDoesNotPublish(t *testing.T) {
	svc, pub := newEventingService()
	ctx := context.Background()

	_, err := svc.GetAST(ctx, "ROE > 15")
	if err != nil {
		t.Fatalf("GetAST failed: %v", err)
	}

	events := pub.Published()
	if len(events) != 0 {
		t.Errorf("expected 0 events from GetAST, got %d", len(events))
	}
}

func TestEventing_ListFunctionsDoesNotPublish(t *testing.T) {
	svc, pub := newEventingService()
	ctx := context.Background()

	_, err := svc.ListFunctions(ctx)
	if err != nil {
		t.Fatalf("ListFunctions failed: %v", err)
	}

	events := pub.Published()
	if len(events) != 0 {
		t.Errorf("expected 0 events from ListFunctions, got %d", len(events))
	}
}
