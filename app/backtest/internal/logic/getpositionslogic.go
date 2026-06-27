package logic

import (
	"context"

	"github.com/agoXQ/QuantLab/app/backtest/internal/svc"
	"github.com/agoXQ/QuantLab/app/backtest/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetPositionsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetPositionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPositionsLogic {
	return &GetPositionsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetPositions returns the portfolio snapshots for a job. Each snapshot
// contains per-stock position data; the proto Position is a flat
// projection of the latest snapshot's holdings.
func (l *GetPositionsLogic) GetPositions(in *pb.GetPositionsRequest) (*pb.GetPositionsResponse, error) {
	snapshots, err := l.svcCtx.BacktestSvc.GetSnapshots(l.ctx, in.JobId)
	if err != nil {
		return nil, err
	}

	var positions []*pb.Position
	for _, snap := range snapshots {
		// Filter by trade_date if specified.
		if in.TradeDate != "" && snap.TradeDate.Format("2006-01-02") != in.TradeDate {
			continue
		}
		for _, h := range snap.Positions {
			positions = append(positions, &pb.Position{
				JobId:        in.JobId,
				StockCode:    h.StockCode,
				Quantity:     h.Quantity,
				CostPrice:    h.CostPrice,
				MarketPrice:  h.MarketPrice,
				MarketValue:  h.MarketValue,
				TradeDate:    snap.TradeDate.Format("2006-01-02"),
			})
		}
	}

	return &pb.GetPositionsResponse{Positions: positions}, nil
}
