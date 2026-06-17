package evaluator

import (
	"fmt"
	"strings"

	domainAST "github.com/agoXQ/QuantLab/app/formula/domain/ast"
	domainErrors "github.com/agoXQ/QuantLab/app/formula/domain/errors"
	domainInd "github.com/agoXQ/QuantLab/app/formula/domain/indicators"
)

// evalFunctionCall resolves the function in the indicator library, evaluates
// each argument in turn, then dispatches to the indicator implementation.
//
// All built-in indicators in the library accept domainInd.Arg values; this
// function bridges between AST values and indicator args, applying the
// minimal coercion needed (Number stays Number, Bool stays Bool, Series stays
// Series). Indicators broadcast scalars internally where it makes sense.
func (e *evaluator) evalFunctionCall(c *stockContext, call *domainAST.FunctionCall) (value, error) {
	fn, ok := e.indicators.Get(call.Name)
	if !ok {
		return value{}, domainErrors.NewError(domainErrors.ErrUnknownFunction,
			fmt.Sprintf("evaluator: unknown function %s", call.Name))
	}

	args := make([]domainInd.Arg, len(call.Args))
	for i, raw := range call.Args {
		v, err := e.evalNode(c, raw)
		if err != nil {
			return value{}, err
		}
		switch v.kind {
		case valueKindNumber:
			args[i] = domainInd.NumberArg(v.num)
		case valueKindBool:
			args[i] = domainInd.BoolArg(v.bul)
		case valueKindSeries:
			args[i] = domainInd.SeriesArg(v.ser)
		default:
			return value{}, domainErrors.NewError(domainErrors.ErrTypeError,
				fmt.Sprintf("evaluator: %s arg %d has unknown kind", call.Name, i+1))
		}
	}

	out, err := fn(args)
	if err != nil {
		// Indicator-side errors (bad arg count, period type) map to the same
		// invalid-argument code the validator uses, so callers can rely on a
		// stable error code regardless of where validation falls through.
		return value{}, domainErrors.NewError(domainErrors.ErrInvalidArgCount,
			fmt.Sprintf("evaluator: %s: %s", strings.ToUpper(call.Name), err.Error()))
	}
	return seriesValue(out), nil
}
