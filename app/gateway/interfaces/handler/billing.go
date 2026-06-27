package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	commonpb "github.com/agoXQ/QuantLab/api/common/v1"
	billingpb "github.com/agoXQ/QuantLab/app/billing/pb"
	"github.com/agoXQ/QuantLab/pkg/errors"
	"github.com/agoXQ/QuantLab/pkg/response"
)

// registerBilling mounts /api/v1/billing.
func (h *Handler) registerBilling(rg *gin.RouterGroup) {
	b := rg.Group("/billing")
	b.GET("/membership", h.billingGetMembership)
	b.POST("/membership/purchase", h.billingPurchaseMembership)
	b.POST("/membership/cancel", h.billingCancelMembership)

	b.POST("/subscriptions", h.billingCreateSubscription)
	b.DELETE("/subscriptions/:id", h.billingCancelSubscription)
	b.GET("/subscriptions", h.billingListSubscriptions)

	b.POST("/orders", h.billingCreateOrder)
	b.GET("/orders/:id", h.billingGetOrder)
	b.GET("/orders", h.billingListOrders)

	b.POST("/payments", h.billingCreatePayment)
	b.POST("/payments/webhook", h.billingPaymentWebhook)

	b.GET("/creators/:id/revenue", h.billingCreatorRevenue)
	b.POST("/settlements", h.billingRequestSettlement)
	b.GET("/settlements", h.billingListSettlements)
}

func (h *Handler) billingGetMembership(c *gin.Context) {
	out, err := h.svc.Billing.GetMembership(c.Request.Context(), &billingpb.GetMembershipRequest{
		UserId: queryInt64(c, "user_id"),
	})
	if err != nil {
		grpcErr(c, err)
		return
	}
	response.OK(c, gin.H{"membership": out.Membership})
}

func (h *Handler) billingPurchaseMembership(c *gin.Context) {
	var req struct {
		Tier         string `json:"tier" binding:"required"`
		BillingCycle string `json:"billing_cycle" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, errors.ErrInvalidParam)
		return
	}
	out, err := h.svc.Billing.PurchaseMembership(c.Request.Context(), &billingpb.PurchaseMembershipRequest{
		Tier: req.Tier, BillingCycle: req.BillingCycle,
	})
	if err != nil {
		grpcErr(c, err)
		return
	}
	response.Created(c, gin.H{"order_id": out.OrderId})
}

func (h *Handler) billingCancelMembership(c *gin.Context) {
	if _, err := h.svc.Billing.CancelMembership(c.Request.Context(), &billingpb.CancelMembershipRequest{}); err != nil {
		grpcErr(c, err)
		return
	}
	response.OK(c, gin.H{})
}

func (h *Handler) billingCreateSubscription(c *gin.Context) {
	var req struct {
		ResourceType string `json:"resource_type" binding:"required"`
		ResourceID   string `json:"resource_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, errors.ErrInvalidParam)
		return
	}
	out, err := h.svc.Billing.CreateSubscription(c.Request.Context(), &billingpb.CreateSubscriptionRequest{
		ResourceType: req.ResourceType, ResourceId: req.ResourceID,
	})
	if err != nil {
		grpcErr(c, err)
		return
	}
	response.Created(c, gin.H{"subscription_id": out.SubscriptionId})
}

func (h *Handler) billingCancelSubscription(c *gin.Context) {
	sid := c.Param("id")
	if _, err := h.svc.Billing.CancelSubscription(c.Request.Context(), &billingpb.CancelSubscriptionRequest{SubscriptionId: sid}); err != nil {
		grpcErr(c, err)
		return
	}
	response.NoContent(c)
}

func (h *Handler) billingListSubscriptions(c *gin.Context) {
	out, err := h.svc.Billing.ListSubscriptions(c.Request.Context(), &billingpb.ListSubscriptionsRequest{
		UserId: queryInt64(c, "user_id"),
		Cursor: &commonpb.Cursor{NextCursor: queryCursor(c)},
		Limit:  queryLimit(c),
	})
	if err != nil {
		grpcErr(c, err)
		return
	}
	response.OKWithMeta(c, gin.H{"items": out.Subscriptions}, out.Cursor)
}

func (h *Handler) billingCreateOrder(c *gin.Context) {
	var req struct {
		OrderType  string  `json:"order_type" binding:"required"`
		ResourceID string  `json:"resource_id"`
		Amount     float64 `json:"amount"`
		Currency   string  `json:"currency"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, errors.ErrInvalidParam)
		return
	}
	out, err := h.svc.Billing.CreateOrder(c.Request.Context(), &billingpb.CreateOrderRequest{
		OrderType: req.OrderType, ResourceId: req.ResourceID, Amount: req.Amount, Currency: req.Currency,
	})
	if err != nil {
		grpcErr(c, err)
		return
	}
	response.Created(c, gin.H{"order_id": out.OrderId})
}

func (h *Handler) billingGetOrder(c *gin.Context) {
	oid := c.Param("id")
	out, err := h.svc.Billing.GetOrder(c.Request.Context(), &billingpb.GetOrderRequest{OrderId: oid})
	if err != nil {
		grpcErr(c, err)
		return
	}
	response.OK(c, gin.H{"order": out.Order})
}

func (h *Handler) billingListOrders(c *gin.Context) {
	out, err := h.svc.Billing.ListOrders(c.Request.Context(), &billingpb.ListOrdersRequest{
		UserId: queryInt64(c, "user_id"),
		Cursor: &commonpb.Cursor{NextCursor: queryCursor(c)},
		Limit:  queryLimit(c),
	})
	if err != nil {
		grpcErr(c, err)
		return
	}
	response.OKWithMeta(c, gin.H{"items": out.Orders}, out.Cursor)
}

func (h *Handler) billingCreatePayment(c *gin.Context) {
	var req struct {
		OrderID string `json:"order_id" binding:"required"`
		Channel string `json:"channel" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, errors.ErrInvalidParam)
		return
	}
	out, err := h.svc.Billing.CreatePayment(c.Request.Context(), &billingpb.CreatePaymentRequest{
		OrderId: req.OrderID, Channel: req.Channel,
	})
	if err != nil {
		grpcErr(c, err)
		return
	}
	response.OK(c, gin.H{"payment_url": out.PaymentUrl})
}

func (h *Handler) billingPaymentWebhook(c *gin.Context) {
	raw, _ := c.GetRawData()
	if _, err := h.svc.Billing.PaymentWebhook(c.Request.Context(), &billingpb.PaymentWebhookRequest{
		Channel: c.Query("channel"), RawBody: raw,
	}); err != nil {
		grpcErr(c, err)
		return
	}
	response.OK(c, gin.H{})
}

func (h *Handler) billingCreatorRevenue(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	out, err := h.svc.Billing.GetCreatorRevenue(c.Request.Context(), &billingpb.GetCreatorRevenueRequest{CreatorId: id})
	if err != nil {
		grpcErr(c, err)
		return
	}
	response.OK(c, out)
}

func (h *Handler) billingRequestSettlement(c *gin.Context) {
	var req struct {
		Amount   float64 `json:"amount" binding:"required"`
		Currency string  `json:"currency"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, errors.ErrInvalidParam)
		return
	}
	out, err := h.svc.Billing.RequestSettlement(c.Request.Context(), &billingpb.RequestSettlementRequest{
		Amount: req.Amount, Currency: req.Currency,
	})
	if err != nil {
		grpcErr(c, err)
		return
	}
	response.Created(c, gin.H{"settlement_id": out.SettlementId})
}

func (h *Handler) billingListSettlements(c *gin.Context) {
	out, err := h.svc.Billing.ListSettlements(c.Request.Context(), &billingpb.ListSettlementsRequest{
		CreatorId: queryInt64(c, "creator_id"),
		Cursor:    &commonpb.Cursor{NextCursor: queryCursor(c)},
		Limit:     queryLimit(c),
	})
	if err != nil {
		grpcErr(c, err)
		return
	}
	response.OKWithMeta(c, gin.H{"items": out.Settlements}, out.Cursor)
}

var _ = http.StatusOK
