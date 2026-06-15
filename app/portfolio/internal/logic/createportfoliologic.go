package logic

import (
	"context"

	"github.com/agoXQ/QuantLab/app/portfolio/internal/svc"
	"github.com/agoXQ/QuantLab/app/portfolio/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreatePortfolioLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreatePortfolioLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreatePortfolioLogic {
	return &CreatePortfolioLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreatePortfolioLogic) CreatePortfolio(in *pb.CreatePortfolioRequest) (*pb.CreatePortfolioResponse, error) {
	// todo: add your logic here and delete this line

	return &pb.CreatePortfolioResponse{}, nil
}
