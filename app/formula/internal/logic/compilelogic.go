package logic

import (
	"context"

	"github.com/agoXQ/QuantLab/app/formula/internal/svc"
	"github.com/agoXQ/QuantLab/app/formula/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type CompileLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCompileLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CompileLogic {
	return &CompileLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CompileLogic) Compile(in *pb.CompileRequest) (*pb.CompileResponse, error) {
	// todo: add your logic here and delete this line

	return &pb.CompileResponse{}, nil
}
