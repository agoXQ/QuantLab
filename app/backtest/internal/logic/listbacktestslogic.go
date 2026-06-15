package logic

import (
	"context"

	"github.com/agoXQ/QuantLab/app/backtest/internal/svc"
	"github.com/agoXQ/QuantLab/app/backtest/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListBacktestsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListBacktestsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListBacktestsLogic {
	return &ListBacktestsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListBacktestsLogic) ListBacktests(in *pb.ListBacktestsRequest) (*pb.ListBacktestsResponse, error) {
	// todo: add your logic here and delete this line

	return &pb.ListBacktestsResponse{}, nil
}
