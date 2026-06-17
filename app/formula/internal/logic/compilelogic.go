package logic

import (
	"context"
	"encoding/json"

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
	result, err := l.svcCtx.FormulaService.Compile(l.ctx, in.Formula)
	if err != nil {
		l.Logger.Errorf("compile error: %v", err)
		return &pb.CompileResponse{}, err
	}

	astJSON, _ := json.Marshal(result.AST)
	planJSON, _ := json.Marshal(result.Plan)

	resp := &pb.CompileResponse{
		AstJson:   string(astJSON),
		PlanJson:  string(planJSON),
		Valid:     result.Valid,
		ErrorCode: int32(result.ErrorCode),
		Error:     result.ErrorMsg,
	}

	return resp, nil
}
