package http

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	appNotif "github.com/agoXQ/QuantLab/app/notification/application/notification"
	domNotif "github.com/agoXQ/QuantLab/app/notification/domain/notification"
	domPref "github.com/agoXQ/QuantLab/app/notification/domain/preference"
	domSub "github.com/agoXQ/QuantLab/app/notification/domain/subscription"
	"github.com/agoXQ/QuantLab/app/notification/domain/valueobject"
	"github.com/agoXQ/QuantLab/app/user/interfaces/middleware"
	"github.com/agoXQ/QuantLab/pkg/errors"
	"github.com/agoXQ/QuantLab/pkg/response"
)

// readUserID resolves the caller from the auth middleware first and
// falls back to query / body parameters so curl-driven smoke tests
// still work before every client signs JWTs.
func readUserID(c *gin.Context) int64 {
	if id := middleware.UserIDFromGin(c); id > 0 {
		return id
	}
	if v := parseInt64(c.Query("user_id")); v > 0 {
		return v
	}
	return 0
}

func ensureUser(c *gin.Context) (int64, bool) {
	id := readUserID(c)
	if id <= 0 {
		response.Error(c, http.StatusUnauthorized, errors.ErrUnauthorized)
		return 0, false
	}
	return id, true
}

func readPaging(c *gin.Context) (int, string) {
	limit := int(parseInt64(c.Query("limit")))
	cursor := c.Query("cursor")
	return limit, cursor
}

// List handles GET /api/v1/notifications.
func (h *Handler) List(c *gin.Context) {
	userID, ok := ensureUser(c)
	if !ok {
		return
	}
	limit, cursor := readPaging(c)
	out, err := h.svc.ListNotifications(c.Request.Context(), appNotif.ListNotificationsInput{
		UserID: userID,
		Limit:  limit,
		Cursor: cursor,
	})
	if err != nil {
		writeMappedErr(c, err)
		return
	}
	response.OK(c, gin.H{
		"items":       toNotificationViews(out.Items),
		"next_cursor": out.NextCursor,
	})
}

// UnreadCount handles GET /api/v1/notifications/unread-count.
func (h *Handler) UnreadCount(c *gin.Context) {
	userID, ok := ensureUser(c)
	if !ok {
		return
	}
	n, err := h.svc.GetUnreadCount(c.Request.Context(), userID)
	if err != nil {
		writeMappedErr(c, err)
		return
	}
	response.OK(c, gin.H{"count": n})
}

// MarkRead handles POST /api/v1/notifications/:id/read.
func (h *Handler) MarkRead(c *gin.Context) {
	userID, ok := ensureUser(c)
	if !ok {
		return
	}
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	if err := h.svc.MarkRead(c.Request.Context(), userID, id); err != nil {
		writeMappedErr(c, err)
		return
	}
	response.NoContent(c)
}

// MarkAllRead handles POST /api/v1/notifications/read-all.
func (h *Handler) MarkAllRead(c *gin.Context) {
	userID, ok := ensureUser(c)
	if !ok {
		return
	}
	n, err := h.svc.MarkAllRead(c.Request.Context(), userID)
	if err != nil {
		writeMappedErr(c, err)
		return
	}
	response.OK(c, gin.H{"affected": n})
}

// Delete handles DELETE /api/v1/notifications/:id.
func (h *Handler) Delete(c *gin.Context) {
	userID, ok := ensureUser(c)
	if !ok {
		return
	}
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	if err := h.svc.DeleteNotification(c.Request.Context(), userID, id); err != nil {
		writeMappedErr(c, err)
		return
	}
	response.NoContent(c)
}

// GetPreferences handles GET /api/v1/notifications/preferences.
func (h *Handler) GetPreferences(c *gin.Context) {
	userID, ok := ensureUser(c)
	if !ok {
		return
	}
	pref, err := h.svc.GetPreferences(c.Request.Context(), userID)
	if err != nil {
		writeMappedErr(c, err)
		return
	}
	response.OK(c, toPreferenceView(pref))
}

// UpdatePreferences handles PUT /api/v1/notifications/preferences.
func (h *Handler) UpdatePreferences(c *gin.Context) {
	userID, ok := ensureUser(c)
	if !ok {
		return
	}
	var req updatePreferencesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, errors.New(errors.ErrInvalidParam.Code, err.Error()))
		return
	}
	pref, err := h.svc.UpdatePreferences(c.Request.Context(), appNotif.UpdatePreferencesInput{
		UserID:         userID,
		InAppEnabled:   req.InAppEnabled,
		EmailEnabled:   req.EmailEnabled,
		WebhookEnabled: req.WebhookEnabled,
		PushEnabled:    req.PushEnabled,
	})
	if err != nil {
		writeMappedErr(c, err)
		return
	}
	response.OK(c, toPreferenceView(pref))
}

// CreateSubscription handles POST /api/v1/notifications/subscriptions.
func (h *Handler) CreateSubscription(c *gin.Context) {
	userID, ok := ensureUser(c)
	if !ok {
		return
	}
	var req createSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, errors.New(errors.ErrInvalidParam.Code, err.Error()))
		return
	}
	sub, err := h.svc.CreateSubscription(c.Request.Context(), appNotif.CreateSubscriptionInput{
		SubscriberID: userID,
		ObjectType:   req.ObjectType,
		ObjectID:     req.ObjectID,
	})
	if err != nil {
		writeMappedErr(c, err)
		return
	}
	response.Created(c, toSubscriptionView(sub))
}

// CancelSubscription handles DELETE /api/v1/notifications/subscriptions/:id.
func (h *Handler) CancelSubscription(c *gin.Context) {
	userID, ok := ensureUser(c)
	if !ok {
		return
	}
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	if err := h.svc.CancelSubscription(c.Request.Context(), userID, id); err != nil {
		writeMappedErr(c, err)
		return
	}
	response.NoContent(c)
}

// ListSubscriptions handles GET /api/v1/notifications/subscriptions.
func (h *Handler) ListSubscriptions(c *gin.Context) {
	userID, ok := ensureUser(c)
	if !ok {
		return
	}
	limit, cursor := readPaging(c)
	objectType := c.Query("object_type")
	out, err := h.svc.ListSubscriptions(c.Request.Context(), appNotif.ListSubscriptionsInput{
		SubscriberID: userID,
		ObjectType:   objectType,
		Limit:        limit,
		Cursor:       cursor,
	})
	if err != nil {
		writeMappedErr(c, err)
		return
	}
	response.OK(c, gin.H{
		"items":       toSubscriptionViews(out.Items),
		"next_cursor": out.NextCursor,
	})
}

// --- view helpers ---

func toNotificationView(n *domNotif.Notification) gin.H {
	if n == nil {
		return nil
	}
	view := gin.H{
		"id":         n.ID,
		"user_id":    n.UserID,
		"type":       n.Type.String(),
		"title":      n.Title,
		"content":    n.Content,
		"status":     n.Status.String(),
		"created_at": n.CreatedAt.Unix(),
	}
	if n.ReadAt != nil {
		view["read_at"] = n.ReadAt.Unix()
	}
	return view
}

func toNotificationViews(items []*domNotif.Notification) []gin.H {
	out := make([]gin.H, 0, len(items))
	for _, n := range items {
		out = append(out, toNotificationView(n))
	}
	return out
}

func toPreferenceView(p *domPref.Preference) gin.H {
	if p == nil {
		return gin.H{}
	}
	updated := time.Time{}
	if !p.UpdatedAt.IsZero() {
		updated = p.UpdatedAt
	}
	return gin.H{
		"user_id":         p.UserID,
		"in_app_enabled":  p.InAppEnabled,
		"email_enabled":   p.EmailEnabled,
		"webhook_enabled": p.WebhookEnabled,
		"push_enabled":    p.PushEnabled,
		"updated_at":      updated.Unix(),
	}
}

func toSubscriptionView(s *domSub.Subscription) gin.H {
	if s == nil {
		return nil
	}
	return gin.H{
		"id":            s.ID,
		"subscriber_id": s.SubscriberID,
		"object_type":   s.ObjectType,
		"object_id":     s.ObjectID,
		"created_at":    s.CreatedAt.Unix(),
	}
}

func toSubscriptionViews(items []*domSub.Subscription) []gin.H {
	out := make([]gin.H, 0, len(items))
	for _, s := range items {
		out = append(out, toSubscriptionView(s))
	}
	return out
}

// Compile-time assertion the value object converters stay live so a
// renamed enum surfaces during the build instead of at runtime.
var _ = valueobject.NotificationStatusUnread
