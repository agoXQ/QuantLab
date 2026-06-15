package logic

import (
	"context"

	"github.com/agoXQ/QuantLab/app/notification/internal/svc"
	"github.com/agoXQ/QuantLab/app/notification/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdatePreferencesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdatePreferencesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdatePreferencesLogic {
	return &UpdatePreferencesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdatePreferencesLogic) UpdatePreferences(in *pb.UpdatePreferencesRequest) (*pb.UpdatePreferencesResponse, error) {
	// todo: add your logic here and delete this line

	return &pb.UpdatePreferencesResponse{}, nil
}
