package event

import (
	"sync"

	domainEvent "github.com/agoXQ/QuantLab/app/formula/domain/event"
)

type MemoryPublisher struct {
	mu     sync.RWMutex
	Events []domainEvent.Event
}

// NewMemoryPublisher creates a new in-memory event publisher (for testing).
func NewMemoryPublisher() *MemoryPublisher {
	return &MemoryPublisher{
		Events: make([]domainEvent.Event, 0),
	}
}

func (p *MemoryPublisher) Publish(event domainEvent.Event) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Events = append(p.Events, event)
	return nil
}

// Published returns a copy of all published events.
func (p *MemoryPublisher) Published() []domainEvent.Event {
	p.mu.RLock()
	defer p.mu.RUnlock()
	cp := make([]domainEvent.Event, len(p.Events))
	copy(cp, p.Events)
	return cp
}

// Reset clears all published events.
func (p *MemoryPublisher) Reset() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Events = nil
}

// Ensure memoryPublisher implements Publisher.
var _ domainEvent.Publisher = (*MemoryPublisher)(nil)
