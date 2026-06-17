package http

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/agoXQ/QuantLab/app/formula/application/formula"
)

// Handler exposes the Formula Service over HTTP.
type Handler struct {
	svc formula.Service
}

// NewHandler creates a new HTTP handler.
func NewHandler(svc formula.Service) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes registers all formula API routes on the given gin engine.
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/validate", h.Validate)
	rg.POST("/compile", h.Compile)
	rg.POST("/ast", h.GetAST)
	rg.GET("/functions", h.ListFunctions)
	rg.GET("/functions/:name", h.GetFunction)
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
