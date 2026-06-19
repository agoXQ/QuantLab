package logic

import (
	"context"

	appNotif "github.com/agoXQ/QuantLab/app/notification/application/notification"
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
	if _, err := l.svcCtx.Service.UpdatePreferences(l.ctx, appNotif.UpdatePreferencesInput{
		UserID:         userIDFromContext(l.ctx),
		InAppEnabled:   in.InAppEnabled,
		EmailEnabled:   in.EmailEnabled,
		WebhookEnabled: in.WebhookEnabled,
		PushEnabled:    in.PushEnabled,
	}); err != nil {
		return nil, err
	}
	return &pb.UpdatePreferencesResponse{}, nil
}
