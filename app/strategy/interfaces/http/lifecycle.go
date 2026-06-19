package http

import (
	"github.com/gin-gonic/gin"
	"net/http"

	appStrategy "github.com/agoXQ/QuantLab/app/strategy/application/strategy"
	"github.com/agoXQ/QuantLab/pkg/errors"
	"github.com/agoXQ/QuantLab/pkg/response"
)

// CreateVersion handles POST /api/v1/strategies/:id/versions.
func (h *Handler) CreateVersion(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	var req createVersionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, errors.ErrInvalidParam)
		return
	}
	res, err := h.svc.CreateVersion(c.Request.Context(), appStrategy.CreateVersionRequest{
		StrategyID:    id,
		CallerID:      req.CallerID,
		FormulaText:   req.FormulaText,
		BuyRule:       req.BuyRule,
		SellRule:      req.SellRule,
		RiskRule:      req.RiskRule,
		PositionRule:  req.PositionRule,
		RebalanceRule: req.RebalanceRule,
		ChangeLog:     req.ChangeLog,
	})
	if err != nil {
		writeMappedErr(c, err)
		return
	}
	response.Created(c, gin.H{"strategy": res.Strategy, "version": res.Version})
}

// ListVersions handles GET /api/v1/strategies/:id/versions.
func (h *Handler) ListVersions(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	limit := int(parseInt64(c.Query("limit")))
	versions, err := h.svc.ListVersions(c.Request.Context(), id, limit)
	if err != nil {
		writeMappedErr(c, err)
		return
	}
	response.OK(c, gin.H{"items": versions})
}

// GetVersion handles GET /api/v1/strategies/versions/:vid.
func (h *Handler) GetVersion(c *gin.Context) {
	vid, ok := parseID(c, "vid")
	if !ok {
		return
	}
	ver, err := h.svc.GetVersion(c.Request.Context(), vid)
	if err != nil {
		writeMappedErr(c, err)
		return
	}
	response.OK(c, gin.H{"version": ver})
}

// Publish handles POST /api/v1/strategies/:id/publish.
func (h *Handler) Publish(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	var req publishRequest
	_ = c.ShouldBindJSON(&req)
	st, err := h.svc.Publish(c.Request.Context(), appStrategy.PublishRequest{
		StrategyID: id,
		CallerID:   req.CallerID,
		VersionID:  req.VersionID,
	})
	if err != nil {
		writeMappedErr(c, err)
		return
	}
	response.OK(c, gin.H{"strategy": st})
}

// Archive handles POST /api/v1/strategies/:id/archive.
func (h *Handler) Archive(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	var req archiveRequest
	_ = c.ShouldBindJSON(&req)
	st, err := h.svc.Archive(c.Request.Context(), appStrategy.ArchiveRequest{
		StrategyID: id,
		CallerID:   req.CallerID,
	})
	if err != nil {
		writeMappedErr(c, err)
		return
	}
	response.OK(c, gin.H{"strategy": st})
}

// Fork handles POST /api/v1/strategies/:id/fork.
func (h *Handler) Fork(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	var req forkRequest
	_ = c.ShouldBindJSON(&req)
	res, err := h.svc.Fork(c.Request.Context(), appStrategy.ForkRequest{
		SourceStrategyID: id,
		CallerID:         req.CallerID,
		Title:            req.Title,
	})
	if err != nil {
		writeMappedErr(c, err)
		return
	}
	response.Created(c, gin.H{"strategy": res.Strategy, "version": res.Version})
}
