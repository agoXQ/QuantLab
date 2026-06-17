package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/agoXQ/QuantLab/app/market/domain/security"
)

// envelope mirrors the unified response envelope used by the HTTP layer.
type envelope[T any] struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    T      `json:"data"`
	Meta    struct {
		NextCursor string `json:"next_cursor"`
		HasMore    bool   `json:"has_more"`
	} `json:"meta"`
}

func decodeEnvelope[T any](t *testing.T, body []byte) envelope[T] {
	t.Helper()
	var env envelope[T]
	if err := json.Unmarshal(body, &env); err != nil {
		t.Fatalf("decode response: %v\nbody=%s", err, string(body))
	}
	return env
}

func TestHTTP_GetSecurityOK(t *testing.T) {
	f := newHTTPFixture()
	f.seedSecurity("600519", "Moutai")

	w := f.do(http.MethodGet, "/api/v1/market/securities/600519", "")
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body=%s", w.Code, w.Body.String())
	}
	resp := decodeEnvelope[security.Security](t, w.Body.Bytes())
	if resp.Code != 0 {
		t.Fatalf("expected envelope code=0, got %d", resp.Code)
	}
	if resp.Data.StockCode != "600519" {
		t.Fatalf("expected stock_code=600519, got %s", resp.Data.StockCode)
	}
	if resp.Data.StockName != "Moutai" {
		t.Fatalf("expected stock_name=Moutai, got %s", resp.Data.StockName)
	}
}

func TestHTTP_GetSecurityNotFound(t *testing.T) {
	f := newHTTPFixture()

	w := f.do(http.MethodGet, "/api/v1/market/securities/999999", "")
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d, body=%s", w.Code, w.Body.String())
	}
	resp := decodeEnvelope[any](t, w.Body.Bytes())
	if resp.Code != 40010 {
		t.Fatalf("expected code=40010 (security not found), got %d", resp.Code)
	}
}

func TestHTTP_ListSecurities(t *testing.T) {
	f := newHTTPFixture()
	f.seedSecurity("000001", "PAB")
	f.seedSecurity("600519", "Moutai")

	w := f.do(http.MethodGet, "/api/v1/market/securities?market=CN", "")
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body=%s", w.Code, w.Body.String())
	}
	resp := decodeEnvelope[[]security.Security](t, w.Body.Bytes())
	if len(resp.Data) != 2 {
		t.Fatalf("expected 2 securities, got %d", len(resp.Data))
	}
}

func TestHTTP_GetBarsApplyAdjustment(t *testing.T) {
	f := newHTTPFixture()
	f.seedSecurity("600519", "Moutai")
	f.seedBars("600519",
		[]string{"2026-01-02", "2026-01-03"},
		[]float64{100, 200},
		[]float64{1, 2},
	)

	w := f.do(http.MethodGet,
		"/api/v1/market/bars?code=600519&start=2026-01-01&end=2026-01-31",
		"",
	)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body=%s", w.Code, w.Body.String())
	}

	type barWire struct {
		Items []struct {
			TradeDate string  `json:"trade_date"`
			Close     float64 `json:"close"`
			AdjFactor float64 `json:"adj_factor"`
		} `json:"Items"`
		Adjustment string `json:"Adjustment"`
	}
	resp := decodeEnvelope[barWire](t, w.Body.Bytes())
	if resp.Code != 0 {
		t.Fatalf("expected envelope code=0, got %d", resp.Code)
	}
	if len(resp.Data.Items) != 2 {
		t.Fatalf("expected 2 bars, got %d", len(resp.Data.Items))
	}
	// Forward adjustment should rescale the older bar by 1/2.
	if got := resp.Data.Items[0].Close; got >= 100 {
		t.Fatalf("expected forward-adjusted close < 100, got %v", got)
	}
	if got := resp.Data.Items[1].Close; got != 200 {
		t.Fatalf("expected latest bar close = 200, got %v", got)
	}
	if resp.Data.Adjustment != "pre" {
		t.Fatalf("expected default adjustment pre, got %s", resp.Data.Adjustment)
	}
}

func TestHTTP_GetBarsInvalidPeriod(t *testing.T) {
	f := newHTTPFixture()
	f.seedSecurity("600519", "Moutai")

	w := f.do(http.MethodGet,
		"/api/v1/market/bars?code=600519&period=tick",
		"",
	)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d, body=%s", w.Code, w.Body.String())
	}
	resp := decodeEnvelope[any](t, w.Body.Bytes())
	if resp.Code == 0 {
		t.Fatalf("expected non-zero error code, got %+v", resp)
	}
}

func TestHTTP_GetBarsInvalidDate(t *testing.T) {
	f := newHTTPFixture()
	w := f.do(http.MethodGet,
		"/api/v1/market/bars?code=600519&start=not-a-date",
		"",
	)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d, body=%s", w.Code, w.Body.String())
	}
}

func TestHTTP_GetCalendar(t *testing.T) {
	f := newHTTPFixture()
	w := f.do(http.MethodGet,
		"/api/v1/market/calendar?start=2026-01-01&end=2026-01-31",
		"",
	)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body=%s", w.Code, w.Body.String())
	}
	resp := decodeEnvelope[[]any](t, w.Body.Bytes())
	if resp.Code != 0 {
		t.Fatalf("expected envelope code=0, got %d", resp.Code)
	}
	// No data seeded, but the route must still respond with a JSON envelope.
	if resp.Data == nil {
		t.Fatal("expected non-nil data slice")
	}
}

func TestHTTP_ListVersions(t *testing.T) {
	f := newHTTPFixture()
	f.seedVersion("2026.01.02")
	f.seedVersion("2026.02.01")

	w := f.do(http.MethodGet, "/api/v1/market/versions", "")
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body=%s", w.Code, w.Body.String())
	}
	resp := decodeEnvelope[[]struct {
		Version string `json:"version"`
	}](t, w.Body.Bytes())
	if len(resp.Data) != 2 {
		t.Fatalf("expected 2 versions, got %d", len(resp.Data))
	}
}

func TestHTTP_HealthEnvelope(t *testing.T) {
	// The HTTP layer doesn't expose /healthz directly; this guard ensures the
	// envelope decoder helper itself stays honest.
	body := []byte(fmt.Sprintf(`{"code":0,"message":"ok","data":%q}`, "pong"))
	resp := decodeEnvelope[string](&testing.T{}, body)
	if resp.Data != "pong" {
		t.Fatalf("decodeEnvelope helper broken: %s", resp.Data)
	}
}

func TestHTTP_GetFinancials(t *testing.T) {
	f := newHTTPFixture()
	f.seedSecurity("600519", "Moutai")
	f.seedFinancial("600519", "2026-03-31", "q1", 1000, 200)
	f.seedFinancial("600519", "2026-06-30", "interim", 2500, 600)

	w := f.do(http.MethodGet,
		"/api/v1/market/financials?code=600519&start=2026-01-01&end=2026-12-31",
		"",
	)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body=%s", w.Code, w.Body.String())
	}
	resp := decodeEnvelope[[]struct {
		StockCode  string  `json:"stock_code"`
		ReportType string  `json:"report_type"`
		Revenue    float64 `json:"revenue"`
	}](t, w.Body.Bytes())
	if resp.Code != 0 {
		t.Fatalf("expected envelope code=0, got %d", resp.Code)
	}
	if len(resp.Data) != 2 {
		t.Fatalf("expected 2 financials, got %d", len(resp.Data))
	}
	if resp.Data[0].StockCode != "600519" {
		t.Fatalf("unexpected stock_code: %s", resp.Data[0].StockCode)
	}
}

func TestHTTP_GetFinancialsRequiresStockCode(t *testing.T) {
	f := newHTTPFixture()

	w := f.do(http.MethodGet, "/api/v1/market/financials", "")
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for empty code, got %d, body=%s", w.Code, w.Body.String())
	}
	resp := decodeEnvelope[any](t, w.Body.Bytes())
	if resp.Code != 10010 {
		t.Fatalf("expected code=10010 (invalid stock code), got %d", resp.Code)
	}
}

func TestHTTP_GetFinancialsInvalidReportType(t *testing.T) {
	f := newHTTPFixture()
	f.seedSecurity("600519", "Moutai")

	w := f.do(http.MethodGet,
		"/api/v1/market/financials?code=600519&report_type=bogus",
		"",
	)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid report_type, got %d, body=%s", w.Code, w.Body.String())
	}
	resp := decodeEnvelope[any](t, w.Body.Bytes())
	if resp.Code != 10014 {
		t.Fatalf("expected code=10014 (invalid report type), got %d", resp.Code)
	}
}

func TestHTTP_GetFactors(t *testing.T) {
	f := newHTTPFixture()
	f.seedSecurity("600519", "Moutai")
	f.seedFactor("600519", "2026-01-02", "pe_ttm", 22.5)
	f.seedFactor("600519", "2026-01-02", "pb", 8.0)

	w := f.do(http.MethodGet,
		"/api/v1/market/factors?code=600519&factors=pe_ttm,pb",
		"",
	)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body=%s", w.Code, w.Body.String())
	}
	resp := decodeEnvelope[struct {
		Items []struct {
			FactorName  string  `json:"factor_name"`
			FactorValue float64 `json:"factor_value"`
		} `json:"Items"`
	}](t, w.Body.Bytes())
	if len(resp.Data.Items) != 2 {
		t.Fatalf("expected 2 factors, got %d", len(resp.Data.Items))
	}
}

func TestHTTP_GetFactorsRequiresStockCode(t *testing.T) {
	f := newHTTPFixture()

	w := f.do(http.MethodGet, "/api/v1/market/factors", "")
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d, body=%s", w.Code, w.Body.String())
	}
	resp := decodeEnvelope[any](t, w.Body.Bytes())
	if resp.Code != 10010 {
		t.Fatalf("expected code=10010, got %d", resp.Code)
	}
}

func TestHTTP_GetIndex(t *testing.T) {
	f := newHTTPFixture()
	f.seedIndexBar("000300.SH", "2026-01-02", 4321.0)

	w := f.do(http.MethodGet,
		"/api/v1/market/indexes?code=000300.SH&start=2026-01-01&end=2026-01-31",
		"",
	)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body=%s", w.Code, w.Body.String())
	}
	resp := decodeEnvelope[struct {
		Items []struct {
			IndexCode string  `json:"index_code"`
			Close     float64 `json:"close"`
		} `json:"Items"`
	}](t, w.Body.Bytes())
	if len(resp.Data.Items) != 1 {
		t.Fatalf("expected 1 index bar, got %d", len(resp.Data.Items))
	}
	if resp.Data.Items[0].Close != 4321.0 {
		t.Fatalf("unexpected close: %v", resp.Data.Items[0].Close)
	}
}

func TestHTTP_GetIndexRequiresCode(t *testing.T) {
	f := newHTTPFixture()

	w := f.do(http.MethodGet, "/api/v1/market/indexes", "")
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for empty index code, got %d, body=%s", w.Code, w.Body.String())
	}
	resp := decodeEnvelope[any](t, w.Body.Bytes())
	if resp.Code != 40014 {
		t.Fatalf("expected code=40014 (index not found), got %d", resp.Code)
	}
}

func TestHTTP_GetCalendarReturnsSeededDays(t *testing.T) {
	f := newHTTPFixture()
	f.seedCalendar([]string{"2026-01-02", "2026-01-03"}, true)

	w := f.do(http.MethodGet,
		"/api/v1/market/calendar?start=2026-01-01&end=2026-01-31",
		"",
	)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body=%s", w.Code, w.Body.String())
	}
	resp := decodeEnvelope[[]struct {
		TradeDate string `json:"trade_date"`
		IsOpen    bool   `json:"is_open"`
	}](t, w.Body.Bytes())
	if len(resp.Data) != 2 {
		t.Fatalf("expected 2 trading days, got %d", len(resp.Data))
	}
	if !resp.Data[0].IsOpen {
		t.Fatalf("expected is_open=true, got %v", resp.Data[0].IsOpen)
	}
}

func TestHTTP_GetCalendarInvalidRange(t *testing.T) {
	f := newHTTPFixture()

	w := f.do(http.MethodGet,
		"/api/v1/market/calendar?start=2026-02-01&end=2026-01-01",
		"",
	)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for inverted range, got %d, body=%s", w.Code, w.Body.String())
	}
	resp := decodeEnvelope[any](t, w.Body.Bytes())
	if resp.Code != 10013 {
		t.Fatalf("expected code=10013 (invalid date range), got %d", resp.Code)
	}
}

func TestHTTP_ListSecuritiesPaginationMeta(t *testing.T) {
	f := newHTTPFixture()
	f.seedSecurity("600519", "Moutai")

	w := f.do(http.MethodGet, "/api/v1/market/securities?limit=10", "")
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body=%s", w.Code, w.Body.String())
	}
	resp := decodeEnvelope[[]security.Security](t, w.Body.Bytes())
	if resp.Meta.HasMore {
		t.Fatalf("expected has_more=false, got true")
	}
	if resp.Meta.NextCursor != "" {
		t.Fatalf("expected empty next_cursor, got %q", resp.Meta.NextCursor)
	}
}

func TestHTTP_GetBarsResolvesRequestedVersionMissing(t *testing.T) {
	f := newHTTPFixture()
	f.seedSecurity("600519", "Moutai")

	w := f.do(http.MethodGet,
		"/api/v1/market/bars?code=600519&data_version=2099.99.99",
		"",
	)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for missing version, got %d, body=%s", w.Code, w.Body.String())
	}
	resp := decodeEnvelope[any](t, w.Body.Bytes())
	if resp.Code != 40012 {
		t.Fatalf("expected code=40012 (data version not found), got %d", resp.Code)
	}
}

func TestHTTP_UnknownRouteReturns404(t *testing.T) {
	f := newHTTPFixture()

	w := f.do(http.MethodGet, "/api/v1/market/does-not-exist", "")
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for unknown route, got %d", w.Code)
	}
}
