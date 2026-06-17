package tests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	httpHandler "github.com/agoXQ/QuantLab/app/formula/interfaces/http"
	appFormula "github.com/agoXQ/QuantLab/app/formula/application/formula"
	infraFunc "github.com/agoXQ/QuantLab/app/formula/infrastructure/function"
	infraLexer "github.com/agoXQ/QuantLab/app/formula/infrastructure/lexer"
	infraOptimizer "github.com/agoXQ/QuantLab/app/formula/infrastructure/optimizer"
	infraParser "github.com/agoXQ/QuantLab/app/formula/infrastructure/parser"
	infraPlanner "github.com/agoXQ/QuantLab/app/formula/infrastructure/planner"
	infraValidator "github.com/agoXQ/QuantLab/app/formula/infrastructure/validator"
	infraVar "github.com/agoXQ/QuantLab/app/formula/infrastructure/variable"
)

func newHTTPService() appFormula.Service {
	funcReg := infraFunc.NewRegistry()
	varReg := infraVar.NewRegistry()

	return appFormula.NewService(
		infraLexer.NewLexer(),
		infraParser.NewParser(funcReg, varReg),
		infraValidator.NewValidator(funcReg, varReg),
		infraOptimizer.NewOptimizer(),
		infraPlanner.NewPlanner(),
		funcReg,
	)
}

func setupHTTPRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	svc := newHTTPService()
	handler := httpHandler.NewHandler(svc)
	apiGroup := router.Group("/api/v1/formula")
	handler.RegisterRoutes(apiGroup)
	return router
}

func executeRequest(router *gin.Engine, method, path, body string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	return w
}

func TestHTTP_Validate_Valid(t *testing.T) {
	router := setupHTTPRouter()
	w := executeRequest(router, "POST", "/api/v1/formula/validate", `{"formula":"ROE > 15"}`)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp struct {
		Valid bool `json:"valid"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if !resp.Valid {
		t.Errorf("expected valid=true, got false")
	}
}

func TestHTTP_Validate_Invalid(t *testing.T) {
	router := setupHTTPRouter()
	w := executeRequest(router, "POST", "/api/v1/formula/validate", `{"formula":"UNKNOWN_VAR > 15"}`)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp struct {
		Valid     bool   `json:"valid"`
		ErrorCode int    `json:"error_code"`
		Error     string `json:"error"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if resp.Valid {
		t.Errorf("expected valid=false, got true")
	}
	if resp.ErrorCode == 0 {
		t.Errorf("expected non-zero error_code")
	}
	if resp.Error == "" {
		t.Errorf("expected non-empty error message")
	}
}

func TestHTTP_Validate_EmptyBody(t *testing.T) {
	router := setupHTTPRouter()
	w := executeRequest(router, "POST", "/api/v1/formula/validate", `{}`)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestHTTP_Compile_Valid(t *testing.T) {
	router := setupHTTPRouter()
	w := executeRequest(router, "POST", "/api/v1/formula/compile", `{"formula":"ROE > 15"}`)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp struct {
		ASTJSON  string `json:"ast_json"`
		PlanJSON string `json:"plan_json"`
		Valid    bool   `json:"valid"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if !resp.Valid {
		t.Errorf("expected valid=true, got false")
	}
	if resp.ASTJSON == "" {
		t.Errorf("expected non-empty ast_json")
	}
	if resp.PlanJSON == "" {
		t.Errorf("expected non-empty plan_json")
	}
}

func TestHTTP_Compile_Invalid(t *testing.T) {
	router := setupHTTPRouter()
	w := executeRequest(router, "POST", "/api/v1/formula/compile", `{"formula":"UNKNOWN_VAR > 15"}`)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp struct {
		Valid     bool   `json:"valid"`
		ErrorCode int    `json:"error_code"`
		Error     string `json:"error"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if resp.Valid {
		t.Errorf("expected valid=false, got true")
	}
	if resp.Error == "" {
		t.Errorf("expected non-empty error message")
	}
}

func TestHTTP_GetAST(t *testing.T) {
	router := setupHTTPRouter()
	w := executeRequest(router, "POST", "/api/v1/formula/ast", `{"formula":"ROE > 15"}`)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp struct {
		ASTJSON string `json:"ast_json"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if resp.ASTJSON == "" {
		t.Errorf("expected non-empty ast_json")
	}
}

func TestHTTP_ListFunctions(t *testing.T) {
	router := setupHTTPRouter()
	w := executeRequest(router, "GET", "/api/v1/formula/functions", "")

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp struct {
		Functions []struct {
			Name string `json:"name"`
		} `json:"functions"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if len(resp.Functions) == 0 {
		t.Errorf("expected non-empty functions list")
	}
}

func TestHTTP_GetFunction_Found(t *testing.T) {
	router := setupHTTPRouter()
	w := executeRequest(router, "GET", "/api/v1/formula/functions/MA", "")

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp struct {
		Function *struct {
			Name string `json:"name"`
		} `json:"function"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if resp.Function == nil {
		t.Fatal("expected non-nil function")
	}
	if resp.Function.Name != "MA" {
		t.Errorf("expected name=MA, got %s", resp.Function.Name)
	}
}

func TestHTTP_GetFunction_NotFound(t *testing.T) {
	router := setupHTTPRouter()
	w := executeRequest(router, "GET", "/api/v1/formula/functions/NONEXISTENT", "")

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}
