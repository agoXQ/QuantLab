package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	commonpb "github.com/agoXQ/QuantLab/api/common/v1"
	strategypb "github.com/agoXQ/QuantLab/app/strategy/pb"
	"github.com/agoXQ/QuantLab/pkg/errors"
	"github.com/agoXQ/QuantLab/pkg/response"
)

// registerStrategies mounts /api/v1/strategies.
func (h *Handler) registerStrategies(rg *gin.RouterGroup) {
	s := rg.Group("/strategies")
	s.POST("", h.strategyCreate)
	s.GET("", h.strategyList)
	s.GET("/:id", h.strategyGet)
	s.PUT("/:id", h.strategyUpdate)
	s.DELETE("/:id", h.strategyDelete)

	s.POST("/:id/versions", h.strategyCreateVersion)
	s.GET("/:id/versions", h.strategyListVersions)
	s.GET("/versions/:vid", h.strategyGetVersion)

	s.POST("/:id/publish", h.strategyPublish)
	s.POST("/:id/archive", h.strategyArchive)
	s.POST("/:id/fork", h.strategyFork)
}

func (h *Handler) strategyCreate(c *gin.Context) {
	var req struct {
		Title       string   `json:"title" binding:"required"`
		Description string   `json:"description"`
		Category    string   `json:"category"`
		Tags        []string `json:"tags"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, errors.ErrInvalidParam)
		return
	}
	out, err := h.svc.Strategy.CreateStrategy(c.Request.Context(), &strategypb.CreateStrategyRequest{
		Title: req.Title, Description: req.Description, Category: req.Category, Tags: req.Tags,
	})
	if err != nil {
		grpcErr(c, err)
		return
	}
	response.Created(c, gin.H{"strategy_id": out.StrategyId})
}

// strategyList routes to SearchStrategies when keyword/tag/category is
// present, otherwise ListStrategies.
func (h *Handler) strategyList(c *gin.Context) {
	keyword := c.Query("keyword")
	tag := c.Query("tag")
	category := c.Query("category")
	if keyword != "" || tag != "" || category != "" || c.Query("sort") != "" {
		tags := []string{}
		if tag != "" {
			tags = append(tags, tag)
		}
		out, err := h.svc.Strategy.SearchStrategies(c.Request.Context(), &strategypb.SearchStrategiesRequest{
			Keyword:  keyword,
			Tags:     tags,
			Category: category,
			Sort:     c.Query("sort"),
			Cursor:   &commonpb.Cursor{NextCursor: queryCursor(c)},
			Limit:    queryLimit(c),
		})
		if err != nil {
			grpcErr(c, err)
			return
		}
		response.OKWithMeta(c, gin.H{"items": out.Strategies}, out.Cursor)
		return
	}
	out, err := h.svc.Strategy.ListStrategies(c.Request.Context(), &strategypb.ListStrategiesRequest{
		AuthorId: queryInt64(c, "author_id"),
		Cursor:   &commonpb.Cursor{NextCursor: queryCursor(c)},
		Limit:    queryLimit(c),
	})
	if err != nil {
		grpcErr(c, err)
		return
	}
	response.OKWithMeta(c, gin.H{"items": out.Strategies}, out.Cursor)
}

func (h *Handler) strategyGet(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	out, err := h.svc.Strategy.GetStrategy(c.Request.Context(), &strategypb.GetStrategyRequest{StrategyId: id})
	if err != nil {
		grpcErr(c, err)
		return
	}
	response.OK(c, gin.H{"strategy": out.Strategy})
}

func (h *Handler) strategyUpdate(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	var req struct {
		Title       *string   `json:"title"`
		Description *string   `json:"description"`
		Category    *string   `json:"category"`
		Tags        *[]string `json:"tags"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, errors.ErrInvalidParam)
		return
	}
	r := &strategypb.UpdateStrategyRequest{StrategyId: id}
	if req.Title != nil {
		r.Title = *req.Title
	}
	if req.Description != nil {
		r.Description = *req.Description
	}
	if req.Category != nil {
		r.Category = *req.Category
	}
	if req.Tags != nil {
		r.Tags = *req.Tags
	}
	_, err := h.svc.Strategy.UpdateStrategy(c.Request.Context(), r)
	if err != nil {
		grpcErr(c, err)
		return
	}
	response.OK(c, gin.H{})
}

func (h *Handler) strategyDelete(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	if _, err := h.svc.Strategy.DeleteStrategy(c.Request.Context(), &strategypb.DeleteStrategyRequest{StrategyId: id}); err != nil {
		grpcErr(c, err)
		return
	}
	response.NoContent(c)
}

func (h *Handler) strategyCreateVersion(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	var req struct {
		FormulaText   string `json:"formula_text" binding:"required"`
		BuyRule       string `json:"buy_rule"`
		SellRule      string `json:"sell_rule"`
		RiskRule      string `json:"risk_rule"`
		PositionRule  string `json:"position_rule"`
		RebalanceRule string `json:"rebalance_rule"`
		ChangeLog     string `json:"change_log"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, errors.ErrInvalidParam)
		return
	}
	out, err := h.svc.Strategy.CreateVersion(c.Request.Context(), &strategypb.CreateVersionRequest{
		StrategyId:    id,
		FormulaText:   req.FormulaText,
		BuyRule:       req.BuyRule,
		SellRule:      req.SellRule,
		RiskRule:      req.RiskRule,
		PositionRule:  req.PositionRule,
		RebalanceRule: req.RebalanceRule,
		ChangeLog:     req.ChangeLog,
	})
	if err != nil {
		grpcErr(c, err)
		return
	}
	response.Created(c, gin.H{"version_id": out.VersionId})
}

func (h *Handler) strategyListVersions(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	out, err := h.svc.Strategy.ListVersions(c.Request.Context(), &strategypb.ListVersionsRequest{StrategyId: id})
	if err != nil {
		grpcErr(c, err)
		return
	}
	response.OK(c, gin.H{"items": out.Versions})
}

func (h *Handler) strategyGetVersion(c *gin.Context) {
	vid, ok := parseID(c, "vid")
	if !ok {
		return
	}
	out, err := h.svc.Strategy.GetVersion(c.Request.Context(), &strategypb.GetVersionRequest{VersionId: vid})
	if err != nil {
		grpcErr(c, err)
		return
	}
	response.OK(c, gin.H{"version": out.Version})
}

func (h *Handler) strategyPublish(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	var req struct {
		VersionID int64 `json:"version_id"`
	}
	_ = c.ShouldBindJSON(&req)
	if _, err := h.svc.Strategy.PublishStrategy(c.Request.Context(), &strategypb.PublishStrategyRequest{
		StrategyId: id, VersionId: req.VersionID,
	}); err != nil {
		grpcErr(c, err)
		return
	}
	response.OK(c, gin.H{})
}

func (h *Handler) strategyArchive(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	if _, err := h.svc.Strategy.ArchiveStrategy(c.Request.Context(), &strategypb.ArchiveStrategyRequest{StrategyId: id}); err != nil {
		grpcErr(c, err)
		return
	}
	response.OK(c, gin.H{})
}

func (h *Handler) strategyFork(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	out, err := h.svc.Strategy.ForkStrategy(c.Request.Context(), &strategypb.ForkStrategyRequest{SourceStrategyId: id})
	if err != nil {
		grpcErr(c, err)
		return
	}
	response.Created(c, gin.H{"strategy_id": out.NewStrategyId})
}
