// Package middleware exposes the auth wiring shared by the HTTP and
// gRPC surfaces. Both paths reuse the same TokenVerifier so the
// signing key is configured exactly once.
package middleware

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	userErr "github.com/agoXQ/QuantLab/app/user/domain/errors"
	"github.com/agoXQ/QuantLab/app/user/infrastructure/token"
	"github.com/agoXQ/QuantLab/pkg/errors"
	"github.com/agoXQ/QuantLab/pkg/response"
)

// ContextKey is the typed context key used to store the authenticated
// user id; downstream handlers can read it via UserIDFromContext.
type ContextKey string

const (
	ctxKeyUserID ContextKey = "user_id"

	// MetadataKeyUserID lets non-auth callers (legacy clients,
	// internal tooling) inject the caller id without a JWT. Production
	// can disable the fallback by stripping the header at the edge.
	MetadataKeyUserID = "x-user-id"
)

// TokenVerifier abstracts the issuer.Verify dependency so tests can
// supply a stub. The middleware always asks for KindAccess so a
// stolen refresh token alone cannot drive a protected API.
type TokenVerifier interface {
	Verify(token string, kind token.TokenKind) (int64, error)
}

// GinAuth returns a gin middleware that resolves the caller id from
// either the Authorization: Bearer header or the metadata fallback
// header. Required = true short-circuits unauthenticated requests
// with a 401 envelope; required = false still resolves the caller id
// when present so handlers can adjust without forcing every route
// behind auth.
func GinAuth(verifier TokenVerifier, required bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := resolveUserIDFromGin(c, verifier)
		if err != nil {
			if required {
				response.Error(c, http.StatusUnauthorized, errors.New(errors.ErrUnauthorized.Code, err.Error()))
				c.Abort()
				return
			}
		}
		if userID > 0 {
			c.Set(string(ctxKeyUserID), userID)
			ctx := context.WithValue(c.Request.Context(), ctxKeyUserID, userID)
			c.Request = c.Request.WithContext(ctx)
		} else if required {
			response.Error(c, http.StatusUnauthorized, errors.ErrUnauthorized)
			c.Abort()
			return
		}
		c.Next()
	}
}

// GRPCAuth returns a unary interceptor that resolves the caller id
// from gRPC metadata and stamps it onto the context so logic adapters
// can read it via UserIDFromContext. The interceptor never short-
// circuits a call: gRPC handlers decide which methods require auth by
// inspecting the resolved id, mirroring the looser shape today's
// proto contracts already imply.
func GRPCAuth(verifier TokenVerifier) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if userID, err := resolveUserIDFromMetadata(ctx, verifier); err == nil && userID > 0 {
			ctx = context.WithValue(ctx, ctxKeyUserID, userID)
		}
		return handler(ctx, req)
	}
}

// UserIDFromContext returns the authenticated user id stamped onto
// the context by the middleware / interceptor; 0 means anonymous.
func UserIDFromContext(ctx context.Context) int64 {
	if ctx == nil {
		return 0
	}
	if v, ok := ctx.Value(ctxKeyUserID).(int64); ok {
		return v
	}
	return 0
}

// UserIDFromGin is the gin-side counterpart used by route handlers
// that already hold a *gin.Context.
func UserIDFromGin(c *gin.Context) int64 {
	if v, ok := c.Get(string(ctxKeyUserID)); ok {
		if id, ok := v.(int64); ok {
			return id
		}
	}
	return UserIDFromContext(c.Request.Context())
}

func resolveUserIDFromGin(c *gin.Context, verifier TokenVerifier) (int64, error) {
	if header := strings.TrimSpace(c.GetHeader("Authorization")); header != "" {
		if userID, err := verifyBearer(verifier, header); err == nil {
			return userID, nil
		} else {
			return 0, err
		}
	}
	if raw := strings.TrimSpace(c.GetHeader("X-User-Id")); raw != "" {
		return parseUserHeader(raw)
	}
	return 0, nil
}

func resolveUserIDFromMetadata(ctx context.Context, verifier TokenVerifier) (int64, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return 0, nil
	}
	if values := md.Get("authorization"); len(values) > 0 {
		if userID, err := verifyBearer(verifier, values[0]); err == nil {
			return userID, nil
		}
	}
	if values := md.Get(MetadataKeyUserID); len(values) > 0 {
		return parseUserHeader(values[0])
	}
	return 0, nil
}

func verifyBearer(verifier TokenVerifier, header string) (int64, error) {
	if verifier == nil {
		return 0, userErr.ErrTokenInvalid
	}
	const prefix = "Bearer "
	if len(header) <= len(prefix) || !strings.EqualFold(header[:len(prefix)], prefix) {
		return 0, userErr.ErrTokenInvalid
	}
	tokenStr := strings.TrimSpace(header[len(prefix):])
	if tokenStr == "" {
		return 0, userErr.ErrTokenInvalid
	}
	return verifier.Verify(tokenStr, token.KindAccess)
}

func parseUserHeader(raw string) (int64, error) {
	id, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
	if err != nil || id <= 0 {
		return 0, userErr.ErrTokenInvalid
	}
	return id, nil
}
