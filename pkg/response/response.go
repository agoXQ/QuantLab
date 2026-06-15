// Package response provides the unified API response format for all QuantLab services.
//
// Standard response envelope:
//
//	{
//	  "code": 0,
//	  "message": "success",
//	  "data": {},
//	  "meta": {},
//	  "request_id": "req_xxx"
//	}
package response

import (
	"net/http"

	"github.com/agoXQ/QuantLab/pkg/errors"

	"github.com/gin-gonic/gin"
)

// Response is the unified API response envelope.
type Response struct {
	Code      int         `json:"code"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	Meta      interface{} `json:"meta,omitempty"`
	RequestID string      `json:"request_id,omitempty"`
}

// PageMeta contains cursor-based pagination metadata.
type PageMeta struct {
	NextCursor string `json:"next_cursor,omitempty"`
	HasMore    bool   `json:"has_more"`
}

// OK sends a successful response with data.
func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:      0,
		Message:   "success",
		Data:      data,
		RequestID: getRequestID(c),
	})
}

// Created sends a 201 response for resource creation.
func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, Response{
		Code:      0,
		Message:   "success",
		Data:      data,
		RequestID: getRequestID(c),
	})
}

// OKWithMeta sends a successful response with pagination metadata.
func OKWithMeta(c *gin.Context, data interface{}, meta interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:      0,
		Message:   "success",
		Data:      data,
		Meta:      meta,
		RequestID: getRequestID(c),
	})
}

// NoContent sends a 204 response.
func NoContent(c *gin.Context) {
	c.JSON(http.StatusNoContent, Response{
		Code:      0,
		Message:   "success",
		RequestID: getRequestID(c),
	})
}

// Error sends an error response with the given HTTP status and error.
func Error(c *gin.Context, httpStatus int, err *errors.AppError) {
	c.JSON(httpStatus, Response{
		Code:      err.Code,
		Message:   err.Message,
		RequestID: getRequestID(c),
	})
}

func getRequestID(c *gin.Context) string {
	if id, exists := c.Get("request_id"); exists {
		return id.(string)
	}
	return ""
}
