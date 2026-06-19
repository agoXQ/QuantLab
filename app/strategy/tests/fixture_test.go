package tests

import (
	"context"
	"sync"
	"testing"
	"time"

	appStrategy "github.com/agoXQ/QuantLab/app/strategy/application/strategy"
	domevent "github.com/agoXQ/QuantLab/app/strategy/domain/event"
	infraEvent "github.com/agoXQ/QuantLab/app/strategy/infrastructure/event"
	infraMemory "github.com/agoXQ/QuantLab/app/strategy/infrastructure/repository/memory"
)

// recordingPublisher is a Publisher decorator that captures every event
// the application service emits so tests can assert on event flow
// without spinning up Kafka. Falls back to the no-op publisher when no
// recorder is attached.
type recordingPublisher struct {
	mu     sync.Mutex
	events []domevent.Event
}

func newRecordingPublisher() *recordingPublisher { return &recordingPublisher{} }

func (p *recordingPublisher) Publish(_ context.Context, e domevent.Event) error {
	p.mu.Lock()
	p.events = append(p.events, e)
	p.mu.Unlock()
	return nil
}

func (p *recordingPublisher) Snapshot() []domevent.Event {
	p.mu.Lock()
	defer p.mu.Unlock()
	out := make([]domevent.Event, len(p.events))
	copy(out, p.events)
	return out
}

// fixture wires the same dependency graph the production servicecontext
// builds, but with the in-memory repositories so the test stays
// sandbox-friendly. The recording publisher is attached by default; the
// noop fallback below stays available for tests that do not care about
// events.
type fixture struct {
	svc       appStrategy.Service
	publisher *recordingPublisher
	clock     time.Time
}

func newFixture(t *testing.T) *fixture {
	t.Helper()
	clock := time.Date(2024, 6, 28, 15, 0, 0, 0, time.UTC)
	now := func() time.Time { return clock }
	publisher := newRecordingPublisher()
	svc := appStrategy.NewService(appStrategy.Dependencies{
		Strategies: infraMemory.NewStrategyRepository(),
		Versions:   infraMemory.NewVersionRepository(),
		Forks:      infraMemory.NewForkRepository(),
		Publisher:  publisher,
		Clock:      now,
	})
	return &fixture{svc: svc, publisher: publisher, clock: clock}
}

// silence: the Noop publisher is referenced from servicecontext.go, but
// we keep an import-side reference here so accidental refactors that
// drop the package surface fail loudly during the tests.
var _ domevent.Publisher = infraEvent.Noop{}
