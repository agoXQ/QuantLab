package tests

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/metadata"

	httpHandler "github.com/agoXQ/QuantLab/app/user/interfaces/http"
	authmiddleware "github.com/agoXQ/QuantLab/app/user/interfaces/middleware"
	"github.com/agoXQ/QuantLab/app/user/infrastructure/token"
)

// TestAuthMiddleware_BearerHappy / 401 / ForbiddenSelfEdit confirm the
// gin middleware lifts the JWT subject onto the gin context and that
// the handler refuses a mismatched edit.
func TestAuthMiddleware_BearerAndSelfEdit(t *testing.T) {
	fx := newFixture(t)
	issuer := token.MustNewJWTIssuer(token.Config{
		Secret:    "test-signing-key-must-be-long-enough-32-bytes",
		Issuer:    "quantlab.user",
		AccessTTL: 5 * time.Minute,
	})

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(authmiddleware.GinAuth(issuer, false))
	apiGroup := router.Group("/api/v1/users")
	httpHandler.NewHandler(fx.svc).RegisterRoutes(apiGroup)

	id := registerHelper(t, fx, "alice", "alice@example.com")
	pair, err := issuer.Issue(id)
	if err != nil {
		t.Fatalf("issue token: %v", err)
	}

	// Self-edit with bearer token should succeed.
	req, _ := http.NewRequest(http.MethodPut, "/api/v1/users/"+itoa(id)+"/profile", strings.NewReader(`{"bio":"new bio"}`))
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("self edit: expected 200, got %d body=%s", w.Code, w.Body.String())
	}

	// Editing someone else's profile must 403.
	req2, _ := http.NewRequest(http.MethodPut, "/api/v1/users/"+itoa(id+999)+"/profile", strings.NewReader(`{"bio":"new"}`))
	req2.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	if w2.Code != http.StatusForbidden {
		t.Fatalf("foreign edit: expected 403, got %d body=%s", w2.Code, w2.Body.String())
	}
}

// TestAuthMiddleware_AccessRejectsRefreshToken guards the audience
// check on Verify: a refresh token must not be usable as an access
// token (and vice versa).
func TestAuthMiddleware_AccessRejectsRefreshToken(t *testing.T) {
	fx := newFixture(t)
	issuer := token.MustNewJWTIssuer(token.Config{
		Secret:    "test-signing-key-must-be-long-enough-32-bytes",
		Issuer:    "quantlab.user",
		AccessTTL: 5 * time.Minute,
	})

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(authmiddleware.GinAuth(issuer, true))
	apiGroup := router.Group("/api/v1/users")
	httpHandler.NewHandler(fx.svc).RegisterRoutes(apiGroup)

	id := registerHelper(t, fx, "alice", "alice@example.com")
	pair, err := issuer.Issue(id)
	if err != nil {
		t.Fatalf("issue: %v", err)
	}

	// refresh token must not unlock protected routes
	req, _ := http.NewRequest(http.MethodPut, "/api/v1/users/"+itoa(id)+"/profile", strings.NewReader(`{"bio":"x"}`))
	req.Header.Set("Authorization", "Bearer "+pair.RefreshToken)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 with refresh token, got %d body=%s", w.Code, w.Body.String())
	}
}

// TestGRPCAuthInterceptor_StampsUserID confirms the interceptor lifts
// the metadata subject onto the request context so logic adapters can
// read it via UserIDFromContext.
func TestGRPCAuthInterceptor_StampsUserID(t *testing.T) {
	issuer := token.MustNewJWTIssuer(token.Config{
		Secret:    "test-signing-key-must-be-long-enough-32-bytes",
		Issuer:    "quantlab.user",
		AccessTTL: 5 * time.Minute,
	})
	pair, err := issuer.Issue(42)
	if err != nil {
		t.Fatalf("issue: %v", err)
	}
	interceptor := authmiddleware.GRPCAuth(issuer)
	md := metadata.New(map[string]string{
		"authorization": "Bearer " + pair.AccessToken,
	})
	ctx := metadata.NewIncomingContext(context.Background(), md)
	called := false
	_, _ = interceptor(ctx, nil, nil, func(ctx context.Context, req any) (any, error) {
		called = true
		if got := authmiddleware.UserIDFromContext(ctx); got != 42 {
			t.Fatalf("expected user id 42, got %d", got)
		}
		return nil, nil
	})
	if !called {
		t.Fatal("handler not called")
	}

	// Fallback: x-user-id header without a JWT.
	mdFallback := metadata.New(map[string]string{
		"x-user-id": strconv.FormatInt(99, 10),
	})
	ctxFallback := metadata.NewIncomingContext(context.Background(), mdFallback)
	_, _ = interceptor(ctxFallback, nil, nil, func(ctx context.Context, req any) (any, error) {
		if got := authmiddleware.UserIDFromContext(ctx); got != 99 {
			t.Fatalf("expected fallback id 99, got %d", got)
		}
		return nil, nil
	})
}
