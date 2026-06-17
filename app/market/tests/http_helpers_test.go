package tests

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	appMarket "github.com/agoXQ/QuantLab/app/market/application/market"
	"github.com/agoXQ/QuantLab/app/market/domain/calendar"
	"github.com/agoXQ/QuantLab/app/market/domain/dataversion"
	"github.com/agoXQ/QuantLab/app/market/domain/factor"
	"github.com/agoXQ/QuantLab/app/market/domain/financial"
	"github.com/agoXQ/QuantLab/app/market/domain/indexbar"
	"github.com/agoXQ/QuantLab/app/market/domain/marketbar"
	"github.com/agoXQ/QuantLab/app/market/domain/security"
	"github.com/agoXQ/QuantLab/app/market/domain/valueobject"
	infraAdj "github.com/agoXQ/QuantLab/app/market/infrastructure/adjustment"
	httpHandler "github.com/agoXQ/QuantLab/app/market/interfaces/http"
)

// httpFixture wires the application service against in-memory fakes and
// mounts the gin handler at /api/v1/market for use across HTTP tests.
type httpFixture struct {
	router   *gin.Engine
	svc      appMarket.Service
	secs     *fakeSecurityRepo
	bars     *fakeBarRepo
	versions *fakeVersionRepo
	cal      *fakeCalendarRepo
	idx      *fakeIndexRepo
	fin      *fakeFinancialRepo
	fac      *fakeFactorRepo
}

func newHTTPFixture() *httpFixture {
	gin.SetMode(gin.TestMode)

	secs := newFakeSecurityRepo()
	bars := newFakeBarRepo()
	versions := &fakeVersionRepo{}
	cal := &fakeCalendarRepo{}
	idx := &fakeIndexRepo{}
	fin := &fakeFinancialRepo{}
	fac := &fakeFactorRepo{}

	deps := appMarket.Dependencies{
		Securities:   secs,
		Bars:         bars,
		Financials:   fin,
		Factors:      fac,
		Indexes:      idx,
		Calendar:     cal,
		DataVersions: versions,
		Adjuster:     infraAdj.NewFactorAdjuster(),
	}
	svc := appMarket.NewService(deps)

	router := gin.New()
	router.Use(gin.Recovery())
	httpHandler.NewHandler(svc).RegisterRoutes(router.Group("/api/v1/market"))

	return &httpFixture{
		router:   router,
		svc:      svc,
		secs:     secs,
		bars:     bars,
		versions: versions,
		cal:      cal,
		idx:      idx,
		fin:      fin,
		fac:      fac,
	}
}

// do executes a request against the fixture's router. body may be empty.
func (f *httpFixture) do(method, path, body string) *httptest.ResponseRecorder {
	var reader *strings.Reader
	if body != "" {
		reader = strings.NewReader(body)
	} else {
		reader = strings.NewReader("")
	}
	req, _ := http.NewRequest(method, path, reader)
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	w := httptest.NewRecorder()
	f.router.ServeHTTP(w, req)
	return w
}

// seedSecurity inserts a Security into the fake repository.
func (f *httpFixture) seedSecurity(code, name string) *security.Security {
	sec := &security.Security{
		StockCode: code,
		StockName: name,
		Market:    valueobject.MarketCN,
		Exchange:  "SSE",
		AssetType: valueobject.AssetTypeStock,
		Status:    valueobject.StatusListed,
	}
	_ = f.secs.Upsert(context.Background(), sec)
	return sec
}

// seedBars inserts a slice of bars into the fake bar repository.
func (f *httpFixture) seedBars(code string, dates []string, closes []float64, factors []float64) {
	bars := make([]*marketbar.MarketBar, 0, len(dates))
	for i, d := range dates {
		t, _ := valueobject.ParseDate(d)
		bars = append(bars, &marketbar.MarketBar{
			StockCode: code,
			Period:    valueobject.PeriodDay,
			TradeDate: t,
			Close:     closes[i],
			AdjFactor: factors[i],
		})
	}
	_ = f.bars.BulkUpsert(context.Background(), bars)
}

// seedVersion creates a DataVersion and returns its name.
func (f *httpFixture) seedVersion(name string) string {
	_ = f.versions.Create(context.Background(), &dataversion.DataVersion{
		Version:   name,
		CreatedAt: time.Now(),
	})
	return name
}

// seedFinancial inserts a financial statement for the given code/date.
func (f *httpFixture) seedFinancial(code, date string, rt valueobject.ReportType, revenue, netProfit float64) {
	t, _ := valueobject.ParseDate(date)
	_ = f.fin.BulkUpsert(context.Background(), []*financial.FinancialStatement{
		{
			StockCode:  code,
			ReportDate: t,
			ReportType: rt,
			Revenue:    revenue,
			NetProfit:  netProfit,
		},
	})
}

// seedFactor inserts a single factor value.
func (f *httpFixture) seedFactor(code, date, name string, value float64) {
	t, _ := valueobject.ParseDate(date)
	_ = f.fac.BulkUpsert(context.Background(), []*factor.Factor{
		{
			StockCode:   code,
			TradeDate:   t,
			FactorName:  name,
			FactorValue: value,
		},
	})
}

// seedIndexBar inserts a single index bar.
func (f *httpFixture) seedIndexBar(code, date string, close float64) {
	t, _ := valueobject.ParseDate(date)
	_ = f.idx.BulkUpsert(context.Background(), []*indexbar.IndexBar{
		{IndexCode: code, TradeDate: t, Close: close},
	})
}

// seedCalendar inserts trading-day rows for the given dates.
func (f *httpFixture) seedCalendar(dates []string, isOpen bool) {
	days := make([]calendar.TradingDay, 0, len(dates))
	for _, d := range dates {
		t, _ := valueobject.ParseDate(d)
		days = append(days, calendar.TradingDay{TradeDate: t, IsOpen: isOpen})
	}
	_ = f.cal.BulkUpsert(context.Background(), days)
}
