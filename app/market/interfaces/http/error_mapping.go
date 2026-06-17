package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	domainErr "github.com/agoXQ/QuantLab/app/market/domain/errors"
	"github.com/agoXQ/QuantLab/pkg/errors"
	"github.com/agoXQ/QuantLab/pkg/response"
)

// asAppError converts a domain MarketError into the platform AppError so the
// HTTP layer can rely on a single response shape.
func asAppError(err error) *errors.AppError {
	if err == nil {
		return nil
	}
	if appErr, ok := err.(*errors.AppError); ok {
		return appErr
	}
	if mErr, ok := err.(*domainErr.MarketError); ok {
		return errors.New(mErr.Code, mErr.Message)
	}
	return nil
}

// writeMappedErr is a helper used by handlers that prefer explicit mapping.
func writeMappedErr(c *gin.Context, err error) {
	if appErr := asAppError(err); appErr != nil {
		response.Error(c, statusForCode(appErr.Code), appErr)
		return
	}
	response.Error(c, http.StatusInternalServerError, errors.New(errors.ErrInternal.Code, err.Error()))
}
