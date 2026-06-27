package logic

import (
	"context"

	"github.com/agoXQ/QuantLab/app/backtest/internal/svc"
	"github.com/agoXQ/QuantLab/app/backtest/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetTradesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetTradesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTradesLogic {
	return &GetTradesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetTrades returns the trade list for a job.
func (l *GetTradesLogic) GetTrades(in *pb.GetTradesRequest) (*pb.GetTradesResponse, error) {
	trades, err := l.svcCtx.BacktestSvc.GetTrades(l.ctx, in.JobId)
	if err != nil {
		return nil, err
	}

	pbTrades := make([]*pb.Trade, 0, len(trades))
	for _, t := range trades {
		var tradeTime int64
		if !t.TradeTime.IsZero() {
			tradeTime = t.TradeTime.Unix()
		}
		pbTrades = append(pbTrades, &pb.Trade{
			Id:         t.ID,
			OrderId:    t.OrderID,
			StockCode:  t.StockCode,
			Quantity:   t.Quantity,
			Price:      t.Price,
			Commission: t.Commission,
			TradeTime:  tradeTime,
		})
	}

	return &pb.GetTradesResponse{Trades: pbTrades}, nil
}
