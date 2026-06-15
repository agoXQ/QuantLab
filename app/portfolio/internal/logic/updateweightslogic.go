package logic

import (
	"context"

	"github.com/agoXQ/QuantLab/app/portfolio/internal/svc"
	"github.com/agoXQ/QuantLab/app/portfolio/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateWeightsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateWeightsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateWeightsLogic {
	return &UpdateWeightsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateWeightsLogic) UpdateWeights(in *pb.UpdateWeightsRequest) (*pb.UpdateWeightsResponse, error) {
	// todo: add your logic here and delete this line

	return &pb.UpdateWeightsResponse{}, nil
}
