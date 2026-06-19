package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	httpHandler "github.com/agoXQ/QuantLab/app/strategy/interfaces/http"
)

func setupHTTPRouter(t *testing.T, fx *fixture) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	apiGroup := router.Group("/api/v1/strategies")
	httpHandler.NewHandler(fx.svc).RegisterRoutes(apiGroup)
	return router
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

// TestHTTP_CRUDAndLifecycle is the HTTP-side counterpart to the in-process
// TestEndToEnd: it walks the same lifecycle end to end through the gin
// router so we know the JSON envelope and routing wiring stay in sync
// with the application service.
func TestHTTP_CRUDAndLifecycle(t *testing.T) {
	fx := newFixture(t)
	router := setupHTTPRouter(t, fx)

	// Create
	w := executeJSON(router, "POST", "/api/v1/strategies", map[string]any{
		"author_id":  7,
		"title":      "Mean reversion",
		"tags":       []string{"meanrev"},
		"visibility": "PRIVATE",
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d body=%s", w.Code, w.Body.String())
	}
	var createEnv struct {
		Data struct {
			Strategy struct {
				ID int64 `json:"id"`
			} `json:"strategy"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &createEnv); err != nil {
		t.Fatalf("decode create: %v", err)
	}
	id := createEnv.Data.Strategy.ID
	if id == 0 {
		t.Fatalf("expected non-zero id, body=%s", w.Body.String())
	}

	// Get
	w = executeJSON(router, "GET", "/api/v1/strategies", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("list: expected 200, got %d", w.Code)
	}

	// CreateVersion
	w = executeJSON(router, "POST", "/api/v1/strategies/1/versions", map[string]any{
		"caller_id":    7,
		"formula_text": "ROE > 0",
		"change_log":   "initial",
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("create version: expected 201, got %d body=%s", w.Code, w.Body.String())
	}
	var verEnv struct {
		Data struct {
			Version struct {
				ID int64 `json:"id"`
			} `json:"version"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &verEnv); err != nil {
		t.Fatalf("decode version: %v", err)
	}
	if verEnv.Data.Version.ID == 0 {
		t.Fatalf("expected version id non-zero, body=%s", w.Body.String())
	}

	// Publish
	w = executeJSON(router, "POST", "/api/v1/strategies/1/publish", map[string]any{"caller_id": 7})
	if w.Code != http.StatusOK {
		t.Fatalf("publish: expected 200, got %d body=%s", w.Code, w.Body.String())
	}

	// Fork
	w = executeJSON(router, "POST", "/api/v1/strategies/1/fork", map[string]any{"caller_id": 9})
	if w.Code != http.StatusCreated {
		t.Fatalf("fork: expected 201, got %d body=%s", w.Code, w.Body.String())
	}

	// Archive
	w = executeJSON(router, "POST", "/api/v1/strategies/1/archive", map[string]any{"caller_id": 7})
	if w.Code != http.StatusOK {
		t.Fatalf("archive: expected 200, got %d body=%s", w.Code, w.Body.String())
	}

	// Get returns archived state
	w = executeJSON(router, "GET", "/api/v1/strategies/1", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("get archived: expected 200, got %d", w.Code)
	}
}

// TestHTTP_NotFoundMappedTo404 makes sure unknown ids surface as 404
// rather than the default 500 the platform reserves for system errors.
func TestHTTP_NotFoundMappedTo404(t *testing.T) {
	fx := newFixture(t)
	router := setupHTTPRouter(t, fx)

	w := executeJSON(router, "GET", "/api/v1/strategies/999", nil)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for unknown id, got %d body=%s", w.Code, w.Body.String())
	}
}
