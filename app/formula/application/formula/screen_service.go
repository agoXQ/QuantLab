package formula

import (
	"context"
	"fmt"
	"strings"
	"time"

	domainCompiler "github.com/agoXQ/QuantLab/app/formula/domain/compiler"
	domainEval "github.com/agoXQ/QuantLab/app/formula/domain/evaluator"
	"github.com/agoXQ/QuantLab/app/market/domain/security"
	"github.com/agoXQ/QuantLab/app/market/domain/valueobject"
)

type SecuritySource interface {
	GetSecurity(ctx context.Context, stockCode string) (*security.Security, error)
	ListSecurities(ctx context.Context, q security.ListQuery) ([]*security.Security, string, error)
	LatestDataVersion(ctx context.Context) (string, error)
}

type ScreenService interface {
	Screen(ctx context.Context, req ScreenRequest) (*ScreenResult, error)
}

type UniverseFilter struct {
	Market     string
	Exchange   string
	Industry   string
	AssetType  string
	Status     string
	StockCodes []string
}

type ScreenRequest struct {
	Formula        string
	AsOfDate       time.Time
	LookbackBars   int
	DataVersion    string
	Limit          int
	UniverseFilter UniverseFilter
}

type ScreenResult struct {
	FormulaHash  string
	PlanType     string
	DataVersion  string
	UniverseSize int
	Items        []ScreenItem
}

type ScreenItem struct {
	StockCode string
	StockName string
	Exchange  string
	Industry  string
	Score     *float64
	Selected  bool
}

type screenService struct {
	evaluator EvaluatorService
	source    SecuritySource
	dataPort  domainEval.DataPort
}

func NewScreenService(evaluator EvaluatorService, source SecuritySource, dataPort domainEval.DataPort) ScreenService {
	if evaluator == nil || source == nil || dataPort == nil {
		return nil
	}
	return &screenService{evaluator: evaluator, source: source, dataPort: dataPort}
}

func (s *screenService) Screen(ctx context.Context, req ScreenRequest) (*ScreenResult, error) {
	if req.Formula == "" {
		return nil, fmt.Errorf("公式不能为空")
	}
	limit := normalizeScreenResultLimit(req.Limit)
	filter := normalizeUniverseFilter(req.UniverseFilter)
	securities, err := s.loadUniverse(ctx, filter)
	if err != nil {
		return nil, err
	}
	if len(securities) == 0 {
		return nil, fmt.Errorf("股票池为空，请检查交易所、行业或手工股票代码条件")
	}
	version := req.DataVersion
	if version == "" {
		version, err = s.source.LatestDataVersion(ctx)
		if err != nil {
			return nil, fmt.Errorf("获取最新数据版本失败: %w", err)
		}
		if version == "" {
			return nil, fmt.Errorf("暂无可用的数据版本，请先导入 Market Data 数据")
		}
	}
	codes := make([]string, 0, len(securities))
	meta := make(map[string]*security.Security, len(securities))
	for _, sec := range securities {
		codes = append(codes, sec.StockCode)
		meta[sec.StockCode] = sec
	}
	eval, err := s.evaluator.Evaluate(ctx, EvaluateRequest{
		Formula:      req.Formula,
		Universe:     codes,
		AsOfDate:     req.AsOfDate,
		LookbackBars: req.LookbackBars,
		DataVersion:  version,
		DataPort:     s.dataPort,
	})
	if err != nil {
		return nil, fmt.Errorf("执行公式选股失败: %w", err)
	}
	return &ScreenResult{
		FormulaHash:  eval.FormulaHash,
		PlanType:     planTypeOf(eval),
		DataVersion:  version,
		UniverseSize: len(securities),
		Items:        limitScreenItems(buildScreenItems(eval, meta), limit),
	}, nil
}

func (s *screenService) loadUniverse(ctx context.Context, filter UniverseFilter) ([]*security.Security, error) {
	if len(filter.StockCodes) > 0 {
		return s.loadManualUniverse(ctx, filter)
	}
	query := securityQueryFromFilter(filter, screenUniversePageSize)
	out := make([]*security.Security, 0, screenUniversePageSize)
	for {
		securities, next, err := s.source.ListSecurities(ctx, query)
		if err != nil {
			return nil, fmt.Errorf("查询股票池失败: %w", err)
		}
		for _, sec := range securities {
			if !securityMatchesFilter(sec, filter) {
				continue
			}
			out = append(out, sec)
			if len(out) >= maxScreenUniverseSize {
				return out, nil
			}
		}
		if next == "" {
			break
		}
		query.Cursor = next
	}
	return out, nil
}

func (s *screenService) loadManualUniverse(ctx context.Context, filter UniverseFilter) ([]*security.Security, error) {
	out := make([]*security.Security, 0, len(filter.StockCodes))
	seen := map[string]struct{}{}
	for _, raw := range filter.StockCodes {
		if len(out) >= maxScreenUniverseSize {
			break
		}
		code := strings.ToUpper(strings.TrimSpace(raw))
		if code == "" {
			continue
		}
		if _, ok := seen[code]; ok {
			continue
		}
		seen[code] = struct{}{}
		sec, err := s.source.GetSecurity(ctx, code)
		if err != nil || !securityMatchesFilter(sec, filter) {
			continue
		}
		out = append(out, sec)
	}
	return out, nil
}

func normalizeUniverseFilter(f UniverseFilter) UniverseFilter {
	f.Market = strings.ToUpper(strings.TrimSpace(f.Market))
	f.Exchange = strings.ToUpper(strings.TrimSpace(f.Exchange))
	f.Industry = strings.TrimSpace(f.Industry)
	f.AssetType = strings.ToUpper(strings.TrimSpace(f.AssetType))
	f.Status = strings.ToUpper(strings.TrimSpace(f.Status))
	if f.AssetType == "" {
		f.AssetType = string(valueobject.AssetTypeStock)
	}
	if f.Status == "" {
		f.Status = string(valueobject.StatusListed)
	}
	return f
}

func normalizeScreenResultLimit(limit int) int {
	if limit <= 0 {
		return 500
	}
	if limit > 2000 {
		return 2000
	}
	return limit
}

const (
	screenUniversePageSize = 500
	maxScreenUniverseSize  = 10000
)

func securityQueryFromFilter(f UniverseFilter, limit int) security.ListQuery {
	assetType := valueobject.AssetType(f.AssetType)
	market := valueobject.Market(f.Market)
	return security.ListQuery{
		Market:    market,
		Exchange:  f.Exchange,
		AssetType: assetType,
		Industry:  f.Industry,
		Status:    valueobject.SecurityStatus(f.Status),
		Limit:     limit,
	}
}

func limitScreenItems(items []ScreenItem, limit int) []ScreenItem {
	if limit <= 0 || len(items) <= limit {
		return items
	}
	return items[:limit]
}

func securityMatchesFilter(sec *security.Security, f UniverseFilter) bool {
	if sec == nil {
		return false
	}
	if f.Industry != "" && sec.Industry != f.Industry {
		return false
	}
	if f.Exchange != "" && sec.Exchange != f.Exchange {
		return false
	}
	if f.Market != "" && string(sec.Market) != f.Market {
		return false
	}
	if f.AssetType != "" && string(sec.AssetType) != f.AssetType {
		return false
	}
	if f.Status != "" && string(sec.Status) != f.Status {
		return false
	}
	return true
}

func planTypeOf(eval *EvaluateResult) string {
	if eval == nil || eval.Result == nil {
		return ""
	}
	return string(eval.Result.PlanType)
}

func buildScreenItems(eval *EvaluateResult, meta map[string]*security.Security) []ScreenItem {
	if eval == nil || eval.Result == nil {
		return nil
	}
	switch eval.Result.PlanType {
	case domainCompiler.PlanTypeFilter, domainCompiler.PlanTypeSignal:
		if eval.Result.Selection == nil {
			return nil
		}
		items := make([]ScreenItem, 0, len(eval.Result.Selection.StockCodes))
		for _, code := range eval.Result.Selection.StockCodes {
			items = append(items, screenItem(code, meta, nil, true))
		}
		return items
	case domainCompiler.PlanTypeSort:
		if eval.Result.Ranking == nil {
			return nil
		}
		items := make([]ScreenItem, 0, len(eval.Result.Ranking.StockCodes))
		for i, code := range eval.Result.Ranking.StockCodes {
			score := eval.Result.Ranking.Scores[i]
			items = append(items, screenItem(code, meta, &score, true))
		}
		return items
	case domainCompiler.PlanTypeValue:
		if eval.Result.Values == nil {
			return nil
		}
		items := make([]ScreenItem, 0, len(eval.Result.Values.StockCodes))
		for i, code := range eval.Result.Values.StockCodes {
			score := eval.Result.Values.Values[i]
			items = append(items, screenItem(code, meta, &score, true))
		}
		return items
	default:
		return nil
	}
}

func screenItem(code string, meta map[string]*security.Security, score *float64, selected bool) ScreenItem {
	sec := meta[code]
	item := ScreenItem{StockCode: code, Score: score, Selected: selected}
	if sec != nil {
		item.StockName = sec.StockName
		item.Exchange = sec.Exchange
		item.Industry = sec.Industry
	}
	return item
}
