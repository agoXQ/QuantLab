package logic

import (
	"context"

	domPref "github.com/agoXQ/QuantLab/app/notification/domain/preference"
	"github.com/agoXQ/QuantLab/app/notification/internal/svc"
	"github.com/agoXQ/QuantLab/app/notification/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetPreferencesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetPreferencesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPreferencesLogic {
	return &GetPreferencesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetPreferencesLogic) GetPreferences(in *pb.GetPreferencesRequest) (*pb.GetPreferencesResponse, error) {
	pref, err := l.svcCtx.Service.GetPreferences(l.ctx, userIDFromContext(l.ctx))
	if err != nil {
		return nil, err
	}
	return &pb.GetPreferencesResponse{Preferences: preferenceToProto(pref)}, nil
}

func preferenceToProto(p *domPref.Preference) *pb.NotificationPreference {
	if p == nil {
		return nil
	}
	return &pb.NotificationPreference{
		UserId:         p.UserID,
		InAppEnabled:   p.InAppEnabled,
		EmailEnabled:   p.EmailEnabled,
		WebhookEnabled: p.WebhookEnabled,
		PushEnabled:    p.PushEnabled,
	}
}
