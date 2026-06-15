package logic

import (
	"context"

	"github.com/agoXQ/QuantLab/app/portfolio/internal/svc"
	"github.com/agoXQ/QuantLab/app/portfolio/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeletePortfolioLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeletePortfolioLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeletePortfolioLogic {
	return &DeletePortfolioLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DeletePortfolioLogic) DeletePortfolio(in *pb.DeletePortfolioRequest) (*pb.DeletePortfolioResponse, error) {
	// todo: add your logic here and delete this line

	return &pb.DeletePortfolioResponse{}, nil
}
