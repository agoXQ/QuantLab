package logic

import (
	"context"

	"github.com/agoXQ/QuantLab/app/ai/internal/svc"
	"github.com/agoXQ/QuantLab/app/ai/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GeneratePortfolioLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGeneratePortfolioLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GeneratePortfolioLogic {
	return &GeneratePortfolioLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GeneratePortfolioLogic) GeneratePortfolio(in *pb.GeneratePortfolioRequest) (*pb.GeneratePortfolioResponse, error) {
	// todo: add your logic here and delete this line

	return &pb.GeneratePortfolioResponse{}, nil
}
