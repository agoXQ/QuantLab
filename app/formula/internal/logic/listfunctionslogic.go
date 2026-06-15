package logic

import (
	"context"

	"github.com/agoXQ/QuantLab/app/formula/internal/svc"
	"github.com/agoXQ/QuantLab/app/formula/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListFunctionsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListFunctionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListFunctionsLogic {
	return &ListFunctionsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListFunctionsLogic) ListFunctions(in *pb.ListFunctionsRequest) (*pb.ListFunctionsResponse, error) {
	// todo: add your logic here and delete this line

	return &pb.ListFunctionsResponse{}, nil
}
