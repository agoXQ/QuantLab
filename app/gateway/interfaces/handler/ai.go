package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	aipb "github.com/agoXQ/QuantLab/app/ai/pb"
	"github.com/agoXQ/QuantLab/pkg/errors"
	"github.com/agoXQ/QuantLab/pkg/response"
)

// registerAI mounts /api/v1/ai. AI operations are async (return a task
// id); the frontend polls GetTask until COMPLETED.
func (h *Handler) registerAI(rg *gin.RouterGroup) {
	a := rg.Group("/ai")
	a.POST("/strategies/generate", h.aiGenerateStrategy)
	a.POST("/strategies/:id/explain", h.aiExplainStrategy)
	a.POST("/strategies/:id/optimize", h.aiOptimizeStrategy)
	a.POST("/portfolios/generate", h.aiGeneratePortfolio)
	a.POST("/portfolios/:id/optimize", h.aiOptimizePortfolio)
	a.POST("/backtests/:id/analyze", h.aiAnalyzeBacktest)
	a.POST("/chat", h.aiChat)
	a.GET("/tasks/:id", h.aiGetTask)
	a.GET("/reports", h.aiGetReport)
}

func (h *Handler) aiGenerateStrategy(c *gin.Context) {
	var req struct {
		Prompt string `json:"prompt" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, errors.ErrInvalidParam)
		return
	}
	out, err := h.svc.AI.GenerateStrategy(c.Request.Context(), &aipb.GenerateStrategyRequest{Prompt: req.Prompt})
	if err != nil {
		grpcErr(c, err)
		return
	}
	response.Created(c, gin.H{"task_id": out.TaskId})
}

func (h *Handler) aiExplainStrategy(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	out, err := h.svc.AI.ExplainStrategy(c.Request.Context(), &aipb.ExplainStrategyRequest{StrategyId: id})
	if err != nil {
		grpcErr(c, err)
		return
	}
	response.Created(c, gin.H{"task_id": out.TaskId})
}

func (h *Handler) aiOptimizeStrategy(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	out, err := h.svc.AI.OptimizeStrategy(c.Request.Context(), &aipb.OptimizeStrategyRequest{StrategyId: id})
	if err != nil {
		grpcErr(c, err)
		return
	}
	response.Created(c, gin.H{"task_id": out.TaskId})
}

func (h *Handler) aiGeneratePortfolio(c *gin.Context) {
	var req struct {
		Prompt string `json:"prompt" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, errors.ErrInvalidParam)
		return
	}
	out, err := h.svc.AI.GeneratePortfolio(c.Request.Context(), &aipb.GeneratePortfolioRequest{Prompt: req.Prompt})
	if err != nil {
		grpcErr(c, err)
		return
	}
	response.Created(c, gin.H{"task_id": out.TaskId})
}

func (h *Handler) aiOptimizePortfolio(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	out, err := h.svc.AI.OptimizePortfolio(c.Request.Context(), &aipb.OptimizePortfolioRequest{PortfolioId: id})
	if err != nil {
		grpcErr(c, err)
		return
	}
	response.Created(c, gin.H{"task_id": out.TaskId})
}

func (h *Handler) aiAnalyzeBacktest(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	out, err := h.svc.AI.AnalyzeBacktest(c.Request.Context(), &aipb.AnalyzeBacktestRequest{BacktestId: id})
	if err != nil {
		grpcErr(c, err)
		return
	}
	response.Created(c, gin.H{"task_id": out.TaskId})
}

func (h *Handler) aiChat(c *gin.Context) {
	var req struct {
		Message        string `json:"message" binding:"required"`
		ConversationID int64  `json:"conversation_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, errors.ErrInvalidParam)
		return
	}
	out, err := h.svc.AI.Chat(c.Request.Context(), &aipb.ChatRequest{
		Message: req.Message, ConversationId: req.ConversationID,
	})
	if err != nil {
		grpcErr(c, err)
		return
	}
	response.OK(c, gin.H{"reply": out.Reply, "conversation_id": out.ConversationId})
}

func (h *Handler) aiGetTask(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	out, err := h.svc.AI.GetTask(c.Request.Context(), &aipb.GetTaskRequest{TaskId: id})
	if err != nil {
		grpcErr(c, err)
		return
	}
	response.OK(c, gin.H{"task": out.Task})
}

func (h *Handler) aiGetReport(c *gin.Context) {
	out, err := h.svc.AI.GetReport(c.Request.Context(), &aipb.GetReportRequest{
		ObjectType: c.Query("object_type"), ObjectId: queryInt64(c, "object_id"),
	})
	if err != nil {
		grpcErr(c, err)
		return
	}
	response.OK(c, gin.H{"report": out.Report})
}

var _ = http.StatusOK
