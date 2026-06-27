// Package handler translates REST requests into gRPC calls against the
// backend services. One file per resource keeps the mapping surface
// scannable; this file holds the shared helpers every resource file
// reuses.
package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/agoXQ/QuantLab/app/gateway/internal/svc"
	"github.com/agoXQ/QuantLab/pkg/errors"
	"github.com/agoXQ/QuantLab/pkg/response"
)

// Handler is the shared base; every resource handler embeds it so the
// service clients and helpers are one field away.
type Handler struct {
	svc *svc.ServiceContext
}

func NewHandler(s *svc.ServiceContext) *Handler {
	return &Handler{svc: s}
}

// parseID reads an int64 path param, writing a 400 envelope on failure.
func parseID(c *gin.Context, param string) (int64, bool) {
	raw := c.Param(param)
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		response.Error(c, http.StatusBadRequest, errors.ErrInvalidParam)
		return 0, false
	}
	return id, true
}

// queryInt64 reads an optional int64 query param; returns 0 when absent.
func queryInt64(c *gin.Context, key string) int64 {
	v := c.Query(key)
	if v == "" {
		return 0
	}
	n, _ := strconv.ParseInt(v, 10, 64)
	return n
}

// queryLimit returns a cursor limit, defaulting to 20, capped at 100.
func queryLimit(c *gin.Context) int32 {
	return queryLimitWithMax(c, 20, 100)
}

func queryLimitWithMax(c *gin.Context, def, max int64) int32 {
	n := queryInt64(c, "limit")
	if n <= 0 {
		return int32(def)
	}
	if n > max {
		return int32(max)
	}
	return int32(n)
}

// queryCursor returns the cursor query param (empty when absent).
func queryCursor(c *gin.Context) string {
	return c.Query("cursor")
}

// grpcErr maps a gRPC status to the platform's AppError envelope. The
// mapping covers the common code families so a NotFound from any
// service surfaces as the platform's 4xxxx range, an InvalidArgument
// as 1xxxx, and so on; anything unmapped falls back to Internal.
func grpcErr(c *gin.Context, err error) {
	if err == nil {
		return
	}
	st, ok := status.FromError(err)
	if !ok {
		response.Error(c, http.StatusInternalServerError, errors.ErrInternal)
		return
	}
	var appErr *errors.AppError
	httpStatus := http.StatusInternalServerError
	switch st.Code() {
	case codes.OK:
		return
	case codes.NotFound:
		appErr = errors.ErrNotFound
		httpStatus = http.StatusNotFound
	case codes.AlreadyExists:
		appErr = errors.ErrDuplicateOperation
		httpStatus = http.StatusConflict
	case codes.InvalidArgument, codes.FailedPrecondition, codes.OutOfRange:
		appErr = errors.ErrInvalidParam
		httpStatus = http.StatusBadRequest
	case codes.Unauthenticated:
		appErr = errors.ErrUnauthorized
		httpStatus = http.StatusUnauthorized
	case codes.PermissionDenied:
		appErr = errors.ErrForbidden
		httpStatus = http.StatusForbidden
	case codes.DeadlineExceeded:
		appErr = errors.ErrTimeout
		httpStatus = http.StatusGatewayTimeout
	case codes.Unavailable:
		appErr = errors.New(50006, "service unavailable")
		httpStatus = http.StatusServiceUnavailable
	case codes.ResourceExhausted:
		appErr = errors.ErrQuotaExceeded
		httpStatus = http.StatusTooManyRequests
	default:
		appErr = errors.ErrInternal
	}
	if msg := st.Message(); msg != "" {
		appErr = errors.New(appErr.Code, msg)
	}
	response.Error(c, httpStatus, appErr)
}
