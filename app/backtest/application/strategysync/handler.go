// Package strategysync is the application-layer use case that turns a
// decoded Strategy event into a baseline Backtest run. It composes the
// platform-wide pieces: a Resolver to materialise the strategy body,
// the Backtest application Service to create + submit the job, and a
// market-data Provider to learn the available trading window so the
// baseline replays a sensible range without forcing the strategy
// service to embed scheduling metadata in the event payload.
package strategysync

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	appBacktest "github.com/agoXQ/QuantLab/app/backtest/application/backtest"
	"github.com/agoXQ/QuantLab/app/backtest/domain/backtestjob"
	dommarket "github.com/agoXQ/QuantLab/app/backtest/domain/marketdata"
	domsync "github.com/agoXQ/QuantLab/app/backtest/domain/strategysync"
	"github.com/agoXQ/QuantLab/app/backtest/domain/valueobject"
)

// Config controls the baseline behaviour. Defaults match a one-year
// daily replay over the CSI300 baseline universe; production
// deployments override Universe / Range from configuration so the
// baseline tracks the platform's recommended A-share basket.
type Config struct {
	// Universe is the fallback set of stock codes to run the baseline
	// against. The MVP keeps this small so a single publish event does
	// not flood the worker pool. Empty falls back to BaselineUniverse.
	Universe []string
	// Lookback caps how far back the baseline replay starts when the
	// provider does not surface a calendar. Defaults to 365 days.
	Lookback time.Duration
	// InitialCapital seeds the simulated portfolio. Defaults to
	// 1_000_000 RMB which matches the platform's reference settings.
	InitialCapital float64
	// Benchmark is the index code recorded on the job. Defaults to
	// `000300` (CSI300) so the resulting report is comparable across
	// strategies.
	Benchmark string
	// AutoSubmit drives whether the handler hands the job to the queue
	// after creating it. Production keeps this true; tests flip it off
	// so they can assert on the persisted job without spinning up a
	// worker pool.
	AutoSubmit bool
	// Tag is a short label appended to the job name so the operator can
	// tell baseline runs apart from human-submitted ones. Defaults to
	// `auto-baseline`.
	Tag string
}

// BaselineUniverse is the fallback universe used when Config.Universe
// is empty. We pick a single stable A-share so the baseline executes
// even on a fresh deployment with minimal market data, and the operator
// can broaden it through configuration.
var BaselineUniverse = []string{"000001"}

// BaselineHandler implements domsync.Handler by submitting a baseline
// backtest whenever the Strategy service announces a publish or fresh
// version. It is safe to use concurrently across goroutines.
type BaselineHandler struct {
	cfg        Config
	resolver   domsync.Resolver
	backtests  appBacktest.Service
	marketData dommarket.Provider
	clock      func() time.Time

	// dedupe prevents the same (strategy_id, version_id) tuple from
	// triggering two runs when the platform replays an event window.
	// A small map is enough for the MVP; a real cache (Redis) drops in
	// behind the same interface later.
	dedupeMu sync.Mutex
	dedupe   map[string]struct{}
}

// NewBaselineHandler wires the use case. Resolver and BacktestService
// are required; MarketData is optional (when nil the handler picks a
// fixed lookback window so the call still succeeds).
func NewBaselineHandler(
	cfg Config,
	resolver domsync.Resolver,
	backtests appBacktest.Service,
	marketData dommarket.Provider,
	clock func() time.Time,
) *BaselineHandler {
	if clock == nil {
		clock = time.Now
	}
	if cfg.Lookback <= 0 {
		cfg.Lookback = 365 * 24 * time.Hour
	}
	if cfg.InitialCapital <= 0 {
		cfg.InitialCapital = 1_000_000
	}
	if strings.TrimSpace(cfg.Benchmark) == "" {
		cfg.Benchmark = "000300"
	}
	if strings.TrimSpace(cfg.Tag) == "" {
		cfg.Tag = "auto-baseline"
	}
	return &BaselineHandler{
		cfg:        cfg,
		resolver:   resolver,
		backtests:  backtests,
		marketData: marketData,
		clock:      clock,
		dedupe:     map[string]struct{}{},
	}
}

// OnPublished is called when a strategy is published. We treat publish
// as the canonical moment the baseline should run; OnVersionCreated is
// a softer trigger that only fires when the strategy is already public,
// because every version of a draft would otherwise spam the queue.
func (h *BaselineHandler) OnPublished(ctx context.Context, env domsync.Envelope, p domsync.PublishedPayload) error {
	if p.StrategyID == 0 || p.VersionID == 0 {
		return fmt.Errorf("strategysync: published payload missing ids: %+v", p)
	}
	return h.runBaseline(ctx, p.StrategyID, p.VersionID, p.AuthorID, "published")
}

// OnVersionCreated is intentionally a no-op in the MVP: we only auto-run
// on publish so an active editor does not flood the worker pool with one
// run per save. Returning nil keeps the consumer happy while we leave
// the hook live for a future opt-in policy.
func (h *BaselineHandler) OnVersionCreated(ctx context.Context, env domsync.Envelope, p domsync.VersionCreatedPayload) error {
	_ = ctx
	_ = env
	_ = p
	return nil
}

// runBaseline performs the actual create + submit pipeline. Extracted so
// future triggers (OnVersionCreated under a "draft auto-run" policy,
// manual replays) can share the same code.
func (h *BaselineHandler) runBaseline(ctx context.Context, strategyID, versionID, authorID int64, reason string) error {
	if h.markSeen(strategyID, versionID) {
		log.Printf("[strategysync] skip duplicate baseline strategy=%d version=%d reason=%s", strategyID, versionID, reason)
		return nil
	}

	snap, err := h.resolver.Resolve(ctx, strategyID, versionID)
	if err != nil {
		return fmt.Errorf("strategysync: resolve strategy=%d version=%d: %w", strategyID, versionID, err)
	}
	if snap == nil || strings.TrimSpace(snap.FormulaText) == "" {
		return fmt.Errorf("strategysync: empty formula for strategy=%d version=%d", strategyID, versionID)
	}

	rng, err := h.pickRange(ctx)
	if err != nil {
		return fmt.Errorf("strategysync: pick range: %w", err)
	}

	universe := append([]string(nil), h.cfg.Universe...)
	if len(universe) == 0 {
		universe = append([]string(nil), BaselineUniverse...)
	}

	name := fmt.Sprintf("%s/%s/v%d", h.cfg.Tag, snapshotTitle(snap), versionID)

	created, err := h.backtests.Create(ctx, appBacktest.CreateBacktestRequest{
		UserID:         authorID,
		StrategyID:     strategyID,
		VersionID:      versionID,
		Name:           name,
		Formula:        snap.FormulaText,
		Universe:       universe,
		Benchmark:      h.cfg.Benchmark,
		InitialCapital: h.cfg.InitialCapital,
		Range:          rng,
		Config:         backtestjob.DefaultConfig(),
	})
	if err != nil {
		return fmt.Errorf("strategysync: create backtest: %w", err)
	}
	log.Printf("[strategysync] baseline created job=%d strategy=%d version=%d reason=%s",
		created.Job.ID, strategyID, versionID, reason)

	if !h.cfg.AutoSubmit {
		return nil
	}
	if _, err := h.backtests.Submit(ctx, created.Job.ID); err != nil {
		// Submit failures are logged but not propagated: the job row
		// still exists in CREATED so an operator can retry; emitting
		// an error here would NACK the Kafka message and stall the
		// rest of the pipeline.
		log.Printf("[strategysync] warning: submit baseline job=%d: %v", created.Job.ID, err)
		return nil
	}
	log.Printf("[strategysync] baseline submitted job=%d strategy=%d version=%d", created.Job.ID, strategyID, versionID)
	return nil
}

// pickRange asks the market provider for the trading window we should
// replay. We use the most recent Lookback of trading days; when the
// provider is unavailable we fall back to a calendar-day window so the
// call still completes.
func (h *BaselineHandler) pickRange(ctx context.Context) (valueobject.DateRange, error) {
	now := h.clock().UTC().Truncate(24 * time.Hour)
	end := now
	start := end.Add(-h.cfg.Lookback)
	rng := valueobject.DateRange{Start: start, End: end}
	if h.marketData == nil {
		return rng, nil
	}
	days, err := h.marketData.TradingDays(ctx, dommarket.CalendarRequest{Range: rng})
	if err != nil {
		// Calendar errors are noisy but recoverable; the baseline still
		// has a sensible default window.
		log.Printf("[strategysync] warning: trading days unavailable: %v; using fixed window", err)
		return rng, nil
	}
	if len(days) == 0 {
		return rng, nil
	}
	return valueobject.DateRange{Start: days[0], End: days[len(days)-1]}, nil
}

// markSeen records the (strategy, version) pair and reports whether we
// have already seen it in this process. The map grows unbounded in
// theory; in practice the platform restarts each service often enough
// that the memory footprint stays small. A real production deploy
// pushes this state into Redis behind the same interface.
func (h *BaselineHandler) markSeen(strategyID, versionID int64) bool {
	key := fmt.Sprintf("%d/%d", strategyID, versionID)
	h.dedupeMu.Lock()
	defer h.dedupeMu.Unlock()
	if _, ok := h.dedupe[key]; ok {
		return true
	}
	h.dedupe[key] = struct{}{}
	return false
}

// snapshotTitle picks a readable label for the job name, falling back
// to a deterministic placeholder when the resolver did not surface one.
func snapshotTitle(s *domsync.StrategySnapshot) string {
	t := strings.TrimSpace(s.Title)
	if t == "" {
		return fmt.Sprintf("strategy-%d", s.StrategyID)
	}
	return t
}
