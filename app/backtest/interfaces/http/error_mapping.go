package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	bterr "github.com/agoXQ/QuantLab/app/backtest/domain/errors"
	"github.com/agoXQ/QuantLab/pkg/errors"
	"github.com/agoXQ/QuantLab/pkg/response"
)

// asAppError translates a Backtest domain error to the platform AppError so
// the HTTP layer keeps a single response shape across services.
func asAppError(err error) *errors.AppError {
	if err == nil {
		return nil
	}
	if appErr, ok := err.(*errors.AppError); ok {
		return appErr
	}
	if bErr, ok := err.(*bterr.BacktestError); ok {
		return errors.New(bErr.Code, bErr.Message)
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

// statusForCode maps the platform error code prefix to the appropriate HTTP
// status. It mirrors the Market Data version verbatim to keep the rule book
// in lockstep across services.
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
