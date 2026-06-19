package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	httpHandler "github.com/agoXQ/QuantLab/app/user/interfaces/http"
)

func setupHTTPRouter(t *testing.T, fx *fixture) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	apiGroup := router.Group("/api/v1/users")
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

// TestHTTP_RegisterLoginFollow drives the canonical sign-up + follow
// path through the gin router so we know the JSON envelope and routing
// wiring stay in sync with the application service.
func TestHTTP_RegisterLoginFollow(t *testing.T) {
	fx := newFixture(t)
	router := setupHTTPRouter(t, fx)

	// Register Alice
	w := executeJSON(router, "POST", "/api/v1/users/register", map[string]any{
		"username": "alice",
		"email":    "alice@example.com",
		"password": "hunter2hunter",
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("register alice: expected 201, got %d body=%s", w.Code, w.Body.String())
	}
	var aliceEnv struct {
		Data struct {
			User struct {
				ID int64 `json:"id"`
			} `json:"user"`
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &aliceEnv); err != nil {
		t.Fatalf("decode alice: %v", err)
	}
	if aliceEnv.Data.User.ID == 0 || aliceEnv.Data.AccessToken == "" {
		t.Fatalf("expected id+token, body=%s", w.Body.String())
	}
	aliceID := aliceEnv.Data.User.ID

	// Register Bob
	w = executeJSON(router, "POST", "/api/v1/users/register", map[string]any{
		"username": "bob",
		"email":    "bob@example.com",
		"password": "hunter2hunter",
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("register bob: expected 201, got %d body=%s", w.Code, w.Body.String())
	}
	var bobEnv struct {
		Data struct {
			User struct {
				ID int64 `json:"id"`
			} `json:"user"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &bobEnv); err != nil {
		t.Fatalf("decode bob: %v", err)
	}
	bobID := bobEnv.Data.User.ID

	// Login Alice
	w = executeJSON(router, "POST", "/api/v1/users/login", map[string]any{
		"email":    "alice@example.com",
		"password": "hunter2hunter",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("login: expected 200, got %d body=%s", w.Code, w.Body.String())
	}

	// UpdateProfile
	w = executeJSON(router, "PUT", "/api/v1/users/"+itoa(aliceID)+"/profile", map[string]any{
		"bio":      "long-term holder",
		"location": "Tokyo",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("update profile: expected 200, got %d body=%s", w.Code, w.Body.String())
	}

	// Follow
	w = executeJSON(router, "POST", "/api/v1/users/"+itoa(bobID)+"/follow", map[string]any{
		"follower_id": aliceID,
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("follow: expected 201, got %d body=%s", w.Code, w.Body.String())
	}

	// Profile shows the count
	w = executeJSON(router, "GET", "/api/v1/users/"+itoa(bobID)+"/profile", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("profile: expected 200, got %d body=%s", w.Code, w.Body.String())
	}
	var profEnv struct {
		Data struct {
			FollowerCount int64 `json:"follower_count"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &profEnv); err != nil {
		t.Fatalf("decode profile: %v", err)
	}
	if profEnv.Data.FollowerCount != 1 {
		t.Fatalf("expected follower_count=1, got %d body=%s", profEnv.Data.FollowerCount, w.Body.String())
	}

	// Followers list
	w = executeJSON(router, "GET", "/api/v1/users/"+itoa(bobID)+"/followers", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("followers: expected 200, got %d body=%s", w.Code, w.Body.String())
	}

	// Unfollow
	w = executeJSON(router, "DELETE", "/api/v1/users/"+itoa(bobID)+"/follow?follower_id="+itoa(aliceID), nil)
	if w.Code != http.StatusNoContent && w.Code != http.StatusOK {
		t.Fatalf("unfollow: expected 204/200, got %d body=%s", w.Code, w.Body.String())
	}
}

// TestHTTP_DuplicateRegisterConflicts confirms duplicate registration
// surfaces the platform's 409 Conflict.
func TestHTTP_DuplicateRegisterConflicts(t *testing.T) {
	fx := newFixture(t)
	router := setupHTTPRouter(t, fx)

	body := map[string]any{
		"username": "alice",
		"email":    "alice@example.com",
		"password": "hunter2hunter",
	}
	w := executeJSON(router, "POST", "/api/v1/users/register", body)
	if w.Code != http.StatusCreated {
		t.Fatalf("first register: expected 201, got %d", w.Code)
	}
	w = executeJSON(router, "POST", "/api/v1/users/register", body)
	if w.Code != http.StatusConflict {
		t.Fatalf("duplicate register: expected 409, got %d body=%s", w.Code, w.Body.String())
	}
}

func itoa(v int64) string {
	if v == 0 {
		return "0"
	}
	negative := v < 0
	if negative {
		v = -v
	}
	const digits = "0123456789"
	buf := make([]byte, 0, 20)
	for v > 0 {
		buf = append([]byte{digits[v%10]}, buf...)
		v /= 10
	}
	if negative {
		buf = append([]byte{'-'}, buf...)
	}
	return string(buf)
}
