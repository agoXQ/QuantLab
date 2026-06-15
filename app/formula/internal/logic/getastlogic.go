package logic

import (
	"context"

	"github.com/agoXQ/QuantLab/app/formula/internal/svc"
	"github.com/agoXQ/QuantLab/app/formula/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetASTLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetASTLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetASTLogic {
	return &GetASTLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetASTLogic) GetAST(in *pb.GetASTRequest) (*pb.GetASTResponse, error) {
	// todo: add your logic here and delete this line

	return &pb.GetASTResponse{}, nil
}
