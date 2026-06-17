package validator

import (
	"context"
	"fmt"
	"strings"

	domainAST "github.com/agoXQ/QuantLab/app/formula/domain/ast"
	domainErrors "github.com/agoXQ/QuantLab/app/formula/domain/errors"
	domainFunc "github.com/agoXQ/QuantLab/app/formula/domain/function"
	domainVar "github.com/agoXQ/QuantLab/app/formula/domain/variable"
	domainValidator "github.com/agoXQ/QuantLab/app/formula/domain/validator"
)

type validator struct {
	funcRegistry domainFunc.Registry
	varRegistry  domainVar.Registry
	userVars     map[string]bool // user-defined variable names (from assignments)
}

// NewValidator creates a new validator instance.
func NewValidator(funcRegistry domainFunc.Registry, varRegistry domainVar.Registry) domainValidator.Validator {
	return &validator{
		funcRegistry: funcRegistry,
		varRegistry:  varRegistry,
		userVars:     make(map[string]bool),
	}
}

func (v *validator) Validate(_ context.Context, node domainAST.Node) (*domainValidator.ValidationResult, error) {
	// Reset user-defined variables for each validation call
	v.userVars = make(map[string]bool)

	errs := v.validateNode(node, "")
	if len(errs) > 0 {
		return &domainValidator.ValidationResult{
			Valid:  false,
			Errors: errs,
		}, nil
	}
	return &domainValidator.ValidationResult{Valid: true}, nil
}

func (v *validator) validateNode(node domainAST.Node, contextType string) []domainValidator.ValidationError {
	switch n := node.(type) {
	case *domainAST.BinaryExpression:
		return v.validateBinaryExpr(n)
	case *domainAST.UnaryExpression:
		return v.validateUnaryExpr(n)
	case *domainAST.FunctionCall:
		return v.validateFunctionCall(n)
	case *domainAST.Identifier:
		return v.validateIdentifier(n)
	case *domainAST.NumberLiteral:
		return nil
	case *domainAST.StringLiteral:
		return nil
	case *domainAST.BoolLiteral:
		return nil
	case *domainAST.Assignment:
		return v.validateAssignment(n)
	case *domainAST.Program:
		return v.validateProgram(n)
	default:
		return []domainValidator.ValidationError{{
			Code:     0,
			CodeStr:  "UNKNOWN_NODE",
			Message:  fmt.Sprintf("unknown AST node type: %T", node),
			Severity: domainValidator.SeverityError,
		}}
	}
}

func (v *validator) validateAssignment(assign *domainAST.Assignment) []domainValidator.ValidationError {
	var errs []domainValidator.ValidationError

	// Validate the value expression
	errs = append(errs, v.validateNode(assign.Value, "")...)

	// Register the assigned name as a user-defined variable
	v.userVars[strings.ToLower(assign.Name)] = true

	return errs
}

func (v *validator) validateProgram(prog *domainAST.Program) []domainValidator.ValidationError {
	var errs []domainValidator.ValidationError

	for _, stmt := range prog.Statements {
		errs = append(errs, v.validateNode(stmt, "")...)
	}

	return errs
}

func (v *validator) validateBinaryExpr(expr *domainAST.BinaryExpression) []domainValidator.ValidationError {
	var errs []domainValidator.ValidationError

	errs = append(errs, v.validateNode(expr.Left, "")...)
	errs = append(errs, v.validateNode(expr.Right, "")...)

	switch strings.ToUpper(expr.Operator) {
	case "AND", "OR":
		leftType := v.inferType(expr.Left)
		rightType := v.inferType(expr.Right)
		if leftType != "" && leftType != "Boolean" && leftType != "Signal" {
			errs = append(errs, domainValidator.ValidationError{
				Code:     domainErrors.ErrTypeError,
				CodeStr:  "TYPE_ERROR",
				Message:  fmt.Sprintf("AND/OR requires boolean operands, got %s", leftType),
				Severity: domainValidator.SeverityError,
			})
		}
		if rightType != "" && rightType != "Boolean" && rightType != "Signal" {
			errs = append(errs, domainValidator.ValidationError{
				Code:     domainErrors.ErrTypeError,
				CodeStr:  "TYPE_ERROR",
				Message:  fmt.Sprintf("AND/OR requires boolean operands, got %s", rightType),
				Severity: domainValidator.SeverityError,
			})
		}
	case ">", "<", ">=", "<=", "==", "!=":
		leftType := v.inferType(expr.Left)
		rightType := v.inferType(expr.Right)
		if leftType != "" && rightType != "" && leftType != rightType {
			// Allow Series vs Number comparison (element-wise comparison produces filter)
			if !(leftType == "Series" && rightType == "Number") && !(leftType == "Number" && rightType == "Series") {
				errs = append(errs, domainValidator.ValidationError{
					Code:     domainErrors.ErrTypeError,
					CodeStr:  "TYPE_ERROR",
					Message:  fmt.Sprintf("cannot compare %s with %s", leftType, rightType),
					Severity: domainValidator.SeverityError,
				})
			}
		}
	case "+", "-", "*", "/", "%":
		leftType := v.inferType(expr.Left)
		rightType := v.inferType(expr.Right)
		if leftType != "" && leftType != "Number" && leftType != "Series" {
			errs = append(errs, domainValidator.ValidationError{
				Code:     domainErrors.ErrTypeError,
				CodeStr:  "TYPE_ERROR",
				Message:  fmt.Sprintf("arithmetic requires numeric operands, got %s", leftType),
				Severity: domainValidator.SeverityError,
			})
		}
		if rightType != "" && rightType != "Number" && rightType != "Series" {
			errs = append(errs, domainValidator.ValidationError{
				Code:     domainErrors.ErrTypeError,
				CodeStr:  "TYPE_ERROR",
				Message:  fmt.Sprintf("arithmetic requires numeric operands, got %s", rightType),
				Severity: domainValidator.SeverityError,
			})
		}
	}

	return errs
}

func (v *validator) validateUnaryExpr(expr *domainAST.UnaryExpression) []domainValidator.ValidationError {
	return v.validateNode(expr.Operand, "")
}

func (v *validator) validateFunctionCall(call *domainAST.FunctionCall) []domainValidator.ValidationError {
	var errs []domainValidator.ValidationError

	def, exists := v.funcRegistry.GetFunction(call.Name)
	if !exists {
		errs = append(errs, domainValidator.ValidationError{
			Code:     domainErrors.ErrUnknownFunction,
			CodeStr:  "UNKNOWN_FUNCTION",
			Message:  fmt.Sprintf("unknown function: %s", call.Name),
			Severity: domainValidator.SeverityError,
		})
		return errs
	}

	if len(call.Args) < len(def.Args) {
		requiredCount := 0
		for _, arg := range def.Args {
			if arg.Required {
				requiredCount++
			}
		}
		if len(call.Args) < requiredCount {
			errs = append(errs, domainValidator.ValidationError{
				Code:     domainErrors.ErrInvalidArgCount,
				CodeStr:  "INVALID_ARG_COUNT",
				Message:  fmt.Sprintf("function %s expects at least %d arguments, got %d", call.Name, requiredCount, len(call.Args)),
				Severity: domainValidator.SeverityError,
			})
		}
	}

	for i, arg := range call.Args {
		errs = append(errs, v.validateNode(arg, "")...)
		if i < len(def.Args) {
			argType := v.inferType(arg)
			expectedType := def.Args[i].ArgType
			if argType != "" && argType != expectedType {
				// Allow Series where Number is expected for compatible functions (ABS, MAX, MIN)
				if !(expectedType == "Number" && argType == "Series" && (call.Name == "ABS" || call.Name == "MAX" || call.Name == "MIN")) {
					errs = append(errs, domainValidator.ValidationError{
						Code:     domainErrors.ErrTypeError,
						CodeStr:  "TYPE_ERROR",
						Message:  fmt.Sprintf("argument %d of %s expects %s, got %s", i+1, call.Name, expectedType, argType),
						Severity: domainValidator.SeverityError,
					})
				}
			}
		}
	}

	if call.Name == "REF" && len(call.Args) >= 2 {
		if numLit, ok := call.Args[1].(*domainAST.NumberLiteral); ok && numLit.Value < 0 {
			errs = append(errs, domainValidator.ValidationError{
				Code:     domainErrors.ErrFutureFunction,
				CodeStr:  "FUTURE_FUNCTION",
				Message:  "future data reference detected: REF with negative index",
				Severity: domainValidator.SeverityError,
			})
		}
	}

	return errs
}

func (v *validator) validateIdentifier(ident *domainAST.Identifier) []domainValidator.ValidationError {
	if v.varRegistry.Exists(ident.Name) || v.funcRegistry.Exists(ident.Name) {
		return nil
	}

	// Check user-defined variables
	if v.userVars[strings.ToLower(ident.Name)] {
		return nil
	}

	upper := strings.ToUpper(ident.Name)
	if upper == "TRUE" || upper == "FALSE" {
		return nil
	}

	return []domainValidator.ValidationError{{
		Code:     domainErrors.ErrUnknownVariable,
		CodeStr:  "UNKNOWN_VARIABLE",
		Message:  fmt.Sprintf("unknown variable: %s", ident.Name),
		Severity: domainValidator.SeverityError,
	}}
}

func (v *validator) inferType(node domainAST.Node) string {
	switch n := node.(type) {
	case *domainAST.NumberLiteral:
		return "Number"
	case *domainAST.StringLiteral:
		return "String"
	case *domainAST.BoolLiteral:
		return "Boolean"
	case *domainAST.Assignment:
		return v.inferType(n.Value)
	case *domainAST.Program:
		if len(n.Statements) > 0 {
			return v.inferType(n.Statements[len(n.Statements)-1])
		}
		return ""
	case *domainAST.Identifier:
		upper := strings.ToUpper(n.Name)
		if upper == "TRUE" || upper == "FALSE" {
			return "Boolean"
		}
		if varDef, ok := v.varRegistry.GetVariable(n.Name); ok {
			return string(varDef.VarType)
		}
		return ""
	case *domainAST.FunctionCall:
		if def, ok := v.funcRegistry.GetFunction(n.Name); ok {
			return def.ReturnType
		}
		return ""
	case *domainAST.BinaryExpression:
		op := strings.ToUpper(n.Operator)
		switch op {
		case "AND", "OR":
			return "Boolean"
		case ">", "<", ">=", "<=", "==", "!=":
			return "Boolean"
		default:
			// Arithmetic: if either operand is Series, the result is Series.
			// e.g. REF(VOL,1) * 1.5 -> Series, CLOSE + 10 -> Series
			leftType := v.inferType(n.Left)
			rightType := v.inferType(n.Right)
			if leftType == "Series" || rightType == "Series" {
				return "Series"
			}
			return "Number"
		}
	case *domainAST.UnaryExpression:
		op := strings.ToUpper(n.Operator)
		if op == "NOT" {
			return "Boolean"
		}
		return v.inferType(n.Operand)
	default:
		return ""
	}
}
