// Package http exposes the Notification Service application service
// over a thin HTTP surface aligned with the QuantLab API Design
// Standard. Mirrors the User / Strategy / Backtest shape: one Service
// dependency, one RegisterRoutes that takes the api group, and an
// error mapping helper that turns domain errors into the platform's
// AppError envelope.
package http

import (
	"github.com/gin-gonic/gin"

	appNotif "github.com/agoXQ/QuantLab/app/notification/application/notification"
)

// Handler exposes the Notification Service over HTTP.
type Handler struct {
	svc appNotif.Service
}

// NewHandler returns a Handler bound to the given application service.
func NewHandler(svc appNotif.Service) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes mounts every route under the supplied group, expected
// to be /api/v1/notifications.
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("", h.List)
	rg.GET("unread-count", h.UnreadCount)
	rg.POST(":id/read", h.MarkRead)
	rg.POST("read-all", h.MarkAllRead)
	rg.DELETE(":id", h.Delete)

	rg.GET("preferences", h.GetPreferences)
	rg.PUT("preferences", h.UpdatePreferences)

	rg.GET("subscriptions", h.ListSubscriptions)
	rg.POST("subscriptions", h.CreateSubscription)
	rg.DELETE("subscriptions/:id", h.CancelSubscription)
}

// updatePreferencesRequest is the body shape for PUT preferences.
type updatePreferencesRequest struct {
	InAppEnabled   bool `json:"in_app_enabled"`
	EmailEnabled   bool `json:"email_enabled"`
	WebhookEnabled bool `json:"webhook_enabled"`
	PushEnabled    bool `json:"push_enabled"`
}

// createSubscriptionRequest is the body shape for POST subscriptions.
type createSubscriptionRequest struct {
	ObjectType string `json:"object_type" binding:"required"`
	ObjectID   int64  `json:"object_id" binding:"required"`
	UserID     int64  `json:"user_id"`
}
