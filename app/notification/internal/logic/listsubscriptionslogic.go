package logic

import (
	"context"

	appNotif "github.com/agoXQ/QuantLab/app/notification/application/notification"
	"github.com/agoXQ/QuantLab/app/notification/internal/svc"
	"github.com/agoXQ/QuantLab/app/notification/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListSubscriptionsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListSubscriptionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListSubscriptionsLogic {
	return &ListSubscriptionsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListSubscriptionsLogic) ListSubscriptions(in *pb.ListSubscriptionsRequest) (*pb.ListSubscriptionsResponse, error) {
	out, err := l.svcCtx.Service.ListSubscriptions(l.ctx, appNotif.ListSubscriptionsInput{
		SubscriberID: userIDFromContext(l.ctx),
		Cursor:       cursorString(in.Cursor),
		Limit:        int(in.Limit),
	})
	if err != nil {
		return nil, err
	}
	return &pb.ListSubscriptionsResponse{
		Subscriptions: subscriptionsToProto(out.Items),
		Cursor:        cursorProto(out.NextCursor),
	}, nil
}
