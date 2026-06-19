package logic

import (
	"context"

	appStrategy "github.com/agoXQ/QuantLab/app/strategy/application/strategy"
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

// ArchiveStrategy delegates to the application service so archival is
// driven by the same lifecycle rules HTTP uses.
func (l *ArchiveStrategyLogic) ArchiveStrategy(in *pb.ArchiveStrategyRequest) (*pb.ArchiveStrategyResponse, error) {
	if _, err := l.svcCtx.StrategySvc.Archive(l.ctx, appStrategy.ArchiveRequest{
		StrategyID: in.StrategyId,
	}); err != nil {
		return nil, err
	}
	return &pb.ArchiveStrategyResponse{}, nil
}
