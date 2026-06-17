package function

// ArgDef defines a single argument for a function.
type ArgDef struct {
	Name     string `json:"name"`
	ArgType  string `json:"arg_type"`
	Required bool   `json:"required"`
}

// FunctionDefinition defines a built-in or user-registered function.
type FunctionDefinition struct {
	Name        string   `json:"name"`
	Category    string   `json:"category"`
	ReturnType  string   `json:"return_type"`
	Description string   `json:"description"`
	Args        []ArgDef `json:"args"`
}

// Function categories.
const (
	CategoryTechnical    = "Technical"
	CategoryFinancial    = "Financial"
	CategoryLogical      = "Logical"
	CategoryMath         = "Math"
	CategoryTimeSeries   = "TimeSeries"
	CategorySignal       = "Signal"
	CategoryGrowth       = "Growth"
	CategoryMarketCap    = "MarketCap"
	CategoryMarketData   = "MarketData"
)

// Return types.
const (
	TypeNumber  = "Number"
	TypeBoolean = "Boolean"
	TypeSeries  = "Series"
	TypeSignal  = "Signal"
)
