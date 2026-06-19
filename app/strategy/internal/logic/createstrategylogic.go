package logic

import (
	"context"

	appStrategy "github.com/agoXQ/QuantLab/app/strategy/application/strategy"
	"github.com/agoXQ/QuantLab/app/strategy/internal/svc"
	"github.com/agoXQ/QuantLab/app/strategy/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateStrategyLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateStrategyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateStrategyLogic {
	return &CreateStrategyLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// CreateStrategy delegates to the application service so HTTP and gRPC
// share the same business rules.
func (l *CreateStrategyLogic) CreateStrategy(in *pb.CreateStrategyRequest) (*pb.CreateStrategyResponse, error) {
	res, err := l.svcCtx.StrategySvc.Create(l.ctx, appStrategy.CreateRequest{
		Title:       trimSpaces(in.Title),
		Description: trimSpaces(in.Description),
		Category:    trimSpaces(in.Category),
		Tags:        in.Tags,
	})
	if err != nil {
		return nil, err
	}
	return &pb.CreateStrategyResponse{StrategyId: res.Strategy.ID}, nil
}
