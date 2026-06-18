package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	httpHandler "github.com/agoXQ/QuantLab/app/backtest/interfaces/http"
)

func setupHTTPRouter(t *testing.T, fx *fixture) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	apiGroup := router.Group("/api/v1/backtests")
	httpHandler.NewHandler(fx.svc).RegisterRoutes(apiGroup)
	return router
}

// seedSimpleUniverse seeds one stock that always passes ROE>0 across a tiny
// trading window so the synchronous run finishes in well under a second.
func seedSimpleUniverse(fx *fixture) (start, end time.Time) {
	start = time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC)
	calendar := daily(start, 6)
	bars := linearBars(calendar, 30, 0.2)
	fx.provider.SetBars("000001", bars)
	fx.provider.SetCalendar(calendar)
	fx.formula.SetBars("000001", formulaBars(bars))
	fx.formula.SetFinancials("000001", map[string]float64{"ROE": 30})
	return calendar[0], calendar[len(calendar)-1]
}

func executeJSON(router *gin.Engine, method, path string, body any) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	var reader *bytes.Reader
	if body != nil {
		buf, _ := json.Marshal(body)
		reader = bytes.NewReader(buf)
	} else {
		reader = bytes.NewReader(nil)
	}
	req, _ := http.NewRequest(method, path, reader)
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	return w
}

func TestHTTP_CreateAndRun(t *testing.T) {
	fx := newFixture(t)
	start, end := seedSimpleUniverse(fx)
	router := setupHTTPRouter(t, fx)

	body := map[string]any{
		"formula":         "ROE > 0",
		"universe":        []string{"000001"},
		"initial_capital": 100000,
		"start_date":      start.Format("2006-01-02"),
		"end_date":        end.Format("2006-01-02"),
		"config": map[string]any{
			"rebalance_frequency": "daily",
			"max_position_count":  1,
		},
	}
	w := executeJSON(router, "POST", "/api/v1/backtests?run=true", body)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d body=%s", w.Code, w.Body.String())
	}

	var env struct {
		Code int `json:"code"`
		Data struct {
			Job struct {
				ID     int64  `json:"id"`
				Status string `json:"status"`
			} `json:"job"`
			Report struct {
				TradeCount int `json:"trade_count"`
			} `json:"report"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &env); err != nil {
		t.Fatalf("decode: %v body=%s", err, w.Body.String())
	}
	if env.Data.Job.Status != "COMPLETED" {
		t.Errorf("expected COMPLETED, got %s", env.Data.Job.Status)
	}
	if env.Data.Job.ID == 0 {
		t.Errorf("expected non-zero job id")
	}

	// Follow-up reads.
	w = executeJSON(router, "GET", "/api/v1/backtests", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("list: expected 200, got %d", w.Code)
	}
	w = executeJSON(router, "GET", "/api/v1/backtests/1/report", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("report: expected 200, got %d body=%s", w.Code, w.Body.String())
	}
	w = executeJSON(router, "GET", "/api/v1/backtests/1/trades", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("trades: expected 200, got %d", w.Code)
	}
	w = executeJSON(router, "GET", "/api/v1/backtests/1/positions", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("positions: expected 200, got %d", w.Code)
	}
}

func TestHTTP_CreateValidationErrors(t *testing.T) {
	fx := newFixture(t)
	router := setupHTTPRouter(t, fx)

	cases := []struct {
		name string
		body map[string]any
	}{
		{name: "missing formula", body: map[string]any{
			"universe": []string{"000001"}, "initial_capital": 1.0,
			"start_date": "2024-01-01", "end_date": "2024-01-05",
		}},
		{name: "bad date", body: map[string]any{
			"formula": "ROE > 0", "universe": []string{"000001"}, "initial_capital": 1.0,
			"start_date": "not-a-date", "end_date": "2024-01-05",
		}},
		{name: "bad rebalance", body: map[string]any{
			"formula": "ROE > 0", "universe": []string{"000001"}, "initial_capital": 1.0,
			"start_date": "2024-01-01", "end_date": "2024-01-05",
			"config": map[string]any{"rebalance_frequency": "hourly"},
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w := executeJSON(router, "POST", "/api/v1/backtests", tc.body)
			if w.Code != http.StatusBadRequest {
				t.Errorf("expected 400, got %d body=%s", w.Code, w.Body.String())
			}
		})
	}
}

func TestHTTP_GetNotFound(t *testing.T) {
	fx := newFixture(t)
	router := setupHTTPRouter(t, fx)
	w := executeJSON(router, "GET", "/api/v1/backtests/999", nil)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d body=%s", w.Code, w.Body.String())
	}
}
