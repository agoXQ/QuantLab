package logic

import (
	"context"

	appStrategy "github.com/agoXQ/QuantLab/app/strategy/application/strategy"
	"github.com/agoXQ/QuantLab/app/strategy/internal/svc"
	"github.com/agoXQ/QuantLab/app/strategy/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateStrategyLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateStrategyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateStrategyLogic {
	return &UpdateStrategyLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateStrategyLogic) UpdateStrategy(in *pb.UpdateStrategyRequest) (*pb.UpdateStrategyResponse, error) {
	title := trimSpaces(in.Title)
	desc := trimSpaces(in.Description)
	cat := trimSpaces(in.Category)
	tags := append([]string(nil), in.Tags...)
	req := appStrategy.UpdateRequest{
		StrategyID:  in.StrategyId,
		Title:       &title,
		Description: &desc,
		Category:    &cat,
		Tags:        &tags,
	}
	if _, err := l.svcCtx.StrategySvc.Update(l.ctx, req); err != nil {
		return nil, err
	}
	return &pb.UpdateStrategyResponse{}, nil
}
