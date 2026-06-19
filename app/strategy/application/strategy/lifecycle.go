package strategy

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"

	stratErr "github.com/agoXQ/QuantLab/app/strategy/domain/errors"
	domevent "github.com/agoXQ/QuantLab/app/strategy/domain/event"
	domfork "github.com/agoXQ/QuantLab/app/strategy/domain/fork"
	domstrategy "github.com/agoXQ/QuantLab/app/strategy/domain/strategy"
	"github.com/agoXQ/QuantLab/app/strategy/domain/valueobject"
	domversion "github.com/agoXQ/QuantLab/app/strategy/domain/version"
)

// CreateVersion appends a new version snapshot, advances the strategy
// pointer to it, and emits StrategyVersionCreated.
func (s *service) CreateVersion(ctx context.Context, req CreateVersionRequest) (*CreateVersionResult, error) {
	st, err := s.deps.Strategies.Get(ctx, req.StrategyID)
	if err != nil {
		return nil, err
	}
	if err := authorise(st, req.CallerID); err != nil {
		return nil, err
	}
	if st.Status == valueobject.LifecycleStatusArchived {
		return nil, stratErr.ErrStrategyArchived
	}
	latest, err := s.deps.Versions.LatestNumber(ctx, req.StrategyID)
	if err != nil {
		return nil, err
	}
	now := s.deps.Clock()
	ver := &domversion.StrategyVersion{
		StrategyID:    req.StrategyID,
		VersionNo:     fmt.Sprintf("v%d", latest+1),
		FormulaText:   strings.TrimSpace(req.FormulaText),
		BuyRule:       strings.TrimSpace(req.BuyRule),
		SellRule:      strings.TrimSpace(req.SellRule),
		RiskRule:      strings.TrimSpace(req.RiskRule),
		PositionRule:  strings.TrimSpace(req.PositionRule),
		RebalanceRule: strings.TrimSpace(req.RebalanceRule),
		ChangeLog:     strings.TrimSpace(req.ChangeLog),
		CreatedBy:     req.CallerID,
		CreatedAt:     now,
	}
	if err := ver.Validate(); err != nil {
		return nil, err
	}
	if err := s.deps.Versions.Create(ctx, ver); err != nil {
		return nil, err
	}
	if err := st.AttachVersion(ver.ID, now); err != nil {
		return nil, err
	}
	if err := s.deps.Strategies.Update(ctx, st); err != nil {
		return nil, err
	}
	s.publish(ctx, domevent.EventStrategyVersionCreated, st.ID, domevent.StrategyVersionCreatedPayload{
		StrategyID: st.ID,
		VersionID:  ver.ID,
		VersionNo:  ver.VersionNo,
	})
	return &CreateVersionResult{Strategy: st, Version: ver}, nil
}

// GetVersion returns a single version. Used both by direct API reads
// and by Backtest when it materialises the body to run.
func (s *service) GetVersion(ctx context.Context, id int64) (*domversion.StrategyVersion, error) {
	if id <= 0 {
		return nil, stratErr.ErrInvalidVersion
	}
	return s.deps.Versions.Get(ctx, id)
}

// ListVersions returns the version history for one strategy. The
// repository sorts newest-first; the MVP UI surfaces the same order.
func (s *service) ListVersions(ctx context.Context, strategyID int64, limit int) ([]*domversion.StrategyVersion, error) {
	if strategyID <= 0 {
		return nil, stratErr.ErrInvalidStrategy
	}
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	return s.deps.Versions.ListByStrategy(ctx, strategyID, limit)
}

// Publish flips the strategy public against either the supplied version
// or the strategy's current version when the caller leaves it zero.
func (s *service) Publish(ctx context.Context, req PublishRequest) (*domstrategy.Strategy, error) {
	st, err := s.deps.Strategies.Get(ctx, req.StrategyID)
	if err != nil {
		return nil, err
	}
	if err := authorise(st, req.CallerID); err != nil {
		return nil, err
	}
	versionID := req.VersionID
	if versionID == 0 {
		versionID = st.CurrentVersionID
	}
	if versionID == 0 {
		return nil, stratErr.ErrPublishWithoutVersion
	}
	ver, err := s.deps.Versions.Get(ctx, versionID)
	if err != nil {
		return nil, err
	}
	if ver.StrategyID != st.ID {
		return nil, stratErr.ErrInvalidVersion
	}
	if err := st.Publish(versionID, s.deps.Clock()); err != nil {
		return nil, err
	}
	if err := s.deps.Strategies.Update(ctx, st); err != nil {
		return nil, err
	}
	s.publish(ctx, domevent.EventStrategyPublished, st.ID, domevent.StrategyPublishedPayload{
		StrategyID: st.ID,
		VersionID:  versionID,
		AuthorID:   st.AuthorID,
	})
	return st, nil
}

// Archive flips the strategy off the public surface for good. Same
// semantics as Delete; we keep both because the API surface differs.
func (s *service) Archive(ctx context.Context, req ArchiveRequest) (*domstrategy.Strategy, error) {
	st, err := s.deps.Strategies.Get(ctx, req.StrategyID)
	if err != nil {
		return nil, err
	}
	if err := authorise(st, req.CallerID); err != nil {
		return nil, err
	}
	if err := st.Archive(s.deps.Clock()); err != nil {
		return nil, err
	}
	if err := s.deps.Strategies.Update(ctx, st); err != nil {
		return nil, err
	}
	s.publish(ctx, domevent.EventStrategyArchived, st.ID, domevent.StrategyArchivedPayload{
		StrategyID: st.ID,
	})
	return st, nil
}

// Fork copies a public strategy into the caller's namespace. We snapshot
// the source's current version into the new strategy so the fork is
// runnable immediately. Source ownership is not checked because public
// strategies are by definition forkable; private forks are rejected.
func (s *service) Fork(ctx context.Context, req ForkRequest) (*ForkResult, error) {
	if req.CallerID == 0 {
		return nil, stratErr.ErrNotOwner
	}
	src, err := s.deps.Strategies.Get(ctx, req.SourceStrategyID)
	if err != nil {
		return nil, err
	}
	if src.Visibility != valueobject.VisibilityPublic {
		return nil, stratErr.ErrNotOwner
	}
	if src.CurrentVersionID == 0 {
		return nil, stratErr.ErrPublishWithoutVersion
	}
	srcVer, err := s.deps.Versions.Get(ctx, src.CurrentVersionID)
	if err != nil {
		return nil, err
	}
	now := s.deps.Clock()
	title := strings.TrimSpace(req.Title)
	if title == "" {
		title = src.Title
	}
	target := &domstrategy.Strategy{
		AuthorID:         req.CallerID,
		Title:            title,
		Description:      src.Description,
		Category:         src.Category,
		Tags:             append([]string(nil), src.Tags...),
		Status:           valueobject.LifecycleStatusDraft,
		Visibility:       valueobject.VisibilityPrivate,
		SourceStrategyID: src.ID,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	if err := target.Validate(); err != nil {
		return nil, err
	}
	if err := s.deps.Strategies.Create(ctx, target); err != nil {
		return nil, err
	}
	clone := &domversion.StrategyVersion{
		StrategyID:    target.ID,
		VersionNo:     "v1",
		FormulaText:   srcVer.FormulaText,
		BuyRule:       srcVer.BuyRule,
		SellRule:      srcVer.SellRule,
		RiskRule:      srcVer.RiskRule,
		PositionRule:  srcVer.PositionRule,
		RebalanceRule: srcVer.RebalanceRule,
		ChangeLog:     fmt.Sprintf("forked from strategy %d version %s", src.ID, srcVer.VersionNo),
		CreatedBy:     req.CallerID,
		CreatedAt:     now,
	}
	if err := s.deps.Versions.Create(ctx, clone); err != nil {
		return nil, err
	}
	if err := target.AttachVersion(clone.ID, now); err != nil {
		return nil, err
	}
	if err := s.deps.Strategies.Update(ctx, target); err != nil {
		return nil, err
	}
	if err := s.deps.Forks.Create(ctx, &domfork.StrategyFork{
		SourceStrategyID: src.ID,
		TargetStrategyID: target.ID,
		CreatorID:        req.CallerID,
		CreatedAt:        now,
	}); err != nil {
		return nil, err
	}
	if err := s.deps.Strategies.IncrementForkCount(ctx, src.ID); err != nil {
		return nil, err
	}
	s.publish(ctx, domevent.EventStrategyForked, target.ID, domevent.StrategyForkedPayload{
		SourceStrategyID: src.ID,
		TargetStrategyID: target.ID,
		CreatorID:        req.CallerID,
	})
	return &ForkResult{Strategy: target, Version: clone}, nil
}

// MarkBacktested flips the strategy into Backtested when a Backtest
// finished event lands. The use case is exposed on the public service
// so a future Kafka consumer can drive it without depending on the
// service struct directly.
func (s *service) MarkBacktested(ctx context.Context, req MarkBacktestedRequest) (*domstrategy.Strategy, error) {
	st, err := s.deps.Strategies.Get(ctx, req.StrategyID)
	if err != nil {
		return nil, err
	}
	if err := st.MarkBacktested(s.deps.Clock()); err != nil {
		return nil, err
	}
	if err := s.deps.Strategies.Update(ctx, st); err != nil {
		return nil, err
	}
	return st, nil
}

// newEventID exists in a single place so the publishing site can swap
// in a deterministic generator under tests if needed.
func newEventID() string { return uuid.NewString() }
