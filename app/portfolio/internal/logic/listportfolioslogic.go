package logic

import (
	"context"

	"github.com/agoXQ/QuantLab/app/portfolio/internal/svc"
	"github.com/agoXQ/QuantLab/app/portfolio/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListPortfoliosLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListPortfoliosLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListPortfoliosLogic {
	return &ListPortfoliosLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListPortfoliosLogic) ListPortfolios(in *pb.ListPortfoliosRequest) (*pb.ListPortfoliosResponse, error) {
	// todo: add your logic here and delete this line

	return &pb.ListPortfoliosResponse{}, nil
}
