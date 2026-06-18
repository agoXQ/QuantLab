package backtest

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/google/uuid"

	"github.com/agoXQ/QuantLab/app/backtest/domain/backtestjob"
	bterr "github.com/agoXQ/QuantLab/app/backtest/domain/errors"
	domevent "github.com/agoXQ/QuantLab/app/backtest/domain/event"
	domexec "github.com/agoXQ/QuantLab/app/backtest/domain/executor"
	dommatch "github.com/agoXQ/QuantLab/app/backtest/domain/matching"
	dommarket "github.com/agoXQ/QuantLab/app/backtest/domain/marketdata"
	domorder "github.com/agoXQ/QuantLab/app/backtest/domain/order"
	domportfolio "github.com/agoXQ/QuantLab/app/backtest/domain/portfolio"
	domtrade "github.com/agoXQ/QuantLab/app/backtest/domain/trade"
	"github.com/agoXQ/QuantLab/app/backtest/domain/valueobject"
)

// Run executes the job synchronously. The MVP keeps execution in-process
// rather than going through Kafka so the call site (HTTP handler) gets
// immediate feedback and can return a job + report in one round trip.
// The async / queued path is the next iteration once the worker pool lands.
func (s *service) Run(ctx context.Context, jobID int64) (*RunResult, error) {
	job, err := s.deps.Jobs.Get(ctx, jobID)
	if err != nil {
		return nil, err
	}
	if job.Status.IsTerminal() {
		// A terminal job is replayable only by re-creating it; surfacing
		// the existing artefacts keeps the API idempotent without
		// recomputing.
		return s.collectRunResult(ctx, job)
	}
	now := s.deps.Clock()
	if err := job.MarkRunning(now); err != nil {
		return nil, err
	}
	if err := s.deps.Jobs.Update(ctx, job); err != nil {
		return nil, err
	}
	s.publish(ctx, domevent.EventBacktestStarted, job.ID, domevent.BacktestStartedPayload{JobID: job.ID})

	result, runErr := s.runJob(ctx, job)
	if runErr != nil {
		_ = job.MarkFailed(s.deps.Clock(), runErr.Error())
		_ = s.deps.Jobs.Update(ctx, job)
		s.publish(ctx, domevent.EventBacktestFailed, job.ID, domevent.BacktestFailedPayload{
			JobID:  job.ID,
			Reason: runErr.Error(),
		})
		return nil, runErr
	}
	if err := job.MarkCompleted(s.deps.Clock()); err != nil {
		return nil, err
	}
	if err := s.deps.Jobs.Update(ctx, job); err != nil {
		return nil, err
	}
	s.publish(ctx, domevent.EventBacktestFinished, job.ID, domevent.BacktestFinishedPayload{
		JobID:        job.ID,
		StrategyID:   job.StrategyID,
		TotalReturn:  result.Report.TotalReturn,
		AnnualReturn: result.Report.AnnualReturn,
		Sharpe:       result.Report.SharpeRatio,
		MaxDrawdown:  result.Report.MaxDrawdown,
	})
	result.Job = job
	return result, nil
}

// collectRunResult reassembles a RunResult for a job that already finished.
func (s *service) collectRunResult(ctx context.Context, job *backtestjob.BacktestJob) (*RunResult, error) {
	rep, err := s.deps.Reports.Get(ctx, job.ID)
	if err != nil && err != bterr.ErrReportNotFound {
		return nil, err
	}
	trades, err := s.deps.Trades.ListByJob(ctx, job.ID)
	if err != nil {
		return nil, err
	}
	orders, err := s.deps.Orders.ListByJob(ctx, job.ID)
	if err != nil {
		return nil, err
	}
	snapshots, err := s.deps.Portfolios.ListSnapshots(ctx, job.ID)
	if err != nil {
		return nil, err
	}
	return &RunResult{Job: job, Report: rep, Trades: trades, Orders: orders, Snapshots: snapshots}, nil
}

// runJob is the actual day-by-day replay loop.
//
// The structure mirrors the TD's "Market Replay Engine" section: load
// market data once, iterate the trading calendar, evaluate the formula at
// the close of every rebalance day, queue orders, fill them at the next
// open, and snapshot the portfolio at every close. The work is split into
// small helper methods so each step is independently testable and easy to
// swap (e.g. a different rebalance trigger).
func (s *service) runJob(ctx context.Context, job *backtestjob.BacktestJob) (*RunResult, error) {
	// Trading calendar: drives the replay. We tolerate a missing market
	// service by deriving the calendar from the loaded bars below, which
	// keeps the in-memory test fixture viable.
	calendar, err := s.loadCalendar(ctx, job)
	if err != nil {
		return nil, err
	}
	bars, err := s.deps.MarketData.LoadBars(ctx, dommarket.BarsRequest{
		StockCodes:  job.Universe,
		Range:       job.Range,
		DataVersion: job.DataVersion,
	})
	if err != nil {
		return nil, err
	}
	if len(bars) == 0 {
		return nil, bterr.ErrMarketDataMissing
	}
	if len(calendar) == 0 {
		calendar = calendarFromBars(bars, job.Range)
	}
	if len(calendar) == 0 {
		return nil, bterr.ErrMarketDataMissing
	}
	barIndex := indexBars(bars)

	port := domportfolio.New(job.ID, job.InitialCapital)
	pendingOrders := make([]*domorder.Order, 0)
	allOrders := make([]*domorder.Order, 0)
	allTrades := make([]*domtrade.Trade, 0)
	snapshots := make([]domportfolio.Snapshot, 0, len(calendar))

	rebalanceTriggered := newRebalanceTrigger(job.Config.RebalanceFrequency)

	for i, day := range calendar {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		// Step 1: match orders queued the previous day at this day's open.
		if len(pendingOrders) > 0 {
			barSnapshot := buildBarSnapshot(barIndex, day)
			fills, rejects := s.deps.Matching.Match(pendingOrders, barSnapshot)
			for _, t := range fills {
				port.ApplyTrade(t)
			}
			allTrades = append(allTrades, fills...)
			allOrders = append(allOrders, rejects...)
			pendingOrders = pendingOrders[:0]
		}
		// Step 2: mark to market using close prices.
		closePrices := closesAt(barIndex, day)
		port.MarkToMarket(closePrices)
		snapshots = append(snapshots, port.Snapshot(day))

		// Step 3: evaluate strategy on a rebalance day. We trigger on the
		// first replay day too so the engine puts capital to work
		// immediately instead of sitting in cash for one rebalance window.
		if i == 0 || rebalanceTriggered(day) {
			signals, err := s.executeStrategy(ctx, job, day)
			if err != nil {
				return nil, err
			}
			orders := s.generateOrders(job, port, signals, closePrices, day)
			if len(orders) > 0 {
				if err := s.deps.Orders.BulkInsert(ctx, orders); err != nil {
					return nil, err
				}
				pendingOrders = append(pendingOrders, orders...)
				allOrders = append(allOrders, orders...)
			}
		}
	}

	// Persist the trades / snapshots in bulk; report follows.
	if err := s.deps.Trades.BulkInsert(ctx, allTrades); err != nil {
		return nil, err
	}
	if err := s.deps.Portfolios.BulkInsertSnapshots(ctx, snapshots); err != nil {
		return nil, err
	}

	rep := buildReport(metricsConfig{
		jobID:          job.ID,
		initialCapital: job.InitialCapital,
		startDate:      job.Range.Start,
		endDate:        job.Range.End,
		now:            s.deps.Clock(),
	}, snapshots, allTrades)
	if err := s.deps.Reports.Save(ctx, rep); err != nil {
		return nil, err
	}

	return &RunResult{
		Job:       job,
		Report:    rep,
		Trades:    allTrades,
		Orders:    allOrders,
		Snapshots: snapshots,
	}, nil
}

// executeStrategy delegates to the configured StrategyExecutor. The
// LookbackBars hint is forwarded so adapters that talk to Formula's
// EvaluatorService can ask the data port for the right amount of history.
func (s *service) executeStrategy(ctx context.Context, job *backtestjob.BacktestJob, day time.Time) ([]domexec.Signal, error) {
	if s.deps.Executor == nil {
		return nil, bterr.ErrEvaluatorUnavailable
	}
	res, err := s.deps.Executor.Execute(ctx, domexec.Request{
		Formula:      job.Formula,
		Universe:     job.Universe,
		AsOfDate:     day,
		LookbackBars: job.Config.LookbackBars,
		DataVersion:  job.DataVersion,
	})
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, nil
	}
	return res.Signals, nil
}

// generateOrders turns Signals into Orders using equal-weighted target
// allocation: pick up to MaxPositionCount BUY signals (sorted by score
// desc), close any held position not in the target set, and rebalance
// existing names by buying / selling deltas to the target lot.
//
// Lot size is fixed at 100 shares (A-share convention). Cash that does
// not fill a full lot is left idle; the next rebalance picks it up.
func (s *service) generateOrders(
	job *backtestjob.BacktestJob,
	port *domportfolio.Portfolio,
	signals []domexec.Signal,
	closePrices map[string]float64,
	day time.Time,
) []*domorder.Order {
	const lotSize = 100

	buys := make([]domexec.Signal, 0, len(signals))
	for _, sig := range signals {
		if sig.Action != domexec.SignalBuy {
			continue
		}
		buys = append(buys, sig)
	}
	// Stable, score-desc ordering keeps determinism when scores tie.
	sort.SliceStable(buys, func(i, j int) bool {
		if buys[i].Score == buys[j].Score {
			return buys[i].StockCode < buys[j].StockCode
		}
		return buys[i].Score > buys[j].Score
	})
	if cap := job.Config.MaxPositionCount; cap > 0 && len(buys) > cap {
		buys = buys[:cap]
	}

	target := make(map[string]struct{}, len(buys))
	for _, sig := range buys {
		target[sig.StockCode] = struct{}{}
	}

	orders := make([]*domorder.Order, 0)
	now := s.deps.Clock()

	// Sells first: anything held but not in the target set goes flat.
	holdings := make([]string, 0, len(port.Positions))
	for code := range port.Positions {
		holdings = append(holdings, code)
	}
	sort.Strings(holdings)
	for _, code := range holdings {
		if _, keep := target[code]; keep {
			continue
		}
		pos := port.Position(code)
		if pos == nil || pos.Quantity <= 0 {
			continue
		}
		orders = append(orders, &domorder.Order{
			JobID:       job.ID,
			StockCode:   code,
			Side:        valueobject.OrderSideSell,
			Quantity:    pos.Quantity,
			Status:      valueobject.OrderStatusPending,
			SubmittedAt: now,
		})
	}

	// Buys: split available capital evenly across target stocks.
	if len(buys) > 0 {
		totalAsset := port.Cash
		for _, pos := range port.Positions {
			totalAsset += pos.MarketValue
		}
		perStock := totalAsset / float64(len(buys))
		for _, sig := range buys {
			price, ok := closePrices[sig.StockCode]
			if !ok || price <= 0 || math.IsNaN(price) {
				continue
			}
			targetQty := int64(perStock/price/float64(lotSize)) * lotSize
			if targetQty <= 0 {
				continue
			}
			pos := port.Position(sig.StockCode)
			currentQty := int64(0)
			if pos != nil {
				currentQty = pos.Quantity
			}
			delta := targetQty - currentQty
			if delta > 0 {
				orders = append(orders, &domorder.Order{
					JobID:       job.ID,
					StockCode:   sig.StockCode,
					Side:        valueobject.OrderSideBuy,
					Quantity:    delta,
					Status:      valueobject.OrderStatusPending,
					SubmittedAt: now,
				})
			} else if delta < 0 {
				orders = append(orders, &domorder.Order{
					JobID:       job.ID,
					StockCode:   sig.StockCode,
					Side:        valueobject.OrderSideSell,
					Quantity:    -delta,
					Status:      valueobject.OrderStatusPending,
					SubmittedAt: now,
				})
			}
		}
	}

	_ = day // reserved for future per-day audit fields
	return orders
}

// loadCalendar tries the configured market data provider first; if the
// adapter is wired against an in-memory port that does not implement
// TradingDays the caller falls back to a calendar derived from the loaded
// bars (every day with at least one bar across the universe is considered
// a trading day).
func (s *service) loadCalendar(ctx context.Context, job *backtestjob.BacktestJob) ([]time.Time, error) {
	if s.deps.MarketData == nil {
		return nil, bterr.ErrMarketDataMissing
	}
	days, err := s.deps.MarketData.TradingDays(ctx, dommarket.CalendarRequest{
		Range:  job.Range,
		Market: "CN",
	})
	if err != nil {
		return nil, err
	}
	return days, nil
}

// indexBars produces a per-day, per-stock lookup keyed by the trade-date
// midnight (UTC) so day equality checks line up regardless of the source
// timezone.
func indexBars(in map[string][]dommarket.Bar) map[string]map[time.Time]dommarket.Bar {
	out := make(map[string]map[time.Time]dommarket.Bar, len(in))
	for code, bars := range in {
		m := make(map[time.Time]dommarket.Bar, len(bars))
		for _, b := range bars {
			m[normalizeDate(b.TradeDate)] = b
		}
		out[code] = m
	}
	return out
}

// calendarFromBars derives a calendar from the union of bar dates within
// the job range. Used when the market data adapter does not expose a
// TradingDays implementation (in-memory fixture).
func calendarFromBars(in map[string][]dommarket.Bar, rng valueobject.DateRange) []time.Time {
	seen := make(map[time.Time]struct{})
	for _, bars := range in {
		for _, b := range bars {
			d := normalizeDate(b.TradeDate)
			if !rng.Start.IsZero() && d.Before(normalizeDate(rng.Start)) {
				continue
			}
			if !rng.End.IsZero() && d.After(normalizeDate(rng.End)) {
				continue
			}
			seen[d] = struct{}{}
		}
	}
	out := make([]time.Time, 0, len(seen))
	for d := range seen {
		out = append(out, d)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Before(out[j]) })
	return out
}

// buildBarSnapshot returns a per-stock snapshot map keyed by stock code at
// the given trade date. The returned map is the input the matching engine
// expects.
func buildBarSnapshot(idx map[string]map[time.Time]dommarket.Bar, day time.Time) map[string]dommatch.BarSnapshot {
	out := make(map[string]dommatch.BarSnapshot, len(idx))
	d := normalizeDate(day)
	for code, days := range idx {
		bar, ok := days[d]
		if !ok {
			continue
		}
		out[code] = dommatch.BarSnapshot{
			Open:   bar.Open,
			High:   bar.High,
			Low:    bar.Low,
			Close:  bar.Close,
			Volume: bar.Volume,
		}
	}
	return out
}

// closesAt returns a stock-code -> close-price map for the given day.
// Missing bars surface as NaN, which Portfolio.MarkToMarket interprets as
// "no update".
func closesAt(idx map[string]map[time.Time]dommarket.Bar, day time.Time) map[string]float64 {
	out := make(map[string]float64, len(idx))
	d := normalizeDate(day)
	for code, days := range idx {
		bar, ok := days[d]
		if !ok {
			out[code] = math.NaN()
			continue
		}
		out[code] = bar.Close
	}
	return out
}

// normalizeDate truncates t to UTC midnight so map lookups are stable.
func normalizeDate(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

// newRebalanceTrigger returns a closure that decides whether a given trade
// date is a rebalance day. The logic is intentionally calendar-driven for
// the MVP; signal-driven rebalancing belongs in the next iteration.
func newRebalanceTrigger(freq valueobject.RebalanceFrequency) func(time.Time) bool {
	switch freq {
	case valueobject.RebalanceWeekly:
		var lastWeek int
		return func(d time.Time) bool {
			y, w := d.ISOWeek()
			key := y*100 + w
			if key == lastWeek {
				return false
			}
			lastWeek = key
			return true
		}
	case valueobject.RebalanceMonthly:
		var lastMonth time.Month
		var lastYear int
		return func(d time.Time) bool {
			if d.Year() == lastYear && d.Month() == lastMonth {
				return false
			}
			lastYear = d.Year()
			lastMonth = d.Month()
			return true
		}
	default:
		return func(time.Time) bool { return true }
	}
}

// silence "imported and not used" across helper packages that only appear
// in branches when the engine is exercised.
var _ = uuid.NewString
var _ = fmt.Sprintf
