// Package grpc adapts the application Service to the generated gRPC layer.
package grpc

import (
	appMarket "github.com/agoXQ/QuantLab/app/market/application/market"
	"github.com/agoXQ/QuantLab/app/market/domain/calendar"
	"github.com/agoXQ/QuantLab/app/market/domain/dataversion"
	"github.com/agoXQ/QuantLab/app/market/domain/factor"
	"github.com/agoXQ/QuantLab/app/market/domain/financial"
	"github.com/agoXQ/QuantLab/app/market/domain/indexbar"
	"github.com/agoXQ/QuantLab/app/market/domain/marketbar"
	"github.com/agoXQ/QuantLab/app/market/domain/security"
	"github.com/agoXQ/QuantLab/app/market/domain/valueobject"
	"github.com/agoXQ/QuantLab/app/market/pb"
	commonv1 "github.com/agoXQ/QuantLab/api/common/v1"
)

// SecurityToPB maps a domain Security to the wire format.
func SecurityToPB(s *security.Security) *pb.Security {
	if s == nil {
		return nil
	}
	return &pb.Security{
		Id:            s.ID,
		StockCode:     s.StockCode,
		StockName:     s.StockName,
		Exchange:      s.Exchange,
		Industry:      s.Industry,
		ListingDate:   valueobject.FormatDate(s.ListingDate),
		DelistingDate: valueobject.FormatDate(s.DelistingDate),
		Status:        string(s.Status),
	}
}

// SecuritiesToPB maps a slice of domain Securities to the wire format.
func SecuritiesToPB(list []*security.Security) []*pb.Security {
	out := make([]*pb.Security, 0, len(list))
	for _, s := range list {
		out = append(out, SecurityToPB(s))
	}
	return out
}

// MarketBarToPB maps a domain MarketBar to the wire format.
func MarketBarToPB(b *marketbar.MarketBar) *pb.MarketBar {
	if b == nil {
		return nil
	}
	return &pb.MarketBar{
		StockCode:  b.StockCode,
		TradeDate:  valueobject.FormatDate(b.TradeDate),
		Period:     string(b.Period),
		OpenPrice:  b.Open,
		HighPrice:  b.High,
		LowPrice:   b.Low,
		ClosePrice: b.Close,
		Volume:     b.Volume,
		Amount:     b.Amount,
		AdjFactor:  b.AdjFactor,
	}
}

// MarketBarsToPB maps a slice of bars to the wire format.
func MarketBarsToPB(list []*marketbar.MarketBar) []*pb.MarketBar {
	out := make([]*pb.MarketBar, 0, len(list))
	for _, b := range list {
		out = append(out, MarketBarToPB(b))
	}
	return out
}

// FinancialsToPB maps financial statements to the wire format.
func FinancialsToPB(list []*financial.FinancialStatement) []*pb.FinancialStatement {
	out := make([]*pb.FinancialStatement, 0, len(list))
	for _, f := range list {
		out = append(out, &pb.FinancialStatement{
			StockCode:        f.StockCode,
			ReportDate:       valueobject.FormatDate(f.ReportDate),
			ReportType:       string(f.ReportType),
			Revenue:          f.Revenue,
			NetProfit:        f.NetProfit,
			TotalAssets:      f.TotalAssets,
			TotalLiabilities: f.TotalLiabilities,
			NetAssets:        f.NetAssets,
		})
	}
	return out
}

// FactorsToPB maps factors to the wire format.
func FactorsToPB(list []*factor.Factor) []*pb.FactorData {
	out := make([]*pb.FactorData, 0, len(list))
	for _, f := range list {
		out = append(out, &pb.FactorData{
			StockCode:   f.StockCode,
			TradeDate:   valueobject.FormatDate(f.TradeDate),
			FactorName:  f.FactorName,
			FactorValue: f.FactorValue,
		})
	}
	return out
}

// IndexBarsToPB maps index bars to the wire format.
func IndexBarsToPB(list []*indexbar.IndexBar) []*pb.IndexBar {
	out := make([]*pb.IndexBar, 0, len(list))
	for _, b := range list {
		out = append(out, &pb.IndexBar{
			IndexCode:  b.IndexCode,
			TradeDate:  valueobject.FormatDate(b.TradeDate),
			ClosePrice: b.Close,
		})
	}
	return out
}

// CalendarToPB maps a calendar slice to the wire format.
func CalendarToPB(days []calendar.TradingDay) []*pb.TradingCalendar {
	out := make([]*pb.TradingCalendar, 0, len(days))
	for _, d := range days {
		out = append(out, &pb.TradingCalendar{
			TradeDate: valueobject.FormatDate(d.TradeDate),
			IsOpen:    d.IsOpen,
		})
	}
	return out
}

// VersionsToPB maps data versions to the wire format.
func VersionsToPB(list []*dataversion.DataVersion) []*pb.DataVersion {
	out := make([]*pb.DataVersion, 0, len(list))
	for _, v := range list {
		out = append(out, &pb.DataVersion{
			Version:     v.Version,
			Description: v.Description,
			CreatedAt:   v.CreatedAt.Unix(),
		})
	}
	return out
}

// ListQueryFromPB converts a ListSecuritiesRequest into the application query.
func ListQueryFromPB(req *pb.ListSecuritiesRequest) appMarket.ListSecuritiesQuery {
	q := appMarket.ListSecuritiesQuery{
		Limit: int(req.GetLimit()),
	}
	if c := req.GetCursor(); c != nil {
		q.Cursor = c.GetNextCursor()
	}
	return q
}

// CursorPB builds the wire Cursor from an application result.
func CursorPB(next string, hasMore bool) *commonv1.Cursor {
	return &commonv1.Cursor{NextCursor: next, HasMore: hasMore}
}
