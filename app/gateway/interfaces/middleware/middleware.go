// Package middleware wires the cross-cutting concerns the gateway owns
// centrally: request id, CORS, structured logging, panic recovery, and
// JWT auth. Auth is resolved once here and the resolved caller id is
// stamped onto the request context so handlers can forward it as gRPC
// metadata (x-user-id) to downstream services.
package middleware

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"google.golang.org/grpc/metadata"

	"github.com/agoXQ/QuantLab/app/user/infrastructure/token"
	"github.com/agoXQ/QuantLab/pkg/errors"
	"github.com/agoXQ/QuantLab/pkg/response"
)

// ctxKeyCaller is the gin-context key for the authenticated caller id.
const ctxKeyCaller = "gateway_caller_id"

// RequestID injects a request id (from X-Request-Id or freshly minted)
// onto the gin context and response header so every log line and the
// response envelope carry the same trace id.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader("X-Request-Id")
		if id == "" {
			id = uuid.NewString()
		}
		c.Set("request_id", id)
		c.Header("X-Request-Id", id)
		c.Next()
	}
}

// Logging records method, path, status, and latency for every request.
func Logging() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		log.Printf("[gateway] %s %s %d %v", c.Request.Method, c.Request.URL.Path, c.Writer.Status(), time.Since(start))
	}
}

// Recovery turns panics into 500 envelopes so a single handler bug
// never crashes the whole gateway.
func Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, rcv any) {
		log.Printf("[gateway] panic: %v path=%s", rcv, c.Request.URL.Path)
		response.Error(c, http.StatusInternalServerError, errors.ErrInternal)
		c.Abort()
	})
}

// CORS allows the frontend (different origin in dev) to call the
// gateway directly. Tighten the origin list in production.
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-Id, Idempotency-Key")
		c.Header("Access-Control-Max-Age", "86400")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

// Auth returns a gin middleware that verifies the Bearer access token
// via the shared JWT verifier. required=false lets public routes still
// resolve the caller id when a token is present (so a logged-in user
// browsing public strategy pages is identifiable) without rejecting
// anonymous traffic.
func Auth(verifier *token.JWTIssuer, required bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := resolveCaller(c, verifier)
		if userID > 0 {
			c.Set(ctxKeyCaller, userID)
		} else if required {
			response.Error(c, http.StatusUnauthorized, errors.ErrUnauthorized)
			c.Abort()
			return
		}
		c.Next()
	}
}

// CallerID returns the authenticated caller id stamped by Auth, or 0
// for anonymous requests.
func CallerID(c *gin.Context) int64 {
	if v, ok := c.Get(ctxKeyCaller); ok {
		if id, ok := v.(int64); ok {
			return id
		}
	}
	return 0
}

// Outgoing returns a context derived from the gin request context that
// carries the caller id as gRPC metadata (x-user-id) so downstream
// services' GRPCAuth interceptor can read it. Call this at the top of
// every handler before invoking a gRPC client.
func Outgoing(c *gin.Context) context.Context {
	ctx := c.Request.Context()
	if id := CallerID(c); id > 0 {
		ctx = metadata.AppendToOutgoingContext(ctx, "x-user-id", strconv.FormatInt(id, 10))
	}
	return ctx
}

func resolveCaller(c *gin.Context, verifier *token.JWTIssuer) int64 {
	if verifier == nil {
		return 0
	}
	header := strings.TrimSpace(c.GetHeader("Authorization"))
	if header == "" {
		return 0
	}
	const prefix = "Bearer "
	if len(header) <= len(prefix) || !strings.EqualFold(header[:len(prefix)], prefix) {
		return 0
	}
	tok := strings.TrimSpace(header[len(prefix):])
	if tok == "" {
		return 0
	}
	id, err := verifier.Verify(tok, token.KindAccess)
	if err != nil {
		return 0
	}
	return id
}
