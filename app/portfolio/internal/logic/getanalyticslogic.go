package logic

import (
	"context"

	"github.com/agoXQ/QuantLab/app/portfolio/internal/svc"
	"github.com/agoXQ/QuantLab/app/portfolio/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAnalyticsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetAnalyticsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAnalyticsLogic {
	return &GetAnalyticsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetAnalyticsLogic) GetAnalytics(in *pb.GetAnalyticsRequest) (*pb.GetAnalyticsResponse, error) {
	// todo: add your logic here and delete this line

	return &pb.GetAnalyticsResponse{}, nil
}
