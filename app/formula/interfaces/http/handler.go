package http

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/agoXQ/QuantLab/app/formula/application/formula"
	domainCompiler "github.com/agoXQ/QuantLab/app/formula/domain/compiler"
	domainEval "github.com/agoXQ/QuantLab/app/formula/domain/evaluator"
)

// Handler exposes the Formula Service over HTTP.
//
// Evaluate is optional: when the EvaluatorService and DataPort are nil, the
// /evaluate route is not registered. This keeps the handler usable from the
// minimal HTTP test fixtures that only need compile/validate plumbing.
type Handler struct {
	svc       formula.Service
	evaluator formula.EvaluatorService
	screener  formula.ScreenService
	dataPort  domainEval.DataPort
}

// NewHandler creates a new HTTP handler with only the compile-side surface.
func NewHandler(svc formula.Service) *Handler {
	return &Handler{svc: svc}
}

// NewHandlerWithEvaluator wires the evaluation surface on top of the
// compile-side handler. The evaluator and dataPort are kept as separate
// dependencies so callers can swap the data port (in-memory vs repository
// vs gRPC) without rebuilding the EvaluatorService chain.
func NewHandlerWithEvaluator(
	svc formula.Service,
	evaluator formula.EvaluatorService,
	dataPort domainEval.DataPort,
) *Handler {
	return &Handler{svc: svc, evaluator: evaluator, dataPort: dataPort}
}

func NewHandlerWithScreen(
	svc formula.Service,
	evaluator formula.EvaluatorService,
	screener formula.ScreenService,
	dataPort domainEval.DataPort,
) *Handler {
	return &Handler{svc: svc, evaluator: evaluator, screener: screener, dataPort: dataPort}
}

// RegisterRoutes registers all formula API routes on the given gin engine.
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/validate", h.Validate)
	rg.POST("/compile", h.Compile)
	rg.POST("/ast", h.GetAST)
	rg.GET("/functions", h.ListFunctions)
	rg.GET("/functions/:name", h.GetFunction)
	if h.evaluator != nil && h.dataPort != nil {
		rg.POST("/evaluate", h.Evaluate)
	}
	rg.POST("/screen", h.Screen)
}

// --- Request / Response types ---

type formulaRequest struct {
	Formula string `json:"formula" binding:"required"`
}

type validateResponse struct {
	Valid     bool   `json:"valid"`
	ErrorCode int    `json:"error_code,omitempty"`
	Error     string `json:"error,omitempty"`
}

type compileResponse struct {
	ASTJSON   string `json:"ast_json"`
	PlanJSON  string `json:"plan_json"`
	Valid     bool   `json:"valid"`
	ErrorCode int    `json:"error_code,omitempty"`
	Error     string `json:"error,omitempty"`
}

type astResponse struct {
	ASTJSON string `json:"ast_json"`
}

type functionParam struct {
	Name      string `json:"name"`
	ParamType string `json:"param_type"`
}

type functionDefinition struct {
	Name        string          `json:"name"`
	Category    string          `json:"category"`
	ReturnType  string          `json:"return_type"`
	Description string          `json:"description"`
	Params      []functionParam `json:"params"`
}

type listFunctionsResponse struct {
	Functions []functionDefinition `json:"functions"`
}

type getFunctionResponse struct {
	Function *functionDefinition `json:"function,omitempty"`
}

type evaluateRequest struct {
	Formula      string   `json:"formula" binding:"required"`
	Universe     []string `json:"universe"`
	AsOfDate     string   `json:"as_of_date"`
	LookbackBars int      `json:"lookback_bars"`
	DataVersion  string   `json:"data_version"`
}

type universeFilterRequest struct {
	Market     string   `json:"market"`
	Exchange   string   `json:"exchange"`
	Industry   string   `json:"industry"`
	AssetType  string   `json:"asset_type"`
	Status     string   `json:"status"`
	StockCodes []string `json:"stock_codes"`
}

type screenRequest struct {
	Formula        string                `json:"formula" binding:"required"`
	AsOfDate       string                `json:"as_of_date"`
	LookbackBars   int                   `json:"lookback_bars"`
	DataVersion    string                `json:"data_version"`
	Limit          int                   `json:"limit"`
	UniverseFilter universeFilterRequest `json:"universe_filter"`
}

type rankingItem struct {
	StockCode string  `json:"stock_code"`
	Score     float64 `json:"score"`
}

type valueItem struct {
	StockCode string  `json:"stock_code"`
	Value     float64 `json:"value"`
}

type evaluateResponse struct {
	FormulaHash string        `json:"formula_hash"`
	PlanType    string        `json:"plan_type"`
	Selection   []string      `json:"selection,omitempty"`
	Ranking     []rankingItem `json:"ranking,omitempty"`
	Values      []valueItem   `json:"values,omitempty"`
}

type screenItemResponse struct {
	StockCode string   `json:"stock_code"`
	StockName string   `json:"stock_name"`
	Exchange  string   `json:"exchange"`
	Industry  string   `json:"industry"`
	Score     *float64 `json:"score,omitempty"`
	Selected  bool     `json:"selected"`
}

type screenResponse struct {
	FormulaHash  string               `json:"formula_hash"`
	PlanType     string               `json:"plan_type"`
	DataVersion  string               `json:"data_version"`
	UniverseSize int                  `json:"universe_size"`
	Items        []screenItemResponse `json:"items"`
}

// --- Handlers ---

// Validate handles POST /api/v1/formula/validate
func (h *Handler) Validate(c *gin.Context) {
	var req formulaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	result, err := h.svc.Validate(c.Request.Context(), req.Formula)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	resp := validateResponse{Valid: result.Valid}
	if !result.Valid && len(result.Errors) > 0 {
		resp.ErrorCode = result.Errors[0].Code
		resp.Error = result.Errors[0].Message
	}

	c.JSON(http.StatusOK, resp)
}

// Compile handles POST /api/v1/formula/compile
func (h *Handler) Compile(c *gin.Context) {
	var req formulaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	result, err := h.svc.Compile(c.Request.Context(), req.Formula)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	resp := compileResponse{
		Valid:     result.Valid,
		ErrorCode: result.ErrorCode,
		Error:     result.ErrorMsg,
	}
	if result.AST != nil {
		astJSON, _ := json.Marshal(result.AST)
		resp.ASTJSON = string(astJSON)
	}
	if result.Plan != nil {
		planJSON, _ := json.Marshal(result.Plan)
		resp.PlanJSON = string(planJSON)
	}

	c.JSON(http.StatusOK, resp)
}

// GetAST handles POST /api/v1/formula/ast
func (h *Handler) GetAST(c *gin.Context) {
	var req formulaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	node, err := h.svc.GetAST(c.Request.Context(), req.Formula)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	astJSON, _ := json.Marshal(node)
	c.JSON(http.StatusOK, astResponse{ASTJSON: string(astJSON)})
}

// ListFunctions handles GET /api/v1/formula/functions
func (h *Handler) ListFunctions(c *gin.Context) {
	defs, err := h.svc.ListFunctions(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	fns := make([]functionDefinition, 0, len(defs))
	for _, def := range defs {
		params := make([]functionParam, len(def.Args))
		for i, arg := range def.Args {
			params[i] = functionParam{Name: arg.Name, ParamType: arg.ArgType}
		}
		fns = append(fns, functionDefinition{
			Name:        def.Name,
			Category:    def.Category,
			ReturnType:  def.ReturnType,
			Description: def.Description,
			Params:      params,
		})
	}

	c.JSON(http.StatusOK, listFunctionsResponse{Functions: fns})
}

// GetFunction handles GET /api/v1/formula/functions/:name
func (h *Handler) GetFunction(c *gin.Context) {
	name := c.Param("name")

	def, err := h.svc.GetFunction(c.Request.Context(), name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if def == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "function not found"})
		return
	}

	params := make([]functionParam, len(def.Args))
	for i, arg := range def.Args {
		params[i] = functionParam{Name: arg.Name, ParamType: arg.ArgType}
	}

	c.JSON(http.StatusOK, getFunctionResponse{
		Function: &functionDefinition{
			Name:        def.Name,
			Category:    def.Category,
			ReturnType:  def.ReturnType,
			Description: def.Description,
			Params:      params,
		},
	})
}

// Evaluate handles POST /api/v1/formula/evaluate.
//
// The request runs through the same Compile pipeline (cache / log / event /
// metrics) as /compile, then dispatches the produced plan into the AST
// evaluator using the data port wired at boot. Universe is required so the
// evaluator can materialise a deterministic result; AsOfDate defaults to
// "now" when omitted, matching the application-layer EvaluatorService.
func (h *Handler) Evaluate(c *gin.Context) {
	if h.evaluator == nil || h.dataPort == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "evaluator not configured"})
		return
	}

	var req evaluateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	if len(req.Universe) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "universe is required"})
		return
	}

	asOf := time.Now()
	if req.AsOfDate != "" {
		parsed, err := parseAsOfDate(req.AsOfDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid as_of_date: " + err.Error()})
			return
		}
		asOf = parsed
	}

	result, err := h.evaluator.Evaluate(c.Request.Context(), formula.EvaluateRequest{
		Formula:      req.Formula,
		Universe:     req.Universe,
		AsOfDate:     asOf,
		LookbackBars: req.LookbackBars,
		DataVersion:  req.DataVersion,
		DataPort:     h.dataPort,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, buildEvaluateResponse(result))
}

func (h *Handler) Screen(c *gin.Context) {
	if h.screener == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "screener not configured"})
		return
	}
	var req screenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	asOf := time.Now()
	if req.AsOfDate != "" {
		parsed, err := parseAsOfDate(req.AsOfDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid as_of_date: " + err.Error()})
			return
		}
		asOf = parsed
	}
	res, err := h.screener.Screen(c.Request.Context(), formula.ScreenRequest{
		Formula:      req.Formula,
		AsOfDate:     asOf,
		LookbackBars: req.LookbackBars,
		DataVersion:  req.DataVersion,
		Limit:        req.Limit,
		UniverseFilter: formula.UniverseFilter{
			Market:     req.UniverseFilter.Market,
			Exchange:   req.UniverseFilter.Exchange,
			Industry:   req.UniverseFilter.Industry,
			AssetType:  req.UniverseFilter.AssetType,
			Status:     req.UniverseFilter.Status,
			StockCodes: req.UniverseFilter.StockCodes,
		},
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, buildScreenResponse(res))
}

// parseAsOfDate accepts the two shapes the rest of the platform uses: a
// date-only "2006-01-02" string for cross-section requests, and full
// RFC3339 for replay / audit traffic. Anything else surfaces as a 400 to
// the caller rather than silently defaulting to "now".
func parseAsOfDate(raw string) (time.Time, error) {
	if t, err := time.Parse("2006-01-02", raw); err == nil {
		return t, nil
	}
	return time.Parse(time.RFC3339, raw)
}

func buildEvaluateResponse(res *formula.EvaluateResult) evaluateResponse {
	resp := evaluateResponse{FormulaHash: res.FormulaHash}
	if res.Result == nil {
		return resp
	}
	resp.PlanType = string(res.Result.PlanType)
	switch res.Result.PlanType {
	case domainCompiler.PlanTypeFilter, domainCompiler.PlanTypeSignal:
		if res.Result.Selection != nil {
			resp.Selection = res.Result.Selection.StockCodes
		}
	case domainCompiler.PlanTypeSort:
		if res.Result.Ranking != nil {
			items := make([]rankingItem, len(res.Result.Ranking.StockCodes))
			for i, code := range res.Result.Ranking.StockCodes {
				items[i] = rankingItem{StockCode: code, Score: res.Result.Ranking.Scores[i]}
			}
			resp.Ranking = items
		}
	case domainCompiler.PlanTypeValue:
		if res.Result.Values != nil {
			items := make([]valueItem, len(res.Result.Values.StockCodes))
			for i, code := range res.Result.Values.StockCodes {
				items[i] = valueItem{StockCode: code, Value: res.Result.Values.Values[i]}
			}
			resp.Values = items
		}
	}
	return resp
}

func buildScreenResponse(res *formula.ScreenResult) screenResponse {
	if res == nil {
		return screenResponse{}
	}
	items := make([]screenItemResponse, 0, len(res.Items))
	for _, item := range res.Items {
		items = append(items, screenItemResponse{
			StockCode: item.StockCode,
			StockName: item.StockName,
			Exchange:  item.Exchange,
			Industry:  item.Industry,
			Score:     item.Score,
			Selected:  item.Selected,
		})
	}
	return screenResponse{
		FormulaHash:  res.FormulaHash,
		PlanType:     res.PlanType,
		DataVersion:  res.DataVersion,
		UniverseSize: res.UniverseSize,
		Items:        items,
	}
}
