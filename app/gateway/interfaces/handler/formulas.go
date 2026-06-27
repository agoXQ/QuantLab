package handler

import (
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/gin-gonic/gin"

	formulapb "github.com/agoXQ/QuantLab/app/formula/pb"
	"github.com/agoXQ/QuantLab/pkg/errors"
	"github.com/agoXQ/QuantLab/pkg/response"
)

// registerFormulas mounts /api/v1/formulas. These power the Monaco DSL
// editor: validate/compile for live syntax checks, functions for
// autocomplete and hover docs.
func (h *Handler) registerFormulas(rg *gin.RouterGroup) {
	f := rg.Group("/formulas")
	f.POST("/validate", h.formulaValidate)
	f.POST("/compile", h.formulaCompile)
	f.POST("/ast", h.formulaAST)
	f.GET("/functions", h.formulaListFunctions)
	f.GET("/functions/:name", h.formulaGetFunction)
	// Evaluate is proxied to the formula service's own HTTP server
	// because Evaluate is not in the gRPC proto. When FormulaHTTPAddr
	// is empty the route returns 503.
	f.POST("/evaluate", h.formulaEvaluate)
	f.POST("/screen", h.formulaScreen)
}

func (h *Handler) formulaValidate(c *gin.Context) {
	var req struct {
		Formula string `json:"formula" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, errors.ErrInvalidParam)
		return
	}
	out, err := h.svc.Formula.Validate(c.Request.Context(), &formulapb.ValidateRequest{Formula: req.Formula})
	if err != nil {
		grpcErr(c, err)
		return
	}
	response.OK(c, out)
}

func (h *Handler) formulaCompile(c *gin.Context) {
	var req struct {
		Formula string `json:"formula" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, errors.ErrInvalidParam)
		return
	}
	out, err := h.svc.Formula.Compile(c.Request.Context(), &formulapb.CompileRequest{Formula: req.Formula})
	if err != nil {
		grpcErr(c, err)
		return
	}
	response.OK(c, out)
}

func (h *Handler) formulaAST(c *gin.Context) {
	var req struct {
		Formula string `json:"formula" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, errors.ErrInvalidParam)
		return
	}
	out, err := h.svc.Formula.GetAST(c.Request.Context(), &formulapb.GetASTRequest{Formula: req.Formula})
	if err != nil {
		grpcErr(c, err)
		return
	}
	response.OK(c, gin.H{"ast_json": out.AstJson})
}

func (h *Handler) formulaListFunctions(c *gin.Context) {
	out, err := h.svc.Formula.ListFunctions(c.Request.Context(), &formulapb.ListFunctionsRequest{})
	if err != nil {
		grpcErr(c, err)
		return
	}
	response.OK(c, gin.H{"items": out.Functions})
}

func (h *Handler) formulaGetFunction(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		response.Error(c, http.StatusBadRequest, errors.ErrInvalidParam)
		return
	}
	out, err := h.svc.Formula.GetFunction(c.Request.Context(), &formulapb.GetFunctionRequest{Name: name})
	if err != nil {
		grpcErr(c, err)
		return
	}
	response.OK(c, gin.H{"function": out.Function})
}

// formulaEvaluate reverse-proxies POST /api/v1/formulas/evaluate to the
// formula service's HTTP server. The formula service owns the evaluator
// (AST + data port); the gateway just forwards the JSON body and passes
// the response through unchanged.
func (h *Handler) formulaEvaluate(c *gin.Context) {
	h.proxyFormulaHTTP(c, "/api/v1/formula/evaluate")
}

func (h *Handler) formulaScreen(c *gin.Context) {
	h.proxyFormulaHTTP(c, "/api/v1/formula/screen")
}

func (h *Handler) proxyFormulaHTTP(c *gin.Context, targetPath string) {
	if h.svc.FormulaHTTPAddr == "" {
		response.Error(c, http.StatusServiceUnavailable, errors.New(50010, "formula http not configured"))
		return
	}
	target, err := url.Parse(h.svc.FormulaHTTPAddr)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, errors.ErrInternal)
		return
	}
	proxy := httputil.NewSingleHostReverseProxy(target)
	original := c.Request.URL.Path
	c.Request.URL.Path = targetPath
	proxy.ModifyResponse = func(resp *http.Response) error {
		// The formula service returns a bare JSON object (no envelope);
		// wrap it so the frontend's client interceptor handles it uniformly.
		return nil
	}
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		response.Error(c, http.StatusBadGateway, errors.New(50011, "formula service unreachable: "+err.Error()))
	}
	// Restore path after proxy (cleanup for any middleware that runs after).
	defer func() { c.Request.URL.Path = original }()
	proxy.ServeHTTP(c.Writer, c.Request)
}
