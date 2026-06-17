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
	def, err := l.svcCtx.FormulaService.GetFunction(l.ctx, in.Name)
	if err != nil {
		l.Logger.Errorf("get function error: %v", err)
		return &pb.GetFunctionResponse{}, err
	}

	if def == nil {
		return &pb.GetFunctionResponse{}, nil
	}

	params := make([]*pb.FunctionParam, len(def.Args))
	for i, arg := range def.Args {
		params[i] = &pb.FunctionParam{
			Name:      arg.Name,
			ParamType: arg.ArgType,
		}
	}

	return &pb.GetFunctionResponse{
		Function: &pb.FunctionDefinition{
			Name:        def.Name,
			Category:    def.Category,
			ReturnType:  def.ReturnType,
			Description: def.Description,
			Params:      params,
		},
	}, nil
}
