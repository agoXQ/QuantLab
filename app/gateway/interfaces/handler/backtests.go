package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	commonpb "github.com/agoXQ/QuantLab/api/common/v1"
	backtestpb "github.com/agoXQ/QuantLab/app/backtest/pb"
	"github.com/agoXQ/QuantLab/app/gateway/interfaces/middleware"
	"github.com/agoXQ/QuantLab/pkg/errors"
	"github.com/agoXQ/QuantLab/pkg/response"
)

// registerBacktests mounts /api/v1/backtests.
func (h *Handler) registerBacktests(rg *gin.RouterGroup) {
	b := rg.Group("/backtests")
	b.POST("", h.backtestCreate)
	b.GET("", h.backtestList)
	b.GET("/:id", h.backtestGet)
	b.DELETE("/:id", h.backtestCancel)
	b.GET("/:id/report", h.backtestReport)
	b.GET("/:id/trades", h.backtestTrades)
	b.GET("/:id/positions", h.backtestPositions)
}

func (h *Handler) backtestCreate(c *gin.Context) {
	var req struct {
		StrategyID int64  `json:"strategy_id" binding:"required"`
		VersionID  int64  `json:"version_id"`
		StartDate  string `json:"start_date" binding:"required"`
		EndDate    string `json:"end_date" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, errors.ErrInvalidParam)
		return
	}
	// Outgoing forwards the authenticated caller id as "x-user-id"
	// gRPC metadata so the backtest service can stamp it on the job.
	out, err := h.svc.Backtest.CreateBacktest(middleware.Outgoing(c), &backtestpb.CreateBacktestRequest{
		StrategyId: req.StrategyID, VersionId: req.VersionID,
		StartDate: req.StartDate, EndDate: req.EndDate,
	})
	if err != nil {
		grpcErr(c, err)
		return
	}
	response.Created(c, gin.H{"job_id": out.JobId})
}

func (h *Handler) backtestList(c *gin.Context) {
	// Prefer the explicit user_id query param (the frontend sends it);
	// fall back to the caller id resolved from the JWT so a request
	// without the param still returns the caller's own jobs.
	uid := queryInt64(c, "user_id")
	if uid <= 0 {
		uid = middleware.CallerID(c)
	}
	out, err := h.svc.Backtest.ListBacktests(middleware.Outgoing(c), &backtestpb.ListBacktestsRequest{
		UserId: uid,
		Cursor: &commonpb.Cursor{NextCursor: queryCursor(c)},
		Limit:  queryLimit(c),
	})
	if err != nil {
		grpcErr(c, err)
		return
	}
	response.OKWithMeta(c, gin.H{"items": out.Jobs}, out.Cursor)
}

func (h *Handler) backtestGet(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	out, err := h.svc.Backtest.GetBacktest(c.Request.Context(), &backtestpb.GetBacktestRequest{JobId: id})
	if err != nil {
		grpcErr(c, err)
		return
	}
	response.OK(c, gin.H{"job": out.Job, "config": out.Config})
}

func (h *Handler) backtestCancel(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	if _, err := h.svc.Backtest.CancelBacktest(c.Request.Context(), &backtestpb.CancelBacktestRequest{JobId: id}); err != nil {
		grpcErr(c, err)
		return
	}
	response.NoContent(c)
}

func (h *Handler) backtestReport(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	out, err := h.svc.Backtest.GetReport(c.Request.Context(), &backtestpb.GetReportRequest{JobId: id})
	if err != nil {
		grpcErr(c, err)
		return
	}
	response.OK(c, gin.H{"report": out.Report})
}

func (h *Handler) backtestTrades(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	out, err := h.svc.Backtest.GetTrades(c.Request.Context(), &backtestpb.GetTradesRequest{
		JobId: id, Cursor: &commonpb.Cursor{NextCursor: queryCursor(c)}, Limit: queryLimit(c),
	})
	if err != nil {
		grpcErr(c, err)
		return
	}
	response.OKWithMeta(c, gin.H{"items": out.Trades}, out.Cursor)
}

func (h *Handler) backtestPositions(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	out, err := h.svc.Backtest.GetPositions(c.Request.Context(), &backtestpb.GetPositionsRequest{
		JobId: id, TradeDate: c.Query("trade_date"),
	})
	if err != nil {
		grpcErr(c, err)
		return
	}
	response.OK(c, gin.H{"items": out.Positions})
}
