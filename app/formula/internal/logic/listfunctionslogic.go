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
	defs, err := l.svcCtx.FormulaService.ListFunctions(l.ctx)
	if err != nil {
		l.Logger.Errorf("list functions error: %v", err)
		return &pb.ListFunctionsResponse{}, err
	}

	functions := make([]*pb.FunctionDefinition, 0, len(defs))
	for _, def := range defs {
		params := make([]*pb.FunctionParam, len(def.Args))
		for i, arg := range def.Args {
			params[i] = &pb.FunctionParam{
				Name:      arg.Name,
				ParamType: arg.ArgType,
			}
		}
		functions = append(functions, &pb.FunctionDefinition{
			Name:        def.Name,
			Category:    def.Category,
			ReturnType:  def.ReturnType,
			Description: def.Description,
			Params:      params,
		})
	}

	return &pb.ListFunctionsResponse{Functions: functions}, nil
}
