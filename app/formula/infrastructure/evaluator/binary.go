package evaluator

import (
	"fmt"
	"math"
	"strings"
	"time"

	domainAST "github.com/agoXQ/QuantLab/app/formula/domain/ast"
	domainErrors "github.com/agoXQ/QuantLab/app/formula/domain/errors"
	"github.com/agoXQ/QuantLab/app/formula/domain/series"
)

// evalBinary handles arithmetic, comparison, and boolean binary operators.
//
// The evaluator picks the cheapest representation: scalar-scalar produces a
// scalar, series-anything is broadcast against the series timestamps. This
// is the only place where heterogeneous types collapse, which keeps the rest
// of the AST traversal type-stable.
func (e *evaluator) evalBinary(c *stockContext, expr *domainAST.BinaryExpression) (value, error) {
	left, err := e.evalNode(c, expr.Left)
	if err != nil {
		return value{}, err
	}
	right, err := e.evalNode(c, expr.Right)
	if err != nil {
		return value{}, err
	}

	op := strings.ToUpper(expr.Operator)
	switch op {
	case "AND", "OR":
		return logicalCombine(left, right, op), nil
	case ">", "<", ">=", "<=", "==", "!=":
		return compareCombine(left, right, op), nil
	case "+", "-", "*", "/", "%":
		return arithmeticCombine(left, right, op)
	default:
		return value{}, domainErrors.NewError(domainErrors.ErrSyntaxError,
			fmt.Sprintf("evaluator: unsupported operator %s", expr.Operator))
	}
}

// evalUnary handles `-x` and `NOT x`.
func (e *evaluator) evalUnary(c *stockContext, expr *domainAST.UnaryExpression) (value, error) {
	v, err := e.evalNode(c, expr.Operand)
	if err != nil {
		return value{}, err
	}
	op := strings.ToUpper(expr.Operator)
	switch op {
	case "NOT":
		return boolValue(!v.asBool()), nil
	case "-":
		switch v.kind {
		case valueKindNumber:
			return numberValue(-v.num), nil
		case valueKindBool:
			if v.bul {
				return numberValue(-1), nil
			}
			return numberValue(0), nil
		case valueKindSeries:
			return seriesValue(v.ser.Map(func(x float64) float64 { return -x })), nil
		}
	case "+":
		return v, nil
	}
	return value{}, domainErrors.NewError(domainErrors.ErrSyntaxError,
		fmt.Sprintf("evaluator: unsupported unary operator %s", expr.Operator))
}

// --- combinators ---

func logicalCombine(a, b value, op string) value {
	switch op {
	case "AND":
		return boolValue(a.asBool() && b.asBool())
	case "OR":
		return boolValue(a.asBool() || b.asBool())
	}
	return boolValue(false)
}

func compareCombine(a, b value, op string) value {
	if a.kind == valueKindSeries || b.kind == valueKindSeries {
		s := compareSeries(a, b, op)
		return seriesValue(s)
	}
	return boolValue(compareScalar(a.asNumber(), b.asNumber(), op))
}

func compareSeries(a, b value, op string) series.Series {
	left := toSeries(a, b)
	right := toSeries(b, a)
	out := make([]float64, left.Len())
	for i := 0; i < left.Len(); i++ {
		x, y := left.At(i), right.At(i)
		if math.IsNaN(x) || math.IsNaN(y) {
			out[i] = math.NaN()
			continue
		}
		if compareScalar(x, y, op) {
			out[i] = 1
		} else {
			out[i] = 0
		}
	}
	r, _ := series.NewSeries(left.Timestamps(), out)
	return r
}

func compareScalar(x, y float64, op string) bool {
	if math.IsNaN(x) || math.IsNaN(y) {
		return false
	}
	switch op {
	case ">":
		return x > y
	case "<":
		return x < y
	case ">=":
		return x >= y
	case "<=":
		return x <= y
	case "==":
		return x == y
	case "!=":
		return x != y
	}
	return false
}

func arithmeticCombine(a, b value, op string) (value, error) {
	if a.kind == valueKindSeries || b.kind == valueKindSeries {
		left := toSeries(a, b)
		right := toSeries(b, a)
		out := make([]float64, left.Len())
		for i := 0; i < left.Len(); i++ {
			r, err := scalarArithmetic(left.At(i), right.At(i), op)
			if err != nil {
				return value{}, err
			}
			out[i] = r
		}
		r, _ := series.NewSeries(left.Timestamps(), out)
		return seriesValue(r), nil
	}
	r, err := scalarArithmetic(a.asNumber(), b.asNumber(), op)
	if err != nil {
		return value{}, err
	}
	return numberValue(r), nil
}

func scalarArithmetic(x, y float64, op string) (float64, error) {
	if math.IsNaN(x) || math.IsNaN(y) {
		return math.NaN(), nil
	}
	switch op {
	case "+":
		return x + y, nil
	case "-":
		return x - y, nil
	case "*":
		return x * y, nil
	case "/":
		if y == 0 {
			return math.NaN(), nil
		}
		return x / y, nil
	case "%":
		if y == 0 {
			return math.NaN(), nil
		}
		return float64(int64(x) % int64(y)), nil
	default:
		return math.NaN(), domainErrors.NewError(domainErrors.ErrSyntaxError,
			fmt.Sprintf("evaluator: unsupported arithmetic %s", op))
	}
}

// toSeries returns v as a Series. When v is scalar, it is broadcast against
// the timestamps of the companion value.
func toSeries(v, anchor value) series.Series {
	if v.kind == valueKindSeries {
		return v.ser
	}
	var ts []time.Time
	if anchor.kind == valueKindSeries {
		ts = anchor.ser.Timestamps()
	}
	return series.Repeat(ts, v.asNumber())
}
