package logic

import (
	"context"

	"github.com/agoXQ/QuantLab/app/portfolio/internal/svc"
	"github.com/agoXQ/QuantLab/app/portfolio/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetPortfolioLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetPortfolioLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPortfolioLogic {
	return &GetPortfolioLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetPortfolioLogic) GetPortfolio(in *pb.GetPortfolioRequest) (*pb.GetPortfolioResponse, error) {
	// todo: add your logic here and delete this line

	return &pb.GetPortfolioResponse{}, nil
}
