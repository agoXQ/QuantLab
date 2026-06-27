package handler

import (
	"sort"

	"github.com/gin-gonic/gin"

	commonpb "github.com/agoXQ/QuantLab/api/common/v1"
	marketpb "github.com/agoXQ/QuantLab/app/market/pb"
	"github.com/agoXQ/QuantLab/pkg/response"
)

// registerMarkets mounts /api/v1/markets. Market data is read-only and
// public so the frontend can render charts without auth.
func (h *Handler) registerMarkets(rg *gin.RouterGroup) {
	m := rg.Group("/markets")
	m.GET("/securities/:stock_code", h.marketGetSecurity)
	m.GET("/securities", h.marketListSecurities)
	m.GET("/exchanges", h.marketListExchanges)
	m.GET("/industries", h.marketListIndustries)
	m.GET("/bars", h.marketGetBars)
	m.GET("/financials", h.marketGetFinancials)
	m.GET("/factors", h.marketGetFactors)
	m.GET("/index", h.marketGetIndex)
	m.GET("/calendar", h.marketGetCalendar)
	m.GET("/versions", h.marketGetVersions)
}

func (h *Handler) marketGetSecurity(c *gin.Context) {
	code := c.Param("stock_code")
	out, err := h.svc.Market.GetSecurity(c.Request.Context(), &marketpb.GetSecurityRequest{StockCode: code})
	if err != nil {
		grpcErr(c, err)
		return
	}
	response.OK(c, gin.H{"security": out.Security})
}

func (h *Handler) marketListSecurities(c *gin.Context) {
	out, err := h.svc.Market.ListSecurities(c.Request.Context(), &marketpb.ListSecuritiesRequest{
		Cursor: &commonpb.Cursor{NextCursor: queryCursor(c)}, Limit: queryLimit(c),
	})
	if err != nil {
		grpcErr(c, err)
		return
	}
	response.OKWithMeta(c, gin.H{"items": out.Securities}, out.Cursor)
}

func (h *Handler) marketListIndustries(c *gin.Context) {
	h.marketListSecurityAttribute(c, func(sec *marketpb.Security) string { return sec.Industry })
}

func (h *Handler) marketListExchanges(c *gin.Context) {
	h.marketListSecurityAttribute(c, func(sec *marketpb.Security) string { return sec.Exchange })
}

func (h *Handler) marketListSecurityAttribute(c *gin.Context, pick func(*marketpb.Security) string) {
	seen := make(map[string]struct{})
	cursor := ""
	for page := 0; page < 100; page++ {
		out, err := h.svc.Market.ListSecurities(c.Request.Context(), &marketpb.ListSecuritiesRequest{
			Cursor: &commonpb.Cursor{NextCursor: cursor}, Limit: 100,
		})
		if err != nil {
			grpcErr(c, err)
			return
		}
		for _, sec := range out.Securities {
			if sec == nil {
				continue
			}
			value := pick(sec)
			if value == "" {
				continue
			}
			seen[value] = struct{}{}
		}
		if out.Cursor == nil || !out.Cursor.HasMore || out.Cursor.NextCursor == "" {
			break
		}
		cursor = out.Cursor.NextCursor
	}
	items := make([]string, 0, len(seen))
	for industry := range seen {
		items = append(items, industry)
	}
	sort.Strings(items)
	response.OK(c, gin.H{"items": items})
}

func (h *Handler) marketGetBars(c *gin.Context) {
	out, err := h.svc.Market.GetBars(c.Request.Context(), &marketpb.GetBarsRequest{
		StockCode:  c.Query("stock_code"),
		StartDate:  c.Query("start_date"),
		EndDate:    c.Query("end_date"),
		Period:     c.Query("period"),
		Adjustment: c.Query("adjustment"),
		Cursor:     &commonpb.Cursor{NextCursor: queryCursor(c)},
		Limit:      queryLimitWithMax(c, 1000, 10000),
	})
	if err != nil {
		grpcErr(c, err)
		return
	}
	response.OKWithMeta(c, gin.H{"items": out.Bars}, out.Cursor)
}

func (h *Handler) marketGetFinancials(c *gin.Context) {
	out, err := h.svc.Market.GetFinancials(c.Request.Context(), &marketpb.GetFinancialsRequest{
		StockCode:  c.Query("stock_code"),
		ReportType: c.Query("report_type"),
		Cursor:     &commonpb.Cursor{NextCursor: queryCursor(c)},
		Limit:      queryLimit(c),
	})
	if err != nil {
		grpcErr(c, err)
		return
	}
	response.OKWithMeta(c, gin.H{"items": out.Statements}, out.Cursor)
}

func (h *Handler) marketGetFactors(c *gin.Context) {
	out, err := h.svc.Market.GetFactors(c.Request.Context(), &marketpb.GetFactorsRequest{
		StockCode: c.Query("stock_code"),
		StartDate: c.Query("start_date"),
		EndDate:   c.Query("end_date"),
		Cursor:    &commonpb.Cursor{NextCursor: queryCursor(c)},
		Limit:     queryLimitWithMax(c, 1000, 10000),
	})
	if err != nil {
		grpcErr(c, err)
		return
	}
	response.OKWithMeta(c, gin.H{"items": out.Factors}, out.Cursor)
}

func (h *Handler) marketGetIndex(c *gin.Context) {
	out, err := h.svc.Market.GetIndex(c.Request.Context(), &marketpb.GetIndexRequest{
		IndexCode: c.Query("index_code"),
		StartDate: c.Query("start_date"),
		EndDate:   c.Query("end_date"),
	})
	if err != nil {
		grpcErr(c, err)
		return
	}
	response.OK(c, gin.H{"items": out.Bars})
}

func (h *Handler) marketGetCalendar(c *gin.Context) {
	out, err := h.svc.Market.GetCalendar(c.Request.Context(), &marketpb.GetCalendarRequest{
		StartDate: c.Query("start_date"),
		EndDate:   c.Query("end_date"),
	})
	if err != nil {
		grpcErr(c, err)
		return
	}
	response.OK(c, gin.H{"items": out.Days})
}

func (h *Handler) marketGetVersions(c *gin.Context) {
	out, err := h.svc.Market.GetVersions(c.Request.Context(), &marketpb.GetVersionsRequest{})
	if err != nil {
		grpcErr(c, err)
		return
	}
	response.OK(c, gin.H{"items": out.Versions})
}

// guard against unused import if http ever drops out.
