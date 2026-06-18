// Package http exposes the Backtest Engine application service over a thin
// HTTP surface aligned with the QuantLab API Design Standard.
package http

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	appBacktest "github.com/agoXQ/QuantLab/app/backtest/application/backtest"
	"github.com/agoXQ/QuantLab/app/backtest/domain/backtestjob"
	"github.com/agoXQ/QuantLab/app/backtest/domain/valueobject"
	"github.com/agoXQ/QuantLab/pkg/errors"
	"github.com/agoXQ/QuantLab/pkg/response"
)

// Handler exposes the Backtest Engine over HTTP.
//
// The handler does not run the backtest synchronously inside Create unless
// the caller asks for it via the run query parameter. This keeps the API
// usable from the queued-worker path that we will introduce alongside
// Kafka, while still letting MVP callers (and integration tests) get the
// "submit + wait" experience in a single round trip.
type Handler struct {
	svc appBacktest.Service
}

// NewHandler returns a Handler bound to the given application service.
func NewHandler(svc appBacktest.Service) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes mounts all backtest routes under the given group, expected
// to be /api/v1/backtests.
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("", h.Create)
	rg.GET("", h.List)
	rg.GET(":id", h.Get)
	rg.POST(":id/run", h.Run)
	rg.POST(":id/cancel", h.Cancel)
	rg.GET(":id/status", h.GetStatus)
	rg.GET(":id/report", h.GetReport)
	rg.GET(":id/trades", h.GetTrades)
	rg.GET(":id/positions", h.GetPositions)
}

// --- request / response shapes ---

type createBacktestRequest struct {
	UserID         int64    `json:"user_id"`
	StrategyID     int64    `json:"strategy_id"`
	VersionID      int64    `json:"version_id"`
	Name           string   `json:"name"`
	Formula        string   `json:"formula" binding:"required"`
	Universe       []string `json:"universe" binding:"required"`
	Benchmark      string   `json:"benchmark"`
	DataVersion    string   `json:"data_version"`
	InitialCapital float64  `json:"initial_capital" binding:"required"`
	StartDate      string   `json:"start_date" binding:"required"`
	EndDate        string   `json:"end_date" binding:"required"`
	Config         configPayload `json:"config"`
}

type configPayload struct {
	CommissionRate     float64 `json:"commission_rate"`
	SlippageRate       float64 `json:"slippage_rate"`
	StampDutyRate      float64 `json:"stamp_duty_rate"`
	MinCommission      float64 `json:"min_commission"`
	MaxPositionCount   int     `json:"max_position_count"`
	RebalanceFrequency string  `json:"rebalance_frequency"`
	LookbackBars       int     `json:"lookback_bars"`
}

type runResponse struct {
	Job    any `json:"job"`
	Report any `json:"report,omitempty"`
}

// statusView is the lightweight projection returned by GET :id/status.
// It is decoupled from the aggregate JSON so we can evolve the projection
// (e.g. add progress percentage) without changing the canonical job
// payload returned by GET :id.
type statusView struct {
	ID           int64      `json:"id"`
	Status       string     `json:"status"`
	Progress     float64    `json:"progress"`
	ErrorMessage string     `json:"error_message,omitempty"`
	StartedAt    *time.Time `json:"started_at,omitempty"`
	FinishedAt   *time.Time `json:"finished_at,omitempty"`
}

// Create handles POST /api/v1/backtests.
//
// When the request carries ?run=true the handler runs the job synchronously
// after persistence and returns the report inline; otherwise it returns the
// freshly-created job for the queued worker to pick up.
func (h *Handler) Create(c *gin.Context) {
	var req createBacktestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, errors.ErrInvalidParam)
		return
	}
	start, err := valueobject.ParseDate(req.StartDate)
	if err != nil {
		response.Error(c, http.StatusBadRequest, errors.New(errors.ErrInvalidFormat.Code, "invalid start_date"))
		return
	}
	end, err := valueobject.ParseDate(req.EndDate)
	if err != nil {
		response.Error(c, http.StatusBadRequest, errors.New(errors.ErrInvalidFormat.Code, "invalid end_date"))
		return
	}
	freq, err := valueobject.ParseRebalanceFrequency(req.Config.RebalanceFrequency)
	if err != nil {
		response.Error(c, http.StatusBadRequest, errors.New(errors.ErrInvalidFormat.Code, "invalid rebalance_frequency"))
		return
	}

	createReq := appBacktest.CreateBacktestRequest{
		UserID:         req.UserID,
		StrategyID:     req.StrategyID,
		VersionID:      req.VersionID,
		Name:           req.Name,
		Formula:        req.Formula,
		Universe:       req.Universe,
		Benchmark:      req.Benchmark,
		DataVersion:    req.DataVersion,
		InitialCapital: req.InitialCapital,
		Range:          valueobject.DateRange{Start: start, End: end},
		Config: backtestjob.Config{
			CommissionRate:     req.Config.CommissionRate,
			SlippageRate:       req.Config.SlippageRate,
			StampDutyRate:      req.Config.StampDutyRate,
			MinCommission:      req.Config.MinCommission,
			MaxPositionCount:   req.Config.MaxPositionCount,
			RebalanceFrequency: freq,
			LookbackBars:       req.Config.LookbackBars,
		},
	}
	created, err := h.svc.Create(c.Request.Context(), createReq)
	if err != nil {
		writeMappedErr(c, err)
		return
	}

	if !shouldRunInline(c) {
		response.Created(c, runResponse{Job: created.Job})
		return
	}
	res, err := h.svc.Run(c.Request.Context(), created.Job.ID)
	if err != nil {
		writeMappedErr(c, err)
		return
	}
	response.Created(c, runResponse{Job: res.Job, Report: res.Report})
}

// shouldRunInline returns true when the caller asks the API to run the job
// synchronously. We honour both the legacy ?run=true (used by the
// existing e2e harness and Create's inline branch) and the new
// ?wait=true (used on POST /:id/run).
func shouldRunInline(c *gin.Context) bool {
	switch strings.ToLower(c.Query("run")) {
	case "1", "true", "yes":
		return true
	}
	switch strings.ToLower(c.Query("wait")) {
	case "1", "true", "yes":
		return true
	}
	return false
}

// Get handles GET /api/v1/backtests/:id.
func (h *Handler) Get(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	job, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		writeMappedErr(c, err)
		return
	}
	response.OK(c, job)
}

// Run handles POST /api/v1/backtests/:id/run.
//
// By default it Submits the job to the queue and returns 202 Accepted so
// the client can poll status; if the caller passes ?wait=true the handler
// runs the job synchronously and returns 200 with the report inline. The
// synchronous path is preserved for the e2e regression harness and for
// debugging; production clients should prefer the asynchronous shape.
func (h *Handler) Run(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	if shouldRunInline(c) {
		res, err := h.svc.Run(c.Request.Context(), id)
		if err != nil {
			writeMappedErr(c, err)
			return
		}
		response.OK(c, runResponse{Job: res.Job, Report: res.Report})
		return
	}
	job, err := h.svc.Submit(c.Request.Context(), id)
	if err != nil {
		writeMappedErr(c, err)
		return
	}
	c.JSON(http.StatusAccepted, gin.H{
		"code":    0,
		"message": "queued",
		"data":    runResponse{Job: job},
	})
}

// Cancel handles POST /api/v1/backtests/:id/cancel.
func (h *Handler) Cancel(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var body struct {
		Reason string `json:"reason"`
	}
	_ = c.ShouldBindJSON(&body)
	job, err := h.svc.Cancel(c.Request.Context(), id, body.Reason)
	if err != nil {
		writeMappedErr(c, err)
		return
	}
	response.OK(c, runResponse{Job: job})
}

// GetStatus handles GET /api/v1/backtests/:id/status. It returns a
// trimmed projection of the job aggregate so the polling frontend does
// not have to download the full row on every tick.
func (h *Handler) GetStatus(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	job, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		writeMappedErr(c, err)
		return
	}
	response.OK(c, statusView{
		ID:           job.ID,
		Status:       string(job.Status),
		Progress:     clampProgress(job.Progress),
		ErrorMessage: job.ErrorMessage,
		StartedAt:    job.StartedAt,
		FinishedAt:   job.FinishedAt,
	})
}

// clampProgress maps any stored value to the [0,1] band the HTTP
// contract advertises. A stale or partially-written row never produces a
// number a UI cannot render.
func clampProgress(p float64) float64 {
	if p < 0 {
		return 0
	}
	if p > 1 {
		return 1
	}
	return p
}

// shouldWaitInline reports whether the caller asked the API to block on a
// synchronous run via ?wait=true. We accept the legacy ?run=true alias
// because both the e2e harness and the existing http_test rely on it.
func shouldWaitInline(c *gin.Context) bool {
	switch strings.ToLower(c.Query("wait")) {
	case "1", "true", "yes":
		return true
	}
	return false
}

// List handles GET /api/v1/backtests.
func (h *Handler) List(c *gin.Context) {
	q := appBacktest.ListJobsQuery{
		UserID:     parseInt64(c.Query("user_id")),
		StrategyID: parseInt64(c.Query("strategy_id")),
		Status:     valueobject.JobStatus(strings.ToUpper(c.Query("status"))),
		Limit:      int(parseInt64(c.Query("limit"))),
	}
	jobs, err := h.svc.List(c.Request.Context(), q)
	if err != nil {
		writeMappedErr(c, err)
		return
	}
	response.OK(c, gin.H{"items": jobs})
}

// GetReport handles GET /api/v1/backtests/:id/report.
func (h *Handler) GetReport(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	rep, err := h.svc.GetReport(c.Request.Context(), id)
	if err != nil {
		writeMappedErr(c, err)
		return
	}
	response.OK(c, rep)
}

// GetTrades handles GET /api/v1/backtests/:id/trades.
func (h *Handler) GetTrades(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	trades, err := h.svc.GetTrades(c.Request.Context(), id)
	if err != nil {
		writeMappedErr(c, err)
		return
	}
	response.OK(c, gin.H{"items": trades})
}

// GetPositions handles GET /api/v1/backtests/:id/positions.
//
// "Positions" in the API surface means the portfolio snapshots stream, which
// includes the per-stock position list at every trade date. The route name
// follows the TD wording so the API stays predictable for clients that have
// only read the spec.
func (h *Handler) GetPositions(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	snapshots, err := h.svc.GetSnapshots(c.Request.Context(), id)
	if err != nil {
		writeMappedErr(c, err)
		return
	}
	response.OK(c, gin.H{"items": snapshots})
}

// parseID parses :id into int64 and writes a 400 on failure.
func parseID(c *gin.Context) (int64, bool) {
	raw := c.Param("id")
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		response.Error(c, http.StatusBadRequest, errors.New(errors.ErrInvalidParam.Code, "invalid id"))
		return 0, false
	}
	return id, true
}

func parseInt64(s string) int64 {
	if s == "" {
		return 0
	}
	v, _ := strconv.ParseInt(s, 10, 64)
	return v
}

