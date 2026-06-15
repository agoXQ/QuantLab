package logic

import (
	"context"

	"github.com/agoXQ/QuantLab/app/strategy/internal/svc"
	"github.com/agoXQ/QuantLab/app/strategy/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type ArchiveStrategyLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewArchiveStrategyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ArchiveStrategyLogic {
	return &ArchiveStrategyLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ArchiveStrategyLogic) ArchiveStrategy(in *pb.ArchiveStrategyRequest) (*pb.ArchiveStrategyResponse, error) {
	// todo: add your logic here and delete this line

	return &pb.ArchiveStrategyResponse{}, nil
}
