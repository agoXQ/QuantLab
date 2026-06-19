package logic

import (
	"context"

	appStrategy "github.com/agoXQ/QuantLab/app/strategy/application/strategy"
	"github.com/agoXQ/QuantLab/app/strategy/internal/svc"
	"github.com/agoXQ/QuantLab/app/strategy/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateVersionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateVersionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateVersionLogic {
	return &CreateVersionLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// CreateVersion appends a new version snapshot via the application
// service. The trim helper keeps gRPC and HTTP behaviour identical.
func (l *CreateVersionLogic) CreateVersion(in *pb.CreateVersionRequest) (*pb.CreateVersionResponse, error) {
	res, err := l.svcCtx.StrategySvc.CreateVersion(l.ctx, appStrategy.CreateVersionRequest{
		StrategyID:    in.StrategyId,
		FormulaText:   trimSpaces(in.FormulaText),
		BuyRule:       trimSpaces(in.BuyRule),
		SellRule:      trimSpaces(in.SellRule),
		RiskRule:      trimSpaces(in.RiskRule),
		PositionRule:  trimSpaces(in.PositionRule),
		RebalanceRule: trimSpaces(in.RebalanceRule),
		ChangeLog:     trimSpaces(in.ChangeLog),
	})
	if err != nil {
		return nil, err
	}
	return &pb.CreateVersionResponse{VersionId: res.Version.ID}, nil
}
