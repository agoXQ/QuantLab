// Package evaluator implements the AST evaluator for the Formula Engine.
//
// It traverses an ExecutionPlan once per stock in the requested universe,
// resolves identifiers against the per-stock data context, dispatches
// function calls through the indicator library, and produces a typed
// per-stock value. The application layer reduces those values into the
// shape the plan promises (FILTER -> Selection, SORT -> Ranking,
// VALUE -> ValueMap).
package evaluator

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"

	domainAST "github.com/agoXQ/QuantLab/app/formula/domain/ast"
	domainCompiler "github.com/agoXQ/QuantLab/app/formula/domain/compiler"
	domainErrors "github.com/agoXQ/QuantLab/app/formula/domain/errors"
	domainEval "github.com/agoXQ/QuantLab/app/formula/domain/evaluator"
	domainInd "github.com/agoXQ/QuantLab/app/formula/domain/indicators"
	domainVar "github.com/agoXQ/QuantLab/app/formula/domain/variable"
	"github.com/agoXQ/QuantLab/app/formula/domain/series"
)

// evaluator implements the domain Evaluator interface.
type evaluator struct {
	indicators domainInd.Library
	variables  domainVar.Registry
}

// New builds an evaluator wired to the given indicator library and variable
// registry. Both dependencies are immutable lookup tables; the evaluator
// itself holds no per-call state and is therefore safe for concurrent use.
func New(library domainInd.Library, variables domainVar.Registry) domainEval.Evaluator {
	return &evaluator{
		indicators: library,
		variables:  variables,
	}
}

func (e *evaluator) Evaluate(
	ctx context.Context,
	plan *domainCompiler.ExecutionPlan,
	req domainEval.Request,
) (*domainEval.Result, error) {
	if plan == nil || plan.Root == nil {
		return nil, domainErrors.NewError(domainErrors.ErrSyntaxError, "evaluator: empty plan")
	}
	if len(req.Universe) == 0 {
		return emptyResult(plan.PlanType), nil
	}
	if req.LookbackBars <= 0 {
		req.LookbackBars = defaultLookback
	}

	deps, ok := dataPortFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("evaluator: no DataPort in context (use evaluator.WithDataPort)")
	}

	bars, err := deps.LoadBars(ctx, domainEval.BarsRequest{
		StockCodes:   req.Universe,
		AsOfDate:     req.AsOfDate,
		LookbackBars: req.LookbackBars,
		DataVersion:  req.DataVersion,
	})
	if err != nil {
		return nil, fmt.Errorf("evaluator: load bars: %w", err)
	}

	metrics := collectFinancialIdentifiers(plan.Root, e.variables)
	var fins map[string]map[string]float64
	if len(metrics) > 0 {
		fins, err = deps.LoadFinancialsLatest(ctx, domainEval.FinancialsRequest{
			StockCodes:  req.Universe,
			Metrics:     metrics,
			AsOfDate:    req.AsOfDate,
			DataVersion: req.DataVersion,
		})
		if err != nil {
			return nil, fmt.Errorf("evaluator: load financials: %w", err)
		}
	}

	stocks := make([]string, 0, len(req.Universe))
	values := make([]value, 0, len(req.Universe))
	for _, code := range req.Universe {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		stockBars := bars[code]
		// We accept stocks with insufficient history; indicators surface NaN
		// and the reducer drops them. This keeps a single missing security
		// from killing a 5000-stock universe.
		stockFins := fins[code]
		ctxStock := newStockContext(code, stockBars, stockFins)
		v, evalErr := e.evalNode(ctxStock, plan.Root)
		if evalErr != nil {
			return nil, evalErr
		}
		stocks = append(stocks, code)
		values = append(values, v)
	}

	return reduceResult(plan, stocks, values), nil
}

const defaultLookback = 250

// evalNode walks the AST rooted at node and returns the typed per-stock value.
func (e *evaluator) evalNode(c *stockContext, node domainAST.Node) (value, error) {
	switch n := node.(type) {
	case *domainAST.Program:
		return e.evalProgram(c, n)
	case *domainAST.Assignment:
		return e.evalAssignment(c, n)
	case *domainAST.BinaryExpression:
		return e.evalBinary(c, n)
	case *domainAST.UnaryExpression:
		return e.evalUnary(c, n)
	case *domainAST.FunctionCall:
		return e.evalFunctionCall(c, n)
	case *domainAST.Identifier:
		return e.evalIdentifier(c, n)
	case *domainAST.NumberLiteral:
		return numberValue(n.Value), nil
	case *domainAST.BoolLiteral:
		return boolValue(n.Value), nil
	case *domainAST.StringLiteral:
		// Strings are not first-class values in the MVP DSL; surface a typed
		// error so the user is steered back to numeric/boolean expressions.
		return value{}, domainErrors.NewError(domainErrors.ErrTypeError,
			"evaluator: string literals are not supported in the current DSL")
	default:
		return value{}, domainErrors.NewError(domainErrors.ErrTypeError,
			fmt.Sprintf("evaluator: unsupported AST node %T", node))
	}
}

func (e *evaluator) evalProgram(c *stockContext, p *domainAST.Program) (value, error) {
	if len(p.Statements) == 0 {
		return value{}, domainErrors.NewError(domainErrors.ErrSyntaxError, "evaluator: empty program")
	}
	var last value
	for _, stmt := range p.Statements {
		v, err := e.evalNode(c, stmt)
		if err != nil {
			return value{}, err
		}
		last = v
	}
	return last, nil
}

func (e *evaluator) evalAssignment(c *stockContext, a *domainAST.Assignment) (value, error) {
	v, err := e.evalNode(c, a.Value)
	if err != nil {
		return value{}, err
	}
	c.locals[strings.ToLower(a.Name)] = v
	return v, nil
}

func (e *evaluator) evalIdentifier(c *stockContext, id *domainAST.Identifier) (value, error) {
	upper := strings.ToUpper(id.Name)
	if upper == "TRUE" {
		return boolValue(true), nil
	}
	if upper == "FALSE" {
		return boolValue(false), nil
	}

	if s, ok := c.resolveBuiltinSeries(id.Name); ok {
		return seriesValue(s), nil
	}

	if def, ok := e.variables.GetVariable(id.Name); ok {
		switch def.VarType {
		case domainVar.TypeNumber:
			if v, ok := c.resolveFinancial(def.Name); ok {
				return numberValue(v), nil
			}
			return numberValue(math.NaN()), nil
		case domainVar.TypeSeries:
			// Already handled by resolveBuiltinSeries; reaching here means
			// the registry advertises a series identifier that is not yet
			// surfaced by the data context. Surface NaN rather than failing.
			return seriesValue(series.Repeat(c.timestamps(), math.NaN())), nil
		}
	}

	if v, ok := c.locals[strings.ToLower(id.Name)]; ok {
		return v, nil
	}

	return value{}, domainErrors.NewError(domainErrors.ErrUnknownVariable,
		fmt.Sprintf("evaluator: unknown identifier %s", id.Name))
}

// emptyResult returns a zero-valued Result for a plan with no universe.
func emptyResult(planType domainCompiler.PlanType) *domainEval.Result {
	switch planType {
	case domainCompiler.PlanTypeFilter, domainCompiler.PlanTypeSignal:
		return &domainEval.Result{PlanType: planType, Selection: &domainEval.Selection{}}
	case domainCompiler.PlanTypeSort:
		return &domainEval.Result{PlanType: planType, Ranking: &domainEval.Ranking{}}
	default:
		return &domainEval.Result{PlanType: planType, Values: &domainEval.ValueMap{}}
	}
}

// reduceResult turns the per-stock values produced by the evaluator into the
// shape required by the plan type.
func reduceResult(plan *domainCompiler.ExecutionPlan, stocks []string, values []value) *domainEval.Result {
	switch plan.PlanType {
	case domainCompiler.PlanTypeFilter, domainCompiler.PlanTypeSignal:
		out := make([]string, 0, len(stocks))
		for i, v := range values {
			if v.asBool() {
				out = append(out, stocks[i])
			}
		}
		return &domainEval.Result{
			PlanType:  plan.PlanType,
			Selection: &domainEval.Selection{StockCodes: out},
		}
	case domainCompiler.PlanTypeSort:
		return reduceRanking(plan.PlanType, stocks, values)
	default:
		out := make([]float64, len(stocks))
		for i, v := range values {
			out[i] = v.asNumber()
		}
		return &domainEval.Result{
			PlanType: plan.PlanType,
			Values:   &domainEval.ValueMap{StockCodes: append([]string(nil), stocks...), Values: out},
		}
	}
}

func reduceRanking(planType domainCompiler.PlanType, stocks []string, values []value) *domainEval.Result {
	type pair struct {
		code  string
		score float64
	}
	pairs := make([]pair, 0, len(stocks))
	for i, v := range values {
		s := v.asNumber()
		if math.IsNaN(s) {
			continue
		}
		pairs = append(pairs, pair{code: stocks[i], score: s})
	}
	sort.SliceStable(pairs, func(i, j int) bool {
		return pairs[i].score > pairs[j].score
	})
	codes := make([]string, len(pairs))
	scores := make([]float64, len(pairs))
	for i, p := range pairs {
		codes[i] = p.code
		scores[i] = p.score
	}
	return &domainEval.Result{
		PlanType: planType,
		Ranking:  &domainEval.Ranking{StockCodes: codes, Scores: scores},
	}
}
