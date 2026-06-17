package variable

import (
	"strings"
	"sync"

	domainVar "github.com/agoXQ/QuantLab/app/formula/domain/variable"
)

type registry struct {
	mu       sync.RWMutex
	byLower  map[string]domainVar.VariableDefinition
	aliases  map[string]string // short name -> canonical name
}

// NewRegistry creates a new in-memory variable registry with all built-in variables.
func NewRegistry() domainVar.Registry {
	r := &registry{
		byLower: make(map[string]domainVar.VariableDefinition),
		aliases: make(map[string]string),
	}
	r.registerBuiltins()
	return r
}

func (r *registry) GetVariable(name string) (domainVar.VariableDefinition, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	lower := strings.ToLower(name)
	// Check direct lookup first
	if def, ok := r.byLower[lower]; ok {
		return def, ok
	}
	// Check aliases
	if canonical, ok := r.aliases[lower]; ok {
		def, ok := r.byLower[canonical]
		return def, ok
	}
	return domainVar.VariableDefinition{}, false
}

func (r *registry) ListVariables() []domainVar.VariableDefinition {
	r.mu.RLock()
	defer r.mu.RUnlock()
	defs := make([]domainVar.VariableDefinition, 0, len(r.byLower))
	for _, def := range r.byLower {
		defs = append(defs, def)
	}
	return defs
}

func (r *registry) Exists(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	lower := strings.ToLower(name)
	if _, ok := r.byLower[lower]; ok {
		return true
	}
	_, ok := r.aliases[lower]
	return ok
}

func (r *registry) registerBuiltins() {
	builtins := []domainVar.VariableDefinition{
		{Name: "OPEN", VarType: domainVar.TypeSeries, Category: "MarketData", Description: "Opening price"},
		{Name: "HIGH", VarType: domainVar.TypeSeries, Category: "MarketData", Description: "Highest price"},
		{Name: "LOW", VarType: domainVar.TypeSeries, Category: "MarketData", Description: "Lowest price"},
		{Name: "CLOSE", VarType: domainVar.TypeSeries, Category: "MarketData", Description: "Closing price"},
		{Name: "VOL", VarType: domainVar.TypeSeries, Category: "MarketData", Description: "Volume"},
		{Name: "AMOUNT", VarType: domainVar.TypeSeries, Category: "MarketData", Description: "Trading amount"},
		{Name: "PE", VarType: domainVar.TypeNumber, Category: "Financial", Description: "Price-to-Earnings ratio"},
		{Name: "PB", VarType: domainVar.TypeNumber, Category: "Financial", Description: "Price-to-Book ratio"},
		{Name: "PS", VarType: domainVar.TypeNumber, Category: "Financial", Description: "Price-to-Sales ratio"},
		{Name: "ROE", VarType: domainVar.TypeNumber, Category: "Financial", Description: "Return on Equity"},
		{Name: "ROA", VarType: domainVar.TypeNumber, Category: "Financial", Description: "Return on Assets"},
		{Name: "EPS", VarType: domainVar.TypeNumber, Category: "Financial", Description: "Earnings Per Share"},
		{Name: "RevenueGrowth", VarType: domainVar.TypeNumber, Category: "Growth", Description: "Revenue growth rate"},
		{Name: "ProfitGrowth", VarType: domainVar.TypeNumber, Category: "Growth", Description: "Profit growth rate"},
		{Name: "MarketCap", VarType: domainVar.TypeNumber, Category: "MarketCap", Description: "Total market capitalization"},
		{Name: "FloatMarketCap", VarType: domainVar.TypeNumber, Category: "MarketCap", Description: "Float market capitalization"},
	}
	for _, def := range builtins {
		r.byLower[strings.ToLower(def.Name)] = def
	}

	// Register short aliases (C, O, H, L, V)
	aliasList := []struct{ short, full string }{
		{"C", "CLOSE"},
		{"O", "OPEN"},
		{"H", "HIGH"},
		{"L", "LOW"},
		{"V", "VOL"},
	}
	for _, a := range aliasList {
		r.aliases[strings.ToLower(a.short)] = strings.ToLower(a.full)
	}
}
