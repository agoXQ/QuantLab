package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	commonpb "github.com/agoXQ/QuantLab/api/common/v1"
	notificationpb "github.com/agoXQ/QuantLab/app/notification/pb"
	"github.com/agoXQ/QuantLab/pkg/errors"
	"github.com/agoXQ/QuantLab/pkg/response"
)

// registerNotifications mounts /api/v1/notifications.
func (h *Handler) registerNotifications(rg *gin.RouterGroup) {
	n := rg.Group("/notifications")
	n.GET("", h.notifList)
	n.GET("/unread-count", h.notifUnreadCount)
	n.POST("/:id/read", h.notifMarkRead)
	n.POST("/read-all", h.notifMarkAllRead)
	n.DELETE("/:id", h.notifDelete)
	n.GET("/preferences", h.notifGetPreferences)
	n.PUT("/preferences", h.notifUpdatePreferences)
	n.POST("/subscriptions", h.notifCreateSubscription)
	n.DELETE("/subscriptions/:id", h.notifCancelSubscription)
	n.GET("/subscriptions", h.notifListSubscriptions)
}

func (h *Handler) notifList(c *gin.Context) {
	out, err := h.svc.Notification.ListNotifications(c.Request.Context(), &notificationpb.ListNotificationsRequest{
		Cursor: &commonpb.Cursor{NextCursor: queryCursor(c)}, Limit: queryLimit(c),
	})
	if err != nil {
		grpcErr(c, err)
		return
	}
	response.OKWithMeta(c, gin.H{"items": out.Notifications}, out.Cursor)
}

func (h *Handler) notifUnreadCount(c *gin.Context) {
	out, err := h.svc.Notification.GetUnreadCount(c.Request.Context(), &notificationpb.GetUnreadCountRequest{})
	if err != nil {
		grpcErr(c, err)
		return
	}
	response.OK(c, gin.H{"count": out.Count})
}

func (h *Handler) notifMarkRead(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	if _, err := h.svc.Notification.MarkRead(c.Request.Context(), &notificationpb.MarkReadRequest{NotificationId: id}); err != nil {
		grpcErr(c, err)
		return
	}
	response.OK(c, gin.H{})
}

func (h *Handler) notifMarkAllRead(c *gin.Context) {
	if _, err := h.svc.Notification.MarkAllRead(c.Request.Context(), &notificationpb.MarkAllReadRequest{}); err != nil {
		grpcErr(c, err)
		return
	}
	response.OK(c, gin.H{})
}

func (h *Handler) notifDelete(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	if _, err := h.svc.Notification.DeleteNotification(c.Request.Context(), &notificationpb.DeleteNotificationRequest{NotificationId: id}); err != nil {
		grpcErr(c, err)
		return
	}
	response.NoContent(c)
}

func (h *Handler) notifGetPreferences(c *gin.Context) {
	out, err := h.svc.Notification.GetPreferences(c.Request.Context(), &notificationpb.GetPreferencesRequest{})
	if err != nil {
		grpcErr(c, err)
		return
	}
	response.OK(c, gin.H{"preferences": out.Preferences})
}

func (h *Handler) notifUpdatePreferences(c *gin.Context) {
	var req struct {
		InAppEnabled   bool `json:"in_app_enabled"`
		EmailEnabled   bool `json:"email_enabled"`
		WebhookEnabled bool `json:"webhook_enabled"`
		PushEnabled    bool `json:"push_enabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, errors.ErrInvalidParam)
		return
	}
	if _, err := h.svc.Notification.UpdatePreferences(c.Request.Context(), &notificationpb.UpdatePreferencesRequest{
		InAppEnabled: req.InAppEnabled, EmailEnabled: req.EmailEnabled,
		WebhookEnabled: req.WebhookEnabled, PushEnabled: req.PushEnabled,
	}); err != nil {
		grpcErr(c, err)
		return
	}
	response.OK(c, gin.H{})
}

func (h *Handler) notifCreateSubscription(c *gin.Context) {
	var req struct {
		ObjectType string `json:"object_type" binding:"required"`
		ObjectID   int64  `json:"object_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, errors.ErrInvalidParam)
		return
	}
	out, err := h.svc.Notification.CreateSubscription(c.Request.Context(), &notificationpb.CreateSubscriptionRequest{
		ObjectType: req.ObjectType, ObjectId: req.ObjectID,
	})
	if err != nil {
		grpcErr(c, err)
		return
	}
	response.Created(c, gin.H{"subscription_id": out.SubscriptionId})
}

func (h *Handler) notifCancelSubscription(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	if _, err := h.svc.Notification.CancelSubscription(c.Request.Context(), &notificationpb.CancelSubscriptionRequest{SubscriptionId: id}); err != nil {
		grpcErr(c, err)
		return
	}
	response.NoContent(c)
}

func (h *Handler) notifListSubscriptions(c *gin.Context) {
	out, err := h.svc.Notification.ListSubscriptions(c.Request.Context(), &notificationpb.ListSubscriptionsRequest{
		Cursor: &commonpb.Cursor{NextCursor: queryCursor(c)}, Limit: queryLimit(c),
	})
	if err != nil {
		grpcErr(c, err)
		return
	}
	response.OKWithMeta(c, gin.H{"items": out.Subscriptions}, out.Cursor)
}

var _ = http.StatusOK
