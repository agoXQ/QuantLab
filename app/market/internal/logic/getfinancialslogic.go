package logic

import (
	"context"

	appMarket "github.com/agoXQ/QuantLab/app/market/application/market"
	"github.com/agoXQ/QuantLab/app/market/domain/valueobject"
	mappers "github.com/agoXQ/QuantLab/app/market/interfaces/grpc"
	"github.com/agoXQ/QuantLab/app/market/internal/svc"
	"github.com/agoXQ/QuantLab/app/market/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetFinancialsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetFinancialsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetFinancialsLogic {
	return &GetFinancialsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetFinancialsLogic) GetFinancials(in *pb.GetFinancialsRequest) (*pb.GetFinancialsResponse, error) {
	cursor := ""
	if c := in.GetCursor(); c != nil {
		cursor = c.GetNextCursor()
	}
	res, err := l.svcCtx.MarketService.GetFinancials(l.ctx, appMarket.GetFinancialsQuery{
		StockCode:  in.GetStockCode(),
		ReportType: valueobject.ReportType(in.GetReportType()),
		Cursor:     cursor,
		Limit:      int(in.GetLimit()),
	})
	if err != nil {
		return nil, err
	}
	return &pb.GetFinancialsResponse{
		Statements: mappers.FinancialsToPB(res.Items),
		Cursor:     mappers.CursorPB(res.NextCursor, res.HasMore),
	}, nil
}
