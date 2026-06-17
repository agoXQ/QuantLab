package tests

import (
	"context"
	"testing"

	domainLexer "github.com/agoXQ/QuantLab/app/formula/domain/lexer"
	infraLexer "github.com/agoXQ/QuantLab/app/formula/infrastructure/lexer"
)

func lexerTokenize(t *testing.T, input string) []domainLexer.Token {
	t.Helper()
	l := infraLexer.NewLexer()
	tokens, err := l.Tokenize(context.Background(), input)
	if err != nil {
		t.Fatalf("Tokenize failed: %v", err)
	}
	return tokens
}

func assertToken(t *testing.T, tok domainLexer.Token, expectedType domainLexer.TokenType, expectedLiteral string) {
	t.Helper()
	if tok.Type != expectedType {
		t.Errorf("expected token type %s, got %s (literal: %q)", expectedType, tok.Type, tok.Literal)
	}
	if tok.Literal != expectedLiteral {
		t.Errorf("expected token literal %q, got %q", expectedLiteral, tok.Literal)
	}
}

func TestLexer_SimpleIdentifier(t *testing.T) {
	tokens := lexerTokenize(t, "ROE")
	if len(tokens) < 2 {
		t.Fatalf("expected at least 2 tokens (identifier + EOF), got %d", len(tokens))
	}
	assertToken(t, tokens[0], domainLexer.TOKEN_IDENTIFIER, "ROE")
	assertToken(t, tokens[len(tokens)-1], domainLexer.TOKEN_EOF, "")
}

func TestLexer_Number(t *testing.T) {
	tokens := lexerTokenize(t, "15")
	assertToken(t, tokens[0], domainLexer.TOKEN_NUMBER, "15")
	assertToken(t, tokens[1], domainLexer.TOKEN_EOF, "")
}

func TestLexer_FloatNumber(t *testing.T) {
	tokens := lexerTokenize(t, "20.5")
	assertToken(t, tokens[0], domainLexer.TOKEN_NUMBER, "20.5")
}

func TestLexer_ComparisonOperators(t *testing.T) {
	tests := []struct {
		input   string
		lit     string
		tokType domainLexer.TokenType
	}{
		{">", ">", domainLexer.TOKEN_GT},
		{"<", "<", domainLexer.TOKEN_LT},
		{">=", ">=", domainLexer.TOKEN_GTE},
		{"<=", "<=", domainLexer.TOKEN_LTE},
		{"==", "==", domainLexer.TOKEN_EQ},
		{"!=", "!=", domainLexer.TOKEN_NEQ},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			tokens := lexerTokenize(t, tt.input)
			assertToken(t, tokens[0], tt.tokType, tt.lit)
		})
	}
}

func TestLexer_ArithmeticOperators(t *testing.T) {
	tests := []struct {
		input   string
		lit     string
		tokType domainLexer.TokenType
	}{
		{"+", "+", domainLexer.TOKEN_PLUS},
		{"-", "-", domainLexer.TOKEN_MINUS},
		{"*", "*", domainLexer.TOKEN_MULTIPLY},
		{"/", "/", domainLexer.TOKEN_DIVIDE},
		{"%", "%", domainLexer.TOKEN_MODULO},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			tokens := lexerTokenize(t, tt.input)
			assertToken(t, tokens[0], tt.tokType, tt.lit)
		})
	}
}

func TestLexer_LogicalKeywords(t *testing.T) {
	tests := []struct {
		input   string
		tokType domainLexer.TokenType
	}{
		{"AND", domainLexer.TOKEN_AND},
		{"OR", domainLexer.TOKEN_OR},
		{"NOT", domainLexer.TOKEN_NOT},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			tokens := lexerTokenize(t, tt.input)
			assertToken(t, tokens[0], tt.tokType, tt.input)
		})
	}
}

func TestLexer_Delimiters(t *testing.T) {
	tokens := lexerTokenize(t, "(,)")
	assertToken(t, tokens[0], domainLexer.TOKEN_LPAREN, "(")
	assertToken(t, tokens[1], domainLexer.TOKEN_COMMA, ",")
	assertToken(t, tokens[2], domainLexer.TOKEN_RPAREN, ")")
}

func TestLexer_StringLiteral(t *testing.T) {
	tokens := lexerTokenize(t, `"银行"`)
	assertToken(t, tokens[0], domainLexer.TOKEN_STRING, "银行")
}

func TestLexer_FullExpression(t *testing.T) {
	input := "ROE > 15 AND PE < 20"
	tokens := lexerTokenize(t, input)

	expected := []struct {
		typ domainLexer.TokenType
		lit string
	}{
		{domainLexer.TOKEN_IDENTIFIER, "ROE"},
		{domainLexer.TOKEN_GT, ">"},
		{domainLexer.TOKEN_NUMBER, "15"},
		{domainLexer.TOKEN_AND, "AND"},
		{domainLexer.TOKEN_IDENTIFIER, "PE"},
		{domainLexer.TOKEN_LT, "<"},
		{domainLexer.TOKEN_NUMBER, "20"},
		{domainLexer.TOKEN_EOF, ""},
	}

	if len(tokens) != len(expected) {
		t.Fatalf("expected %d tokens, got %d", len(expected), len(tokens))
	}

	for i, e := range expected {
		assertToken(t, tokens[i], e.typ, e.lit)
	}
}

func TestLexer_FunctionCall(t *testing.T) {
	input := "MA(CLOSE,5)"
	tokens := lexerTokenize(t, input)

	expected := []struct {
		typ domainLexer.TokenType
		lit string
	}{
		{domainLexer.TOKEN_IDENTIFIER, "MA"},
		{domainLexer.TOKEN_LPAREN, "("},
		{domainLexer.TOKEN_IDENTIFIER, "CLOSE"},
		{domainLexer.TOKEN_COMMA, ","},
		{domainLexer.TOKEN_NUMBER, "5"},
		{domainLexer.TOKEN_RPAREN, ")"},
		{domainLexer.TOKEN_EOF, ""},
	}

	if len(tokens) != len(expected) {
		t.Fatalf("expected %d tokens, got %d", len(expected), len(tokens))
	}

	for i, e := range expected {
		assertToken(t, tokens[i], e.typ, e.lit)
	}
}

func TestLexer_IllegalChar(t *testing.T) {
	tokens := lexerTokenize(t, "@")
	assertToken(t, tokens[0], domainLexer.TOKEN_ILLEGAL, "@")
}

func TestLexer_EmptyInput(t *testing.T) {
	tokens := lexerTokenize(t, "")
	if len(tokens) != 1 {
		t.Fatalf("expected 1 token (EOF), got %d", len(tokens))
	}
	assertToken(t, tokens[0], domainLexer.TOKEN_EOF, "")
}

func TestLexer_MultiLine(t *testing.T) {
	input := "ROE > 15\nAND PE < 20"
	tokens := lexerTokenize(t, input)
	if len(tokens) < 7 {
		t.Fatalf("expected at least 7 tokens, got %d", len(tokens))
	}
	assertToken(t, tokens[0], domainLexer.TOKEN_IDENTIFIER, "ROE")
	assertToken(t, tokens[3], domainLexer.TOKEN_AND, "AND")
}
