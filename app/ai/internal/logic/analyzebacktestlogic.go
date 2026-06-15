package logic

import (
	"context"

	"github.com/agoXQ/QuantLab/app/ai/internal/svc"
	"github.com/agoXQ/QuantLab/app/ai/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type AnalyzeBacktestLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAnalyzeBacktestLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AnalyzeBacktestLogic {
	return &AnalyzeBacktestLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *AnalyzeBacktestLogic) AnalyzeBacktest(in *pb.AnalyzeBacktestRequest) (*pb.AnalyzeBacktestResponse, error) {
	// todo: add your logic here and delete this line

	return &pb.AnalyzeBacktestResponse{}, nil
}
