package logic

import (
	"context"

	appNotif "github.com/agoXQ/QuantLab/app/notification/application/notification"
	"github.com/agoXQ/QuantLab/app/notification/internal/svc"
	"github.com/agoXQ/QuantLab/app/notification/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListNotificationsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListNotificationsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListNotificationsLogic {
	return &ListNotificationsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListNotificationsLogic) ListNotifications(in *pb.ListNotificationsRequest) (*pb.ListNotificationsResponse, error) {
	out, err := l.svcCtx.Service.ListNotifications(l.ctx, appNotif.ListNotificationsInput{
		UserID: userIDFromContext(l.ctx),
		Cursor: cursorString(in.Cursor),
		Limit:  int(in.Limit),
	})
	if err != nil {
		return nil, err
	}
	return &pb.ListNotificationsResponse{
		Notifications: notificationsToProto(out.Items),
		Cursor:        cursorProto(out.NextCursor),
	}, nil
}
