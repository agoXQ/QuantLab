package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	commonpb "github.com/agoXQ/QuantLab/api/common/v1"
	portfoliopb "github.com/agoXQ/QuantLab/app/portfolio/pb"
	"github.com/agoXQ/QuantLab/pkg/errors"
	"github.com/agoXQ/QuantLab/pkg/response"
)

// registerPortfolios mounts /api/v1/portfolios.
func (h *Handler) registerPortfolios(rg *gin.RouterGroup) {
	p := rg.Group("/portfolios")
	p.POST("", h.portfolioCreate)
	p.GET("", h.portfolioList)
	p.GET("/:id", h.portfolioGet)
	p.PUT("/:id", h.portfolioUpdate)
	p.DELETE("/:id", h.portfolioDelete)
	p.POST("/:id/items", h.portfolioAddItem)
	p.DELETE("/:id/items/:item_id", h.portfolioRemoveItem)
	p.PUT("/:id/weights", h.portfolioUpdateWeights)
	p.POST("/:id/publish", h.portfolioPublish)
	p.GET("/:id/analytics", h.portfolioAnalytics)
	p.GET("/:id/equity-curve", h.portfolioEquityCurve)
}

func (h *Handler) portfolioCreate(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
		Visibility  string `json:"visibility"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, errors.ErrInvalidParam)
		return
	}
	out, err := h.svc.Portfolio.CreatePortfolio(c.Request.Context(), &portfoliopb.CreatePortfolioRequest{
		Name: req.Name, Description: req.Description, Visibility: parseVisibility(req.Visibility),
	})
	if err != nil {
		grpcErr(c, err)
		return
	}
	response.Created(c, gin.H{"portfolio_id": out.PortfolioId})
}

func (h *Handler) portfolioList(c *gin.Context) {
	out, err := h.svc.Portfolio.ListPortfolios(c.Request.Context(), &portfoliopb.ListPortfoliosRequest{
		OwnerId: queryInt64(c, "owner_id"),
		Cursor:  &commonpb.Cursor{NextCursor: queryCursor(c)},
		Limit:   queryLimit(c),
	})
	if err != nil {
		grpcErr(c, err)
		return
	}
	response.OKWithMeta(c, gin.H{"items": out.Portfolios}, out.Cursor)
}

func (h *Handler) portfolioGet(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	out, err := h.svc.Portfolio.GetPortfolio(c.Request.Context(), &portfoliopb.GetPortfolioRequest{PortfolioId: id})
	if err != nil {
		grpcErr(c, err)
		return
	}
	response.OK(c, gin.H{"portfolio": out.Portfolio})
}

func (h *Handler) portfolioUpdate(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	var req struct {
		Name        *string `json:"name"`
		Description *string `json:"description"`
		Visibility  *string `json:"visibility"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, errors.ErrInvalidParam)
		return
	}
	r := &portfoliopb.UpdatePortfolioRequest{PortfolioId: id}
	if req.Name != nil {
		r.Name = *req.Name
	}
	if req.Description != nil {
		r.Description = *req.Description
	}
	if req.Visibility != nil {
		r.Visibility = parseVisibility(*req.Visibility)
	}
	if _, err := h.svc.Portfolio.UpdatePortfolio(c.Request.Context(), r); err != nil {
		grpcErr(c, err)
		return
	}
	response.OK(c, gin.H{})
}

func (h *Handler) portfolioDelete(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	if _, err := h.svc.Portfolio.DeletePortfolio(c.Request.Context(), &portfoliopb.DeletePortfolioRequest{PortfolioId: id}); err != nil {
		grpcErr(c, err)
		return
	}
	response.NoContent(c)
}

func (h *Handler) portfolioAddItem(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	var req struct {
		StrategyID int64   `json:"strategy_id" binding:"required"`
		Weight     float64 `json:"weight"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, errors.ErrInvalidParam)
		return
	}
	out, err := h.svc.Portfolio.AddItem(c.Request.Context(), &portfoliopb.AddItemRequest{
		PortfolioId: id, StrategyId: req.StrategyID, Weight: req.Weight,
	})
	if err != nil {
		grpcErr(c, err)
		return
	}
	response.Created(c, gin.H{"item_id": out.ItemId})
}

func (h *Handler) portfolioRemoveItem(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	itemID, ok := parseID(c, "item_id")
	if !ok {
		return
	}
	if _, err := h.svc.Portfolio.RemoveItem(c.Request.Context(), &portfoliopb.RemoveItemRequest{
		PortfolioId: id, ItemId: itemID,
	}); err != nil {
		grpcErr(c, err)
		return
	}
	response.NoContent(c)
}

func (h *Handler) portfolioUpdateWeights(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	var req struct {
		Items []struct {
			StrategyID int64   `json:"strategy_id"`
			Weight     float64 `json:"weight"`
		} `json:"items" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, errors.ErrInvalidParam)
		return
	}
	items := make([]*portfoliopb.WeightItem, 0, len(req.Items))
	for _, it := range req.Items {
		items = append(items, &portfoliopb.WeightItem{StrategyId: it.StrategyID, Weight: it.Weight})
	}
	if _, err := h.svc.Portfolio.UpdateWeights(c.Request.Context(), &portfoliopb.UpdateWeightsRequest{
		PortfolioId: id, Items: items,
	}); err != nil {
		grpcErr(c, err)
		return
	}
	response.OK(c, gin.H{})
}

func (h *Handler) portfolioPublish(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	if _, err := h.svc.Portfolio.PublishPortfolio(c.Request.Context(), &portfoliopb.PublishPortfolioRequest{PortfolioId: id}); err != nil {
		grpcErr(c, err)
		return
	}
	response.OK(c, gin.H{})
}

func (h *Handler) portfolioAnalytics(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	out, err := h.svc.Portfolio.GetAnalytics(c.Request.Context(), &portfoliopb.GetAnalyticsRequest{PortfolioId: id})
	if err != nil {
		grpcErr(c, err)
		return
	}
	response.OK(c, gin.H{"analytics": out.Analytics})
}

func (h *Handler) portfolioEquityCurve(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	out, err := h.svc.Portfolio.GetEquityCurve(c.Request.Context(), &portfoliopb.GetEquityCurveRequest{PortfolioId: id})
	if err != nil {
		grpcErr(c, err)
		return
	}
	response.OK(c, gin.H{"items": out.Points})
}

// parseVisibility maps the REST string to the proto enum; defaults to
// PUBLIC when unset so the common case needs no extra field.
func parseVisibility(s string) commonpb.Visibility {
	switch s {
	case "private", "PRIVATE", "1":
		return commonpb.Visibility_PRIVATE
	case "unlisted", "UNLISTED", "3":
		return commonpb.Visibility_UNLISTED
	default:
		return commonpb.Visibility_PUBLIC
	}
}
