// Package http exposes the Market Data application service over a thin HTTP
// surface aligned with the QuantLab API Design Standard.
package http

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	appMarket "github.com/agoXQ/QuantLab/app/market/application/market"
	"github.com/agoXQ/QuantLab/app/market/domain/valueobject"
	"github.com/agoXQ/QuantLab/pkg/errors"
	"github.com/agoXQ/QuantLab/pkg/response"
)

// Handler exposes the Market Data Service over HTTP.
type Handler struct {
	svc appMarket.Service
}

// NewHandler creates a new HTTP handler.
func NewHandler(svc appMarket.Service) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes mounts all market routes under the given group. The group is
// expected to be /api/v1/market.
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/securities", h.ListSecurities)
	rg.GET("/securities/:code", h.GetSecurity)

	rg.GET("/bars", h.GetBars)
	rg.GET("/financials", h.GetFinancials)
	rg.GET("/factors", h.GetFactors)
	rg.GET("/indexes", h.GetIndex)
	rg.GET("/calendar", h.GetCalendar)
	rg.GET("/versions", h.ListVersions)
}

// GetSecurity handles GET /securities/:code.
func (h *Handler) GetSecurity(c *gin.Context) {
	code := c.Param("code")
	sec, err := h.svc.GetSecurity(c.Request.Context(), code)
	if err != nil {
		writeMappedErr(c, err)
		return
	}
	response.OK(c, sec)
}

// ListSecurities handles GET /securities.
func (h *Handler) ListSecurities(c *gin.Context) {
	q := appMarket.ListSecuritiesQuery{
		Market:    valueobject.Market(strings.ToUpper(c.Query("market"))),
		Exchange:  c.Query("exchange"),
		AssetType: valueobject.AssetType(strings.ToUpper(c.Query("asset_type"))),
		Cursor:    c.Query("cursor"),
		Limit:     atoiDefault(c.Query("limit"), 0),
	}
	res, err := h.svc.ListSecurities(c.Request.Context(), q)
	if err != nil {
		writeMappedErr(c, err)
		return
	}
	response.OKWithMeta(c, res.Items, response.PageMeta{NextCursor: res.NextCursor, HasMore: res.HasMore})
}

// GetBars handles GET /bars.
func (h *Handler) GetBars(c *gin.Context) {
	period, err := valueobject.ParsePeriod(c.Query("period"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, errors.ErrInvalidParam)
		return
	}
	mode, err := valueobject.ParseAdjustment(c.Query("adjustment"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, errors.ErrInvalidParam)
		return
	}
	rng, err := buildRange(c)
	if err != nil {
		response.Error(c, http.StatusBadRequest, errors.ErrInvalidParam)
		return
	}
	res, err := h.svc.GetBars(c.Request.Context(), appMarket.GetBarsQuery{
		StockCode:   c.Query("code"),
		Period:      period,
		Adjustment:  mode,
		Range:       rng,
		DataVersion: c.Query("data_version"),
		Limit:       atoiDefault(c.Query("limit"), 0),
	})
	if err != nil {
		writeMappedErr(c, err)
		return
	}
	response.OK(c, res)
}

// GetFinancials handles GET /financials.
func (h *Handler) GetFinancials(c *gin.Context) {
	rng, err := buildRange(c)
	if err != nil {
		response.Error(c, http.StatusBadRequest, errors.ErrInvalidParam)
		return
	}
	res, err := h.svc.GetFinancials(c.Request.Context(), appMarket.GetFinancialsQuery{
		StockCode:   c.Query("code"),
		ReportType:  valueobject.ReportType(strings.ToLower(c.Query("report_type"))),
		Range:       rng,
		DataVersion: c.Query("data_version"),
		Cursor:      c.Query("cursor"),
		Limit:       atoiDefault(c.Query("limit"), 0),
	})
	if err != nil {
		writeMappedErr(c, err)
		return
	}
	response.OKWithMeta(c, res.Items, response.PageMeta{NextCursor: res.NextCursor, HasMore: res.HasMore})
}

// GetFactors handles GET /factors.
func (h *Handler) GetFactors(c *gin.Context) {
	rng, err := buildRange(c)
	if err != nil {
		response.Error(c, http.StatusBadRequest, errors.ErrInvalidParam)
		return
	}
	res, err := h.svc.GetFactors(c.Request.Context(), appMarket.GetFactorsQuery{
		StockCode:   c.Query("code"),
		FactorNames: splitCSV(c.Query("factors")),
		Range:       rng,
		DataVersion: c.Query("data_version"),
		Limit:       atoiDefault(c.Query("limit"), 0),
	})
	if err != nil {
		writeMappedErr(c, err)
		return
	}
	response.OK(c, res)
}

// GetIndex handles GET /indexes.
func (h *Handler) GetIndex(c *gin.Context) {
	rng, err := buildRange(c)
	if err != nil {
		response.Error(c, http.StatusBadRequest, errors.ErrInvalidParam)
		return
	}
	res, err := h.svc.GetIndex(c.Request.Context(), appMarket.GetIndexQuery{
		IndexCode:   c.Query("code"),
		Range:       rng,
		DataVersion: c.Query("data_version"),
	})
	if err != nil {
		writeMappedErr(c, err)
		return
	}
	response.OK(c, res)
}

// GetCalendar handles GET /calendar.
func (h *Handler) GetCalendar(c *gin.Context) {
	rng, err := buildRange(c)
	if err != nil {
		response.Error(c, http.StatusBadRequest, errors.ErrInvalidParam)
		return
	}
	res, err := h.svc.GetCalendar(c.Request.Context(), appMarket.CalendarQuery{Range: rng})
	if err != nil {
		writeMappedErr(c, err)
		return
	}
	response.OK(c, res.Days)
}

// ListVersions handles GET /versions.
func (h *Handler) ListVersions(c *gin.Context) {
	res, err := h.svc.ListVersions(c.Request.Context(), atoiDefault(c.Query("limit"), 0))
	if err != nil {
		writeMappedErr(c, err)
		return
	}
	response.OK(c, res.Items)
}

// --- helpers ---

func statusForCode(code int) int {
	switch {
	case code >= 10000 && code < 20000:
		return http.StatusBadRequest
	case code >= 20000 && code < 30000:
		return http.StatusUnprocessableEntity
	case code >= 30000 && code < 40000:
		return http.StatusForbidden
	case code >= 40000 && code < 50000:
		return http.StatusNotFound
	default:
		return http.StatusInternalServerError
	}
}

func atoiDefault(raw string, def int) int {
	if raw == "" {
		return def
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		return def
	}
	return n
}

func splitCSV(raw string) []string {
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := parts[:0]
	for _, p := range parts {
		if v := strings.TrimSpace(p); v != "" {
			out = append(out, v)
		}
	}
	return out
}

func buildRange(c *gin.Context) (valueobject.DateRange, error) {
	var rng valueobject.DateRange
	if start := c.Query("start"); start != "" {
		t, err := valueobject.ParseDate(start)
		if err != nil {
			return rng, err
		}
		rng.Start = t
	}
	if end := c.Query("end"); end != "" {
		t, err := valueobject.ParseDate(end)
		if err != nil {
			return rng, err
		}
		rng.End = t
	}
	return rng, nil
}
