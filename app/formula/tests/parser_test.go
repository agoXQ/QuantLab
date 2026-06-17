package tests

import (
	"context"
	"testing"

	domainAST "github.com/agoXQ/QuantLab/app/formula/domain/ast"
	infraFunc "github.com/agoXQ/QuantLab/app/formula/infrastructure/function"
	infraLexer "github.com/agoXQ/QuantLab/app/formula/infrastructure/lexer"
	infraParser "github.com/agoXQ/QuantLab/app/formula/infrastructure/parser"
	infraVar "github.com/agoXQ/QuantLab/app/formula/infrastructure/variable"
)

func parseFormula(t *testing.T, input string) domainAST.Node {
	t.Helper()
	funcReg := infraFunc.NewRegistry()
	varReg := infraVar.NewRegistry()
	lexer := infraLexer.NewLexer()
	parser := infraParser.NewParser(funcReg, varReg)

	tokens, err := lexer.Tokenize(context.Background(), input)
	if err != nil {
		t.Fatalf("Tokenize failed: %v", err)
	}

	node, err := parser.Parse(context.Background(), tokens)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	return node
}

func parseFormulaErr(t *testing.T, input string) error {
	t.Helper()
	funcReg := infraFunc.NewRegistry()
	varReg := infraVar.NewRegistry()
	lexer := infraLexer.NewLexer()
	parser := infraParser.NewParser(funcReg, varReg)

	tokens, err := lexer.Tokenize(context.Background(), input)
	if err != nil {
		return err
	}

	_, err = parser.Parse(context.Background(), tokens)
	return err
}

func TestParser_SimpleIdentifier(t *testing.T) {
	node := parseFormula(t, "ROE")
	ident, ok := node.(*domainAST.Identifier)
	if !ok {
		t.Fatalf("expected *Identifier, got %T", node)
	}
	if ident.Name != "ROE" {
		t.Errorf("expected name ROE, got %s", ident.Name)
	}
}

func TestParser_NumberLiteral(t *testing.T) {
	node := parseFormula(t, "15")
	num, ok := node.(*domainAST.NumberLiteral)
	if !ok {
		t.Fatalf("expected *NumberLiteral, got %T", node)
	}
	if num.Value != 15 {
		t.Errorf("expected 15, got %f", num.Value)
	}
}

func TestParser_FloatLiteral(t *testing.T) {
	node := parseFormula(t, "20.5")
	num, ok := node.(*domainAST.NumberLiteral)
	if !ok {
		t.Fatalf("expected *NumberLiteral, got %T", node)
	}
	if num.Value != 20.5 {
		t.Errorf("expected 20.5, got %f", num.Value)
	}
}

func TestParser_Comparison(t *testing.T) {
	node := parseFormula(t, "ROE > 15")
	bin, ok := node.(*domainAST.BinaryExpression)
	if !ok {
		t.Fatalf("expected *BinaryExpression, got %T", node)
	}
	if bin.Operator != ">" {
		t.Errorf("expected operator '>', got %q", bin.Operator)
	}

	left, ok := bin.Left.(*domainAST.Identifier)
	if !ok || left.Name != "ROE" {
		t.Errorf("expected left identifier ROE, got %T %+v", bin.Left, bin.Left)
	}

	right, ok := bin.Right.(*domainAST.NumberLiteral)
	if !ok || right.Value != 15 {
		t.Errorf("expected right number 15, got %T %+v", bin.Right, bin.Right)
	}
}

func TestParser_AndExpression(t *testing.T) {
	node := parseFormula(t, "ROE > 15 AND PE < 20")
	bin, ok := node.(*domainAST.BinaryExpression)
	if !ok {
		t.Fatalf("expected *BinaryExpression, got %T", node)
	}
	if bin.Operator != "AND" {
		t.Errorf("expected AND, got %q", bin.Operator)
	}
}

func TestParser_FunctionCall(t *testing.T) {
	node := parseFormula(t, "MA(CLOSE,5)")
	fn, ok := node.(*domainAST.FunctionCall)
	if !ok {
		t.Fatalf("expected *FunctionCall, got %T", node)
	}
	if fn.Name != "MA" {
		t.Errorf("expected MA, got %s", fn.Name)
	}
	if len(fn.Args) != 2 {
		t.Fatalf("expected 2 args, got %d", len(fn.Args))
	}

	arg0, ok := fn.Args[0].(*domainAST.Identifier)
	if !ok || arg0.Name != "CLOSE" {
		t.Errorf("expected arg0 CLOSE, got %T %+v", fn.Args[0], fn.Args[0])
	}

	arg1, ok := fn.Args[1].(*domainAST.NumberLiteral)
	if !ok || arg1.Value != 5 {
		t.Errorf("expected arg1 5, got %T %+v", fn.Args[1], fn.Args[1])
	}
}

func TestParser_NestedFunctionCall(t *testing.T) {
	node := parseFormula(t, "CROSS(MA(CLOSE,5),MA(CLOSE,20))")
	fn, ok := node.(*domainAST.FunctionCall)
	if !ok {
		t.Fatalf("expected *FunctionCall, got %T", node)
	}
	if fn.Name != "CROSS" {
		t.Errorf("expected CROSS, got %s", fn.Name)
	}
	if len(fn.Args) != 2 {
		t.Fatalf("expected 2 args, got %d", len(fn.Args))
	}

	_, ok = fn.Args[0].(*domainAST.FunctionCall)
	if !ok {
		t.Errorf("expected arg0 to be FunctionCall, got %T", fn.Args[0])
	}
	_, ok = fn.Args[1].(*domainAST.FunctionCall)
	if !ok {
		t.Errorf("expected arg1 to be FunctionCall, got %T", fn.Args[1])
	}
}

func TestParser_ParenthesizedExpression(t *testing.T) {
	node := parseFormula(t, "(ROE > 15)")
	bin, ok := node.(*domainAST.BinaryExpression)
	if !ok {
		t.Fatalf("expected *BinaryExpression, got %T", node)
	}
	if bin.Operator != ">" {
		t.Errorf("expected >, got %q", bin.Operator)
	}
}

func TestParser_OperatorPrecedence(t *testing.T) {
	node := parseFormula(t, "ROE > 15 AND PE < 20")
	bin, ok := node.(*domainAST.BinaryExpression)
	if !ok {
		t.Fatalf("expected *BinaryExpression, got %T", node)
	}
	if bin.Operator != "AND" {
		t.Errorf("expected AND at root, got %q", bin.Operator)
	}

	left, ok := bin.Left.(*domainAST.BinaryExpression)
	if !ok {
		t.Fatalf("expected left to be BinaryExpression, got %T", bin.Left)
	}
	if left.Operator != ">" {
		t.Errorf("expected left operator >, got %q", left.Operator)
	}

	right, ok := bin.Right.(*domainAST.BinaryExpression)
	if !ok {
		t.Fatalf("expected right to be BinaryExpression, got %T", bin.Right)
	}
	if right.Operator != "<" {
		t.Errorf("expected right operator <, got %q", right.Operator)
	}
}

func TestParser_ArithmeticPrecedence(t *testing.T) {
	node := parseFormula(t, "1 + 2 * 3")
	bin, ok := node.(*domainAST.BinaryExpression)
	if !ok {
		t.Fatalf("expected *BinaryExpression, got %T", node)
	}
	if bin.Operator != "+" {
		t.Errorf("expected + at root, got %q", bin.Operator)
	}

	_, ok = bin.Right.(*domainAST.BinaryExpression)
	if !ok {
		t.Errorf("expected right to be BinaryExpression (2*3), got %T", bin.Right)
	}
}

func TestParser_UnaryMinus(t *testing.T) {
	node := parseFormula(t, "-5")
	unary, ok := node.(*domainAST.UnaryExpression)
	if !ok {
		t.Fatalf("expected *UnaryExpression, got %T", node)
	}
	if unary.Operator != "-" {
		t.Errorf("expected operator '-', got %q", unary.Operator)
	}
	num, ok := unary.Operand.(*domainAST.NumberLiteral)
	if !ok || num.Value != 5 {
		t.Errorf("expected operand 5, got %T %+v", unary.Operand, unary.Operand)
	}
}

func TestParser_NotExpression(t *testing.T) {
	node := parseFormula(t, "NOT ROE > 15")
	bin, ok := node.(*domainAST.BinaryExpression)
	if !ok {
		t.Fatalf("expected *BinaryExpression, got %T", node)
	}
	if bin.Operator != ">" {
		t.Errorf("expected > at root, got %q", bin.Operator)
	}
	unary, ok := bin.Left.(*domainAST.UnaryExpression)
	if !ok {
		t.Fatalf("expected left to be UnaryExpression, got %T", bin.Left)
	}
	if unary.Operator != "NOT" {
		t.Errorf("expected NOT, got %q", unary.Operator)
	}
}

func TestParser_StringLiteral(t *testing.T) {
	node := parseFormula(t, `"银行"`)
	str, ok := node.(*domainAST.StringLiteral)
	if !ok {
		t.Fatalf("expected *StringLiteral, got %T", node)
	}
	if str.Value != "银行" {
		t.Errorf("expected 银行, got %q", str.Value)
	}
}

func TestParser_UnknownFunctionError(t *testing.T) {
	err := parseFormulaErr(t, "UNKNOWN_FUNC(1,2)")
	if err == nil {
		t.Fatal("expected error for unknown function")
	}
}

func TestParser_MissingParenError(t *testing.T) {
	err := parseFormulaErr(t, "MA(CLOSE,5")
	if err == nil {
		t.Fatal("expected error for missing parenthesis")
	}
}

func TestParser_EmptyParens(t *testing.T) {
	err := parseFormulaErr(t, "()")
	if err == nil {
		t.Fatal("expected error for empty parentheses")
	}
}

func TestParser_ExtraTokens(t *testing.T) {
	err := parseFormulaErr(t, "ROE 15")
	if err == nil {
		t.Fatal("expected error for missing operator")
	}
}

func TestParser_StringLiteralInExpression(t *testing.T) {
	node := parseFormula(t, `PE > "ABC"`)
	bin, ok := node.(*domainAST.BinaryExpression)
	if !ok {
		t.Fatalf("expected *BinaryExpression, got %T", node)
	}
	if bin.Operator != ">" {
		t.Errorf("expected >, got %q", bin.Operator)
	}
	_, ok = bin.Right.(*domainAST.StringLiteral)
	if !ok {
		t.Errorf("expected right to be StringLiteral, got %T", bin.Right)
	}
}
