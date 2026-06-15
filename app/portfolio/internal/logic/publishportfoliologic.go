package logic

import (
	"context"

	"github.com/agoXQ/QuantLab/app/portfolio/internal/svc"
	"github.com/agoXQ/QuantLab/app/portfolio/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type PublishPortfolioLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewPublishPortfolioLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PublishPortfolioLogic {
	return &PublishPortfolioLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *PublishPortfolioLogic) PublishPortfolio(in *pb.PublishPortfolioRequest) (*pb.PublishPortfolioResponse, error) {
	// todo: add your logic here and delete this line

	return &pb.PublishPortfolioResponse{}, nil
}
