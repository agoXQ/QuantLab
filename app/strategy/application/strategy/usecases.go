package strategy

import (
	"context"
	"fmt"
	"strings"

	stratErr "github.com/agoXQ/QuantLab/app/strategy/domain/errors"
	domevent "github.com/agoXQ/QuantLab/app/strategy/domain/event"
	domstrategy "github.com/agoXQ/QuantLab/app/strategy/domain/strategy"
	"github.com/agoXQ/QuantLab/app/strategy/domain/valueobject"
)

// Create persists a new Strategy in DRAFT.
func (s *service) Create(ctx context.Context, req CreateRequest) (*CreateResult, error) {
	now := s.deps.Clock()
	visibility := req.Visibility
	if visibility == "" {
		visibility = valueobject.VisibilityPrivate
	}
	st := &domstrategy.Strategy{
		AuthorID:    req.AuthorID,
		Title:       strings.TrimSpace(req.Title),
		Description: strings.TrimSpace(req.Description),
		Category:    strings.TrimSpace(req.Category),
		Tags:        domstrategy.NormaliseTags(req.Tags),
		Status:      valueobject.LifecycleStatusDraft,
		Visibility:  visibility,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := st.Validate(); err != nil {
		return nil, err
	}
	if err := s.deps.Strategies.Create(ctx, st); err != nil {
		return nil, err
	}
	s.publish(ctx, domevent.EventStrategyCreated, st.ID, domevent.StrategyCreatedPayload{
		StrategyID: st.ID,
		AuthorID:   st.AuthorID,
		Title:      st.Title,
	})
	return &CreateResult{Strategy: st}, nil
}

// Update applies a metadata patch.
func (s *service) Update(ctx context.Context, req UpdateRequest) (*domstrategy.Strategy, error) {
	st, err := s.deps.Strategies.Get(ctx, req.StrategyID)
	if err != nil {
		return nil, err
	}
	if err := authorise(st, req.CallerID); err != nil {
		return nil, err
	}
	patch := domstrategy.MetaPatch{
		Title:       req.Title,
		Description: req.Description,
		Category:    req.Category,
		Tags:        req.Tags,
		Visibility:  req.Visibility,
	}
	if err := st.UpdateMeta(patch, s.deps.Clock()); err != nil {
		return nil, err
	}
	if err := s.deps.Strategies.Update(ctx, st); err != nil {
		return nil, err
	}
	s.publish(ctx, domevent.EventStrategyUpdated, st.ID, domevent.StrategyUpdatedPayload{
		StrategyID: st.ID,
		Title:      st.Title,
		Tags:       append([]string(nil), st.Tags...),
	})
	return st, nil
}

// Get returns a strategy by ID.
func (s *service) Get(ctx context.Context, id int64) (*domstrategy.Strategy, error) {
	if id <= 0 {
		return nil, stratErr.ErrInvalidStrategy
	}
	return s.deps.Strategies.Get(ctx, id)
}

// List returns strategies filtered by the supplied query.
func (s *service) List(ctx context.Context, q ListQuery) ([]*domstrategy.Strategy, error) {
	limit := q.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	return s.deps.Strategies.List(ctx, domstrategy.ListQuery{
		AuthorID:   q.AuthorID,
		Status:     q.Status,
		Visibility: q.Visibility,
		Category:   q.Category,
		Tag:        q.Tag,
		Keyword:    q.Keyword,
		Sort:       q.Sort,
		Limit:      limit,
		Offset:     q.Offset,
	})
}

// Delete archives the strategy. The MVP keeps the row around (no hard
// delete) so downstream Ranking / Notification consumers see the
// archive event and stop publishing the strategy.
func (s *service) Delete(ctx context.Context, id int64, callerID int64) error {
	st, err := s.deps.Strategies.Get(ctx, id)
	if err != nil {
		return err
	}
	if err := authorise(st, callerID); err != nil {
		return err
	}
	if err := st.Archive(s.deps.Clock()); err != nil {
		return err
	}
	if err := s.deps.Strategies.Update(ctx, st); err != nil {
		return err
	}
	s.publish(ctx, domevent.EventStrategyArchived, st.ID, domevent.StrategyArchivedPayload{
		StrategyID: st.ID,
	})
	return nil
}

// authorise refuses cross-tenant edits. Passing callerID == 0 disables
// the check, which is the path RPC consumers (event handlers, internal
// jobs) take when there is no human in the loop.
func authorise(st *domstrategy.Strategy, callerID int64) error {
	if callerID == 0 {
		return nil
	}
	if st.AuthorID != 0 && st.AuthorID != callerID {
		return stratErr.ErrNotOwner
	}
	return nil
}

// publish is a forgiving wrapper around the event publisher: missing
// publishers and publish errors degrade to logs because the MVP value
// chain must work even when Kafka is offline.
func (s *service) publish(ctx context.Context, t domevent.EventType, aggregateID int64, payload any) {
	if s.deps.Publisher == nil {
		return
	}
	_ = s.deps.Publisher.Publish(ctx, domevent.Event{
		EventID:       newEventID(),
		EventType:     t,
		EventVersion:  domevent.EventVersionV1,
		OccurredAt:    s.deps.Clock(),
		AggregateType: domevent.AggregateTypeStrategy,
		AggregateID:   fmt.Sprintf("%d", aggregateID),
		Producer:      domevent.ProducerStrategy,
		Payload:       payload,
	})
}
