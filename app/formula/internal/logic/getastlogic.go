package logic

import (
	"context"
	"encoding/json"

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
	node, err := l.svcCtx.FormulaService.GetAST(l.ctx, in.Formula)
	if err != nil {
		l.Logger.Errorf("get AST error: %v", err)
		return &pb.GetASTResponse{}, err
	}

	astJSON, _ := json.Marshal(node)

	return &pb.GetASTResponse{
		AstJson: string(astJSON),
	}, nil
}
