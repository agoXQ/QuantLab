package logic

import (
	"context"

	"github.com/agoXQ/QuantLab/app/ai/internal/svc"
	"github.com/agoXQ/QuantLab/app/ai/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type OptimizePortfolioLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewOptimizePortfolioLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OptimizePortfolioLogic {
	return &OptimizePortfolioLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *OptimizePortfolioLogic) OptimizePortfolio(in *pb.OptimizePortfolioRequest) (*pb.OptimizePortfolioResponse, error) {
	// todo: add your logic here and delete this line

	return &pb.OptimizePortfolioResponse{}, nil
}
