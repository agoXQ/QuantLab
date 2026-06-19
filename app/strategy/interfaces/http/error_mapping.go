package http

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	stratErr "github.com/agoXQ/QuantLab/app/strategy/domain/errors"
	"github.com/agoXQ/QuantLab/pkg/errors"
	"github.com/agoXQ/QuantLab/pkg/response"
)

// asAppError translates a Strategy domain error to the platform AppError
// so the HTTP layer keeps a single response shape across services.
func asAppError(err error) *errors.AppError {
	if err == nil {
		return nil
	}
	if appErr, ok := err.(*errors.AppError); ok {
		return appErr
	}
	if sErr, ok := err.(*stratErr.StrategyError); ok {
		return errors.New(sErr.Code, sErr.Message)
	}
	return nil
}

func writeMappedErr(c *gin.Context, err error) {
	if appErr := asAppError(err); appErr != nil {
		response.Error(c, statusForCode(appErr.Code), appErr)
		return
	}
	response.Error(c, http.StatusInternalServerError, errors.New(errors.ErrInternal.Code, err.Error()))
}

// statusForCode maps the platform error-code prefix to the HTTP status.
// Mirrors Backtest verbatim so the platform rule book stays aligned.
func statusForCode(code int) int {
	switch {
	case code >= 10000 && code < 20000:
		return http.StatusBadRequest
	case code >= 20000 && code < 30000:
		return http.StatusConflict
	case code >= 30000 && code < 40000:
		return http.StatusForbidden
	case code >= 40000 && code < 50000:
		return http.StatusNotFound
	case code >= 50000 && code < 60000:
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}

// parseID parses the named path parameter into int64 and writes a 400
// on failure.
func parseID(c *gin.Context, name string) (int64, bool) {
	raw := c.Param(name)
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		response.Error(c, http.StatusBadRequest, errors.New(errors.ErrInvalidParam.Code, "invalid "+name))
		return 0, false
	}
	return id, true
}

func parseInt64(s string) int64 {
	if s == "" {
		return 0
	}
	v, _ := strconv.ParseInt(s, 10, 64)
	return v
}
