// Package http exposes the Strategy Service application service over a
// thin HTTP surface aligned with the QuantLab API Design Standard.
//
// The handler depth-1 matches the Backtest implementation: one Service
// dependency, one RegisterRoutes that accepts the api group, and a small
// error-mapping helper that turns domain errors into the platform's
// AppError envelope. New endpoints belong here rather than in the
// gRPC logic stubs so the HTTP surface stays the canonical client API.
package http

import (
	"github.com/gin-gonic/gin"

	appStrategy "github.com/agoXQ/QuantLab/app/strategy/application/strategy"
)

// Handler exposes the Strategy Service over HTTP.
type Handler struct {
	svc appStrategy.Service
}

// NewHandler returns a Handler bound to the given application service.
func NewHandler(svc appStrategy.Service) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes mounts every route under the supplied group, expected
// to be /api/v1/strategies.
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("", h.Create)
	rg.GET("", h.List)
	rg.GET(":id", h.Get)
	rg.PUT(":id", h.Update)
	rg.DELETE(":id", h.Delete)

	rg.POST(":id/versions", h.CreateVersion)
	rg.GET(":id/versions", h.ListVersions)
	rg.GET("versions/:vid", h.GetVersion)

	rg.POST(":id/publish", h.Publish)
	rg.POST(":id/archive", h.Archive)
	rg.POST(":id/fork", h.Fork)
}

// --- request shapes ---

type createStrategyRequest struct {
	AuthorID    int64    `json:"author_id"`
	Title       string   `json:"title" binding:"required"`
	Description string   `json:"description"`
	Category    string   `json:"category"`
	Tags        []string `json:"tags"`
	Visibility  string   `json:"visibility"`
}

type updateStrategyRequest struct {
	CallerID    int64     `json:"caller_id"`
	Title       *string   `json:"title"`
	Description *string   `json:"description"`
	Category    *string   `json:"category"`
	Tags        *[]string `json:"tags"`
	Visibility  *string   `json:"visibility"`
}

type createVersionRequest struct {
	CallerID      int64  `json:"caller_id"`
	FormulaText   string `json:"formula_text" binding:"required"`
	BuyRule       string `json:"buy_rule"`
	SellRule      string `json:"sell_rule"`
	RiskRule      string `json:"risk_rule"`
	PositionRule  string `json:"position_rule"`
	RebalanceRule string `json:"rebalance_rule"`
	ChangeLog     string `json:"change_log"`
}

type publishRequest struct {
	CallerID  int64 `json:"caller_id"`
	VersionID int64 `json:"version_id"`
}

type archiveRequest struct {
	CallerID int64 `json:"caller_id"`
}

type forkRequest struct {
	CallerID int64  `json:"caller_id"`
	Title    string `json:"title"`
}
