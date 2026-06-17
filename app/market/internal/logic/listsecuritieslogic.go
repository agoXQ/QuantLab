package logic

import (
	"context"

	mappers "github.com/agoXQ/QuantLab/app/market/interfaces/grpc"
	"github.com/agoXQ/QuantLab/app/market/internal/svc"
	"github.com/agoXQ/QuantLab/app/market/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListSecuritiesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListSecuritiesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListSecuritiesLogic {
	return &ListSecuritiesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListSecuritiesLogic) ListSecurities(in *pb.ListSecuritiesRequest) (*pb.ListSecuritiesResponse, error) {
	q := mappers.ListQueryFromPB(in)
	res, err := l.svcCtx.MarketService.ListSecurities(l.ctx, q)
	if err != nil {
		return nil, err
	}
	return &pb.ListSecuritiesResponse{
		Securities: mappers.SecuritiesToPB(res.Items),
		Cursor:     mappers.CursorPB(res.NextCursor, res.HasMore),
	}, nil
}
