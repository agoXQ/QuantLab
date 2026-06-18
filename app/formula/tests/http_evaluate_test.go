package tests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	httpHandler "github.com/agoXQ/QuantLab/app/formula/interfaces/http"
)

// setupEvaluateRouter wires the handler with the evaluator surface enabled,
// using the same in-memory data port the evaluator unit tests rely on so we
// exercise the HTTP layer without dragging Postgres into the loop.
func setupEvaluateRouter(t *testing.T) (*gin.Engine, *evaluateFixture) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	router := gin.New()

	evalSvc, port := newEvaluatorService(t)
	handler := httpHandler.NewHandlerWithEvaluator(newHTTPService(), evalSvc, port)
	apiGroup := router.Group("/api/v1/formula")
	handler.RegisterRoutes(apiGroup)

	return router, &evaluateFixture{port: port}
}

type evaluateFixture struct {
	port interface {
		SetFinancials(string, map[string]float64)
	}
}

func TestHTTP_Evaluate_FilterByFinancials(t *testing.T) {
	router, fx := setupEvaluateRouter(t)
	fx.port.SetFinancials("000001", map[string]float64{"ROE": 20, "PE": 12})
	fx.port.SetFinancials("000002", map[string]float64{"ROE": 5, "PE": 35})
	fx.port.SetFinancials("000003", map[string]float64{"ROE": 18, "PE": 18})

	body := `{"formula":"ROE > 15 AND PE < 20","universe":["000001","000002","000003"],"as_of_date":"2025-01-01"}`
	w := executeRequest(router, "POST", "/api/v1/formula/evaluate", body)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", w.Code, w.Body.String())
	}

	var resp struct {
		FormulaHash string   `json:"formula_hash"`
		PlanType    string   `json:"plan_type"`
		Selection   []string `json:"selection"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if resp.PlanType != "FILTER" {
		t.Errorf("expected plan_type=FILTER, got %s", resp.PlanType)
	}
	if resp.FormulaHash == "" {
		t.Error("expected non-empty formula_hash")
	}
	if len(resp.Selection) != 2 || resp.Selection[0] != "000001" || resp.Selection[1] != "000003" {
		t.Errorf("expected [000001 000003], got %v", resp.Selection)
	}
}

func TestHTTP_Evaluate_MissingUniverse(t *testing.T) {
	router, _ := setupEvaluateRouter(t)
	w := executeRequest(router, "POST", "/api/v1/formula/evaluate", `{"formula":"ROE > 15"}`)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for missing universe, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "universe") {
		t.Errorf("expected error to mention universe, got %s", w.Body.String())
	}
}

func TestHTTP_Evaluate_MissingFormula(t *testing.T) {
	router, _ := setupEvaluateRouter(t)
	w := executeRequest(router, "POST", "/api/v1/formula/evaluate", `{"universe":["000001"]}`)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for missing formula, got %d", w.Code)
	}
}

func TestHTTP_Evaluate_InvalidAsOfDate(t *testing.T) {
	router, _ := setupEvaluateRouter(t)
	body := `{"formula":"ROE > 15","universe":["000001"],"as_of_date":"not-a-date"}`
	w := executeRequest(router, "POST", "/api/v1/formula/evaluate", body)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid date, got %d", w.Code)
	}
}

func TestHTTP_Evaluate_InvalidFormula(t *testing.T) {
	router, _ := setupEvaluateRouter(t)
	body := `{"formula":"UNKNOWN_VAR > 15","universe":["000001"],"as_of_date":"2025-01-01"}`
	w := executeRequest(router, "POST", "/api/v1/formula/evaluate", body)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid formula, got %d", w.Code)
	}
}

func TestHTTP_Evaluate_NotConfigured(t *testing.T) {
	// When the handler is built without an evaluator, the route is not
	// registered at all; gin returns 404 rather than 503.
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := httpHandler.NewHandler(newHTTPService())
	handler.RegisterRoutes(router.Group("/api/v1/formula"))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/formula/evaluate",
		strings.NewReader(`{"formula":"ROE > 15","universe":["000001"]}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404 when evaluator unwired, got %d", w.Code)
	}
}
