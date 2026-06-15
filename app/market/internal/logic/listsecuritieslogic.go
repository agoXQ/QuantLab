package logic

import (
	"context"

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
	// todo: add your logic here and delete this line

	return &pb.ListSecuritiesResponse{}, nil
}
