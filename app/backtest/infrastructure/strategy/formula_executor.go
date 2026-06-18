// Package strategy contains StrategyExecutor adapters bridging Backtest to
// the Formula Engine. The default adapter wraps the in-process
// formula.EvaluatorService; a future adapter will wrap a gRPC client when
// Formula and Backtest live in separate processes.
package strategy

import (
	"context"
	"fmt"

	domexec "github.com/agoXQ/QuantLab/app/backtest/domain/executor"
	appFormula "github.com/agoXQ/QuantLab/app/formula/application/formula"
	domainCompiler "github.com/agoXQ/QuantLab/app/formula/domain/compiler"
	domainEval "github.com/agoXQ/QuantLab/app/formula/domain/evaluator"
)

// FormulaExecutor wraps formula.EvaluatorService and a DataPort. The
// DataPort is request-injected by the Formula Service via context, but
// Backtest does not want to leak that detail; the adapter takes the port
// once at construction and passes it through on every call.
type FormulaExecutor struct {
	svc      appFormula.EvaluatorService
	dataPort domainEval.DataPort
}

// NewFormulaExecutor wires the adapter. dataPort must not be nil.
func NewFormulaExecutor(svc appFormula.EvaluatorService, dataPort domainEval.DataPort) *FormulaExecutor {
	return &FormulaExecutor{svc: svc, dataPort: dataPort}
}

// Execute implements executor.StrategyExecutor.
func (f *FormulaExecutor) Execute(ctx context.Context, req domexec.Request) (*domexec.Result, error) {
	if f.svc == nil || f.dataPort == nil {
		return nil, fmt.Errorf("strategy: formula executor not configured")
	}
	res, err := f.svc.Evaluate(ctx, appFormula.EvaluateRequest{
		Formula:      req.Formula,
		Universe:     req.Universe,
		AsOfDate:     req.AsOfDate,
		LookbackBars: req.LookbackBars,
		DataVersion:  req.DataVersion,
		DataPort:     f.dataPort,
	})
	if err != nil {
		return nil, err
	}
	return mapResult(res.Result), nil
}

// mapResult flattens a formula.Result into a stream of executor.Signals.
//
// The mapping rules:
//   - FILTER / SIGNAL: every selected stock becomes a BUY at score 1; the
//     order generator decides weights.
//   - SORT: top-N stocks (decided by the order generator) become BUY with
//     the score attached for ranking-based weighting.
//   - VALUE: stocks with values strictly above zero become BUY; values <=0
//     or NaN become HOLD. This is the simplest mapping that lets a VALUE
//     formula (e.g. CROSS(MA(C,5), MA(C,20))) drive entries; richer
//     mappings can be reintroduced when strategy-service composition lands.
func mapResult(r *domainEval.Result) *domexec.Result {
	out := &domexec.Result{}
	if r == nil {
		return out
	}
	switch r.PlanType {
	case domainCompiler.PlanTypeFilter, domainCompiler.PlanTypeSignal:
		if r.Selection == nil {
			return out
		}
		signals := make([]domexec.Signal, len(r.Selection.StockCodes))
		for i, code := range r.Selection.StockCodes {
			signals[i] = domexec.Signal{StockCode: code, Action: domexec.SignalBuy, Score: 1}
		}
		out.Signals = signals
	case domainCompiler.PlanTypeSort:
		if r.Ranking == nil {
			return out
		}
		signals := make([]domexec.Signal, len(r.Ranking.StockCodes))
		for i, code := range r.Ranking.StockCodes {
			signals[i] = domexec.Signal{StockCode: code, Action: domexec.SignalBuy, Score: r.Ranking.Scores[i]}
		}
		out.Signals = signals
	case domainCompiler.PlanTypeValue:
		if r.Values == nil {
			return out
		}
		signals := make([]domexec.Signal, 0, len(r.Values.StockCodes))
		for i, code := range r.Values.StockCodes {
			v := r.Values.Values[i]
			action := domexec.SignalHold
			if v > 0 {
				action = domexec.SignalBuy
			}
			signals = append(signals, domexec.Signal{StockCode: code, Action: action, Score: v})
		}
		out.Signals = signals
	}
	return out
}
