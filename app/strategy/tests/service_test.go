package tests

import (
	"context"
	"errors"
	"testing"

	appStrategy "github.com/agoXQ/QuantLab/app/strategy/application/strategy"
	stratErr "github.com/agoXQ/QuantLab/app/strategy/domain/errors"
	domevent "github.com/agoXQ/QuantLab/app/strategy/domain/event"
	"github.com/agoXQ/QuantLab/app/strategy/domain/valueobject"
)

// TestEndToEnd walks the canonical Create -> CreateVersion -> Publish
// -> Fork -> Archive happy path and checks the events lined up.
func TestEndToEnd(t *testing.T) {
	fx := newFixture(t)
	ctx := context.Background()

	created, err := fx.svc.Create(ctx, appStrategy.CreateRequest{
		AuthorID: 7,
		Title:    "Mean reversion",
		Tags:     []string{"meanrev", " factor "},
		Visibility: valueobject.VisibilityPrivate,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if created.Strategy.ID == 0 {
		t.Fatalf("expected strategy ID assigned")
	}
	if got := len(created.Strategy.Tags); got != 2 {
		t.Fatalf("expected 2 canonical tags, got %d (%v)", got, created.Strategy.Tags)
	}

	verRes, err := fx.svc.CreateVersion(ctx, appStrategy.CreateVersionRequest{
		StrategyID:  created.Strategy.ID,
		CallerID:    7,
		FormulaText: "ROE > 0",
		ChangeLog:   "initial cut",
	})
	if err != nil {
		t.Fatalf("CreateVersion: %v", err)
	}
	if verRes.Strategy.Status != valueobject.LifecycleStatusConfigured {
		t.Fatalf("expected CONFIGURED after first version, got %s", verRes.Strategy.Status)
	}

	pub, err := fx.svc.Publish(ctx, appStrategy.PublishRequest{
		StrategyID: created.Strategy.ID,
		CallerID:   7,
	})
	if err != nil {
		t.Fatalf("Publish: %v", err)
	}
	if pub.Status != valueobject.LifecycleStatusPublished {
		t.Fatalf("expected PUBLISHED, got %s", pub.Status)
	}
	if pub.Visibility != valueobject.VisibilityPublic {
		t.Fatalf("expected PUBLIC after publish, got %s", pub.Visibility)
	}

	forked, err := fx.svc.Fork(ctx, appStrategy.ForkRequest{
		SourceStrategyID: created.Strategy.ID,
		CallerID:         9,
	})
	if err != nil {
		t.Fatalf("Fork: %v", err)
	}
	if forked.Strategy.AuthorID != 9 {
		t.Fatalf("expected fork owner 9, got %d", forked.Strategy.AuthorID)
	}
	if forked.Strategy.SourceStrategyID != created.Strategy.ID {
		t.Fatalf("expected source pointer, got %d", forked.Strategy.SourceStrategyID)
	}
	if forked.Strategy.Status != valueobject.LifecycleStatusConfigured {
		t.Fatalf("expected fork to start CONFIGURED, got %s", forked.Strategy.Status)
	}

	// fork count on source must increment.
	src, err := fx.svc.Get(ctx, created.Strategy.ID)
	if err != nil {
		t.Fatalf("Get source: %v", err)
	}
	if src.ForkCount != 1 {
		t.Fatalf("expected ForkCount=1 on source, got %d", src.ForkCount)
	}

	if _, err := fx.svc.Archive(ctx, appStrategy.ArchiveRequest{
		StrategyID: created.Strategy.ID,
		CallerID:   7,
	}); err != nil {
		t.Fatalf("Archive: %v", err)
	}

	// Events emitted in order: Created, VersionCreated, Published,
	// Forked (+ Updated for fork attach, Created for fork target,
	// Forked / VersionCreated combos), Archived. We assert at least
	// one of each canonical type lands so a refactor cannot
	// silently drop wiring.
	types := collectTypes(fx.publisher.Snapshot())
	for _, want := range []domevent.EventType{
		domevent.EventStrategyCreated,
		domevent.EventStrategyVersionCreated,
		domevent.EventStrategyPublished,
		domevent.EventStrategyForked,
		domevent.EventStrategyArchived,
	} {
		if _, ok := types[want]; !ok {
			t.Errorf("missing event %s, saw %v", want, types)
		}
	}
}

// TestPrivateForkRejected guards the visibility gate around Fork; the
// platform refuses to clone a private strategy.
func TestPrivateForkRejected(t *testing.T) {
	fx := newFixture(t)
	ctx := context.Background()

	created, err := fx.svc.Create(ctx, appStrategy.CreateRequest{
		AuthorID: 1,
		Title:    "secret",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if _, err := fx.svc.CreateVersion(ctx, appStrategy.CreateVersionRequest{
		StrategyID:  created.Strategy.ID,
		CallerID:    1,
		FormulaText: "ROE>0",
	}); err != nil {
		t.Fatalf("CreateVersion: %v", err)
	}

	_, err = fx.svc.Fork(ctx, appStrategy.ForkRequest{
		SourceStrategyID: created.Strategy.ID,
		CallerID:         2,
	})
	if !errors.Is(err, stratErr.ErrNotOwner) {
		t.Fatalf("expected ErrNotOwner forking private strategy, got %v", err)
	}
}

// TestUpdateRequiresOwner makes sure the authorise() guard fires across
// the application service, not just inside a single use case.
func TestUpdateRequiresOwner(t *testing.T) {
	fx := newFixture(t)
	ctx := context.Background()
	created, err := fx.svc.Create(ctx, appStrategy.CreateRequest{AuthorID: 1, Title: "t"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	title := "renamed"
	if _, err := fx.svc.Update(ctx, appStrategy.UpdateRequest{
		StrategyID: created.Strategy.ID,
		CallerID:   42,
		Title:      &title,
	}); !errors.Is(err, stratErr.ErrNotOwner) {
		t.Fatalf("expected ErrNotOwner, got %v", err)
	}
}

func collectTypes(events []domevent.Event) map[domevent.EventType]int {
	out := map[domevent.EventType]int{}
	for _, e := range events {
		out[e.EventType]++
	}
	return out
}
