package logic

import (
	"context"

	"github.com/agoXQ/QuantLab/app/formula/internal/svc"
	"github.com/agoXQ/QuantLab/app/formula/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetFunctionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetFunctionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetFunctionLogic {
	return &GetFunctionLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetFunctionLogic) GetFunction(in *pb.GetFunctionRequest) (*pb.GetFunctionResponse, error) {
	// todo: add your logic here and delete this line

	return &pb.GetFunctionResponse{}, nil
}
