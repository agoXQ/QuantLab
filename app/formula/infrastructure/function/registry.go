package function

import (
	"fmt"
	"strings"
	"sync"

	domainFunc "github.com/agoXQ/QuantLab/app/formula/domain/function"
)

// registry implements the domain function.Registry interface.
type registry struct {
	mu       sync.RWMutex
	functions map[string]domainFunc.FunctionDefinition
	byLower   map[string]string // lowercase name -> canonical name
}

// NewRegistry creates a new in-memory function registry and registers all built-in functions.
func NewRegistry() domainFunc.Registry {
	r := &registry{
		functions: make(map[string]domainFunc.FunctionDefinition),
		byLower:   make(map[string]string),
	}
	r.registerBuiltins()
	return r
}

func (r *registry) GetFunction(name string) (domainFunc.FunctionDefinition, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	canonical, ok := r.byLower[strings.ToLower(name)]
	if !ok {
		return domainFunc.FunctionDefinition{}, false
	}
	def, ok := r.functions[canonical]
	return def, ok
}

func (r *registry) ListFunctions() []domainFunc.FunctionDefinition {
	r.mu.RLock()
	defer r.mu.RUnlock()

	defs := make([]domainFunc.FunctionDefinition, 0, len(r.functions))
	for _, def := range r.functions {
		defs = append(defs, def)
	}
	return defs
}

func (r *registry) RegisterFunction(def domainFunc.FunctionDefinition) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	lower := strings.ToLower(def.Name)
	if _, exists := r.byLower[lower]; exists {
		return fmt.Errorf("function already registered: %s", def.Name)
	}

	r.functions[def.Name] = def
	r.byLower[lower] = def.Name
	return nil
}

func (r *registry) Exists(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, ok := r.byLower[strings.ToLower(name)]
	return ok
}

func (r *registry) ResolveName(name string) (string, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	canonical, ok := r.byLower[strings.ToLower(name)]
	return canonical, ok
}

func (r *registry) registerBuiltins() {
	builtins := []domainFunc.FunctionDefinition{
		// --- Technical Indicators ---
		{
			Name: "MA", Category: domainFunc.CategoryTechnical, ReturnType: domainFunc.TypeSeries,
			Description: "Moving average: MA(series, period)",
			Args: []domainFunc.ArgDef{
				{Name: "series", ArgType: "Series", Required: true},
				{Name: "period", ArgType: "Number", Required: true},
			},
		},
		{
			Name: "EMA", Category: domainFunc.CategoryTechnical, ReturnType: domainFunc.TypeSeries,
			Description: "Exponential moving average: EMA(series, period)",
			Args: []domainFunc.ArgDef{
				{Name: "series", ArgType: "Series", Required: true},
				{Name: "period", ArgType: "Number", Required: true},
			},
		},
		{
			Name: "SMA", Category: domainFunc.CategoryTechnical, ReturnType: domainFunc.TypeSeries,
			Description: "Smoothed moving average: SMA(series, period)",
			Args: []domainFunc.ArgDef{
				{Name: "series", ArgType: "Series", Required: true},
				{Name: "period", ArgType: "Number", Required: true},
			},
		},
		{
			Name: "STD", Category: domainFunc.CategoryTechnical, ReturnType: domainFunc.TypeSeries,
			Description: "Standard deviation: STD(series, period)",
			Args: []domainFunc.ArgDef{
				{Name: "series", ArgType: "Series", Required: true},
				{Name: "period", ArgType: "Number", Required: true},
			},
		},
		{
			Name: "MACD", Category: domainFunc.CategoryTechnical, ReturnType: domainFunc.TypeSeries,
			Description: "MACD indicator: MACD(series, fast, slow, signal)",
			Args: []domainFunc.ArgDef{
				{Name: "series", ArgType: "Series", Required: true},
				{Name: "fast", ArgType: "Number", Required: true},
				{Name: "slow", ArgType: "Number", Required: true},
				{Name: "signal", ArgType: "Number", Required: true},
			},
		},
		{
			Name: "RSI", Category: domainFunc.CategoryTechnical, ReturnType: domainFunc.TypeSeries,
			Description: "Relative Strength Index: RSI(series, period)",
			Args: []domainFunc.ArgDef{
				{Name: "series", ArgType: "Series", Required: true},
				{Name: "period", ArgType: "Number", Required: true},
			},
		},
		{
			Name: "KDJ", Category: domainFunc.CategoryTechnical, ReturnType: domainFunc.TypeSeries,
			Description: "KDJ indicator: KDJ(series, n, m1, m2)",
			Args: []domainFunc.ArgDef{
				{Name: "series", ArgType: "Series", Required: true},
				{Name: "n", ArgType: "Number", Required: true},
				{Name: "m1", ArgType: "Number", Required: true},
				{Name: "m2", ArgType: "Number", Required: true},
			},
		},
		{
			Name: "BOLL", Category: domainFunc.CategoryTechnical, ReturnType: domainFunc.TypeSeries,
			Description: "Bollinger Bands: BOLL(series, n, k)",
			Args: []domainFunc.ArgDef{
				{Name: "series", ArgType: "Series", Required: true},
				{Name: "n", ArgType: "Number", Required: true},
				{Name: "k", ArgType: "Number", Required: true},
			},
		},
		{
			Name: "ATR", Category: domainFunc.CategoryTechnical, ReturnType: domainFunc.TypeSeries,
			Description: "Average True Range: ATR(series, period)",
			Args: []domainFunc.ArgDef{
				{Name: "series", ArgType: "Series", Required: true},
				{Name: "period", ArgType: "Number", Required: true},
			},
		},

		// --- Math Functions ---
		{
			Name: "ABS", Category: domainFunc.CategoryMath, ReturnType: domainFunc.TypeSeries,
			Description: "Absolute value: ABS(x)",
			Args: []domainFunc.ArgDef{
				{Name: "x", ArgType: "Series", Required: true},
			},
		},
		{
			Name: "MAX", Category: domainFunc.CategoryMath, ReturnType: domainFunc.TypeSeries,
			Description: "Maximum of two values: MAX(a, b)",
			Args: []domainFunc.ArgDef{
				{Name: "a", ArgType: "Series", Required: true},
				{Name: "b", ArgType: "Series", Required: true},
			},
		},
		{
			Name: "MIN", Category: domainFunc.CategoryMath, ReturnType: domainFunc.TypeSeries,
			Description: "Minimum of two values: MIN(a, b)",
			Args: []domainFunc.ArgDef{
				{Name: "a", ArgType: "Series", Required: true},
				{Name: "b", ArgType: "Series", Required: true},
			},
		},

		// --- Statistical Functions ---
		{
			Name: "SUM", Category: domainFunc.CategoryMath, ReturnType: domainFunc.TypeSeries,
			Description: "Sum over period: SUM(series, n)",
			Args: []domainFunc.ArgDef{
				{Name: "series", ArgType: "Series", Required: true},
				{Name: "n", ArgType: "Number", Required: true},
			},
		},
		{
			Name: "AVG", Category: domainFunc.CategoryMath, ReturnType: domainFunc.TypeSeries,
			Description: "Average over period: AVG(series, n)",
			Args: []domainFunc.ArgDef{
				{Name: "series", ArgType: "Series", Required: true},
				{Name: "n", ArgType: "Number", Required: true},
			},
		},
		{
			Name: "COUNT", Category: domainFunc.CategoryMath, ReturnType: domainFunc.TypeNumber,
			Description: "Count occurrences where condition is true: COUNT(condition, n)",
			Args: []domainFunc.ArgDef{
				{Name: "condition", ArgType: "Boolean", Required: true},
				{Name: "n", ArgType: "Number", Required: true},
			},
		},

		// --- Time Series Functions ---
		{
			Name: "REF", Category: domainFunc.CategoryTimeSeries, ReturnType: domainFunc.TypeSeries,
			Description: "Reference previous value: REF(series, n)",
			Args: []domainFunc.ArgDef{
				{Name: "series", ArgType: "Series", Required: true},
				{Name: "n", ArgType: "Number", Required: true},
			},
		},
		{
			Name: "HHV", Category: domainFunc.CategoryTimeSeries, ReturnType: domainFunc.TypeSeries,
			Description: "Highest high over period: HHV(series, n)",
			Args: []domainFunc.ArgDef{
				{Name: "series", ArgType: "Series", Required: true},
				{Name: "n", ArgType: "Number", Required: true},
			},
		},
		{
			Name: "LLV", Category: domainFunc.CategoryTimeSeries, ReturnType: domainFunc.TypeSeries,
			Description: "Lowest low over period: LLV(series, n)",
			Args: []domainFunc.ArgDef{
				{Name: "series", ArgType: "Series", Required: true},
				{Name: "n", ArgType: "Number", Required: true},
			},
		},
		{
			Name: "BARSLAST", Category: domainFunc.CategoryTimeSeries, ReturnType: domainFunc.TypeNumber,
			Description: "Bars since condition was true: BARSLAST(condition)",
			Args: []domainFunc.ArgDef{
				{Name: "condition", ArgType: "Boolean", Required: true},
			},
		},

		// --- Signal Functions ---
		{
			Name: "CROSS", Category: domainFunc.CategorySignal, ReturnType: domainFunc.TypeSignal,
			Description: "Cross above: CROSS(a, b) - true when a crosses above b",
			Args: []domainFunc.ArgDef{
				{Name: "a", ArgType: "Series", Required: true},
				{Name: "b", ArgType: "Series", Required: true},
			},
		},
		{
			Name: "LONGCROSS", Category: domainFunc.CategorySignal, ReturnType: domainFunc.TypeSignal,
			Description: "Long cross: LONGCROSS(a, b, n) - a stays above b for n periods",
			Args: []domainFunc.ArgDef{
				{Name: "a", ArgType: "Series", Required: true},
				{Name: "b", ArgType: "Series", Required: true},
				{Name: "n", ArgType: "Number", Required: true},
			},
		},
		{
			Name: "FILTER", Category: domainFunc.CategorySignal, ReturnType: domainFunc.TypeSignal,
			Description: "Filter signal: FILTER(signal, n) - keep signal for n bars",
			Args: []domainFunc.ArgDef{
				{Name: "signal", ArgType: "Signal", Required: true},
				{Name: "n", ArgType: "Number", Required: true},
			},
		},
	}

	for _, def := range builtins {
		r.functions[def.Name] = def
		r.byLower[strings.ToLower(def.Name)] = def.Name
	}
}
