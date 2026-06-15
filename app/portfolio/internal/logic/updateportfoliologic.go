package logic

import (
	"context"

	"github.com/agoXQ/QuantLab/app/portfolio/internal/svc"
	"github.com/agoXQ/QuantLab/app/portfolio/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdatePortfolioLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdatePortfolioLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdatePortfolioLogic {
	return &UpdatePortfolioLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdatePortfolioLogic) UpdatePortfolio(in *pb.UpdatePortfolioRequest) (*pb.UpdatePortfolioResponse, error) {
	// todo: add your logic here and delete this line

	return &pb.UpdatePortfolioResponse{}, nil
}
