package parser

import (
	"context"
	"fmt"
	"strconv"

	domainAST "github.com/agoXQ/QuantLab/app/formula/domain/ast"
	domainErrors "github.com/agoXQ/QuantLab/app/formula/domain/errors"
	domainFunc "github.com/agoXQ/QuantLab/app/formula/domain/function"
	domainLexer "github.com/agoXQ/QuantLab/app/formula/domain/lexer"
	domainParser "github.com/agoXQ/QuantLab/app/formula/domain/parser"
	domainVar "github.com/agoXQ/QuantLab/app/formula/domain/variable"
)

// Precedence levels for operator precedence parsing.
const (
	_ int = iota
	LOWEST
	LOGICAL_OR    // OR
	LOGICAL_AND   // AND
	COMPARISON    // ==, !=, >, <, >=, <=
	SUM           // +, -
	PRODUCT       // *, /, %
	PREFIX        // -X, NOT X
	CALL          // function(X)
)

var precedences = map[domainLexer.TokenType]int{
	domainLexer.TOKEN_OR:       LOGICAL_OR,
	domainLexer.TOKEN_AND:      LOGICAL_AND,
	domainLexer.TOKEN_EQ:       COMPARISON,
	domainLexer.TOKEN_NEQ:      COMPARISON,
	domainLexer.TOKEN_GT:       COMPARISON,
	domainLexer.TOKEN_LT:       COMPARISON,
	domainLexer.TOKEN_GTE:      COMPARISON,
	domainLexer.TOKEN_LTE:      COMPARISON,
	domainLexer.TOKEN_PLUS:     SUM,
	domainLexer.TOKEN_MINUS:    SUM,
	domainLexer.TOKEN_MULTIPLY: PRODUCT,
	domainLexer.TOKEN_DIVIDE:   PRODUCT,
	domainLexer.TOKEN_MODULO:   PRODUCT,
	domainLexer.TOKEN_LPAREN:   CALL,
}

type parser struct {
	funcRegistry domainFunc.Registry
	varRegistry  domainVar.Registry
	tokens       []domainLexer.Token
	pos          int
}

// NewParser creates a new parser instance.
func NewParser(funcRegistry domainFunc.Registry, varRegistry domainVar.Registry) domainParser.Parser {
	return &parser{
		funcRegistry: funcRegistry,
		varRegistry:  varRegistry,
	}
}

func (p *parser) Parse(_ context.Context, tokens []domainLexer.Token) (domainAST.Node, error) {
	p.tokens = tokens
	p.pos = 0

	// Check if this is a multi-statement program (contains := or ; or : at statement level)
	if p.hasProgramStructure() {
		return p.parseProgram()
	}

	node, err := p.parseExpression(LOWEST)
	if err != nil {
		return nil, err
	}

	if !p.currentIs(domainLexer.TOKEN_EOF) {
		return nil, domainErrors.NewErrorWithPos(
			domainErrors.ErrSyntaxError,
			fmt.Sprintf("unexpected token: %s", p.current().Literal),
			p.current().Pos,
		)
	}

	return node, nil
}

func (p *parser) hasProgramStructure() bool {
	for i := 0; i < len(p.tokens); i++ {
		tok := p.tokens[i]
		if tok.Type == domainLexer.TOKEN_ASSIGN || tok.Type == domainLexer.TOKEN_SEMICOLON || tok.Type == domainLexer.TOKEN_COLON {
			return true
		}
	}
	return false
}

func (p *parser) parseProgram() (domainAST.Node, error) {
	var stmts []domainAST.Node

	for !p.currentIs(domainLexer.TOKEN_EOF) {
		p.skipSemicolons()

		if p.currentIs(domainLexer.TOKEN_EOF) {
			break
		}

		// Check for named result: "name: expression"
		// We peek ahead: if we see IDENTIFIER followed by COLON, it's a named result
		if p.currentIs(domainLexer.TOKEN_IDENTIFIER) && p.peekIs(domainLexer.TOKEN_COLON) {
			name := p.current().Literal
			p.nextToken() // consume identifier
			p.nextToken() // consume colon

			expr, err := p.parseExpression(LOWEST)
			if err != nil {
				return nil, err
			}

			// Treat named result as an assignment
			stmts = append(stmts, &domainAST.Assignment{Name: name, Value: expr})
			p.skipSemicolons()
			continue
		}

		// Check for assignment: "name := expression"
		if p.currentIs(domainLexer.TOKEN_IDENTIFIER) && p.peekIs(domainLexer.TOKEN_ASSIGN) {
			name := p.current().Literal
			p.nextToken() // consume identifier
			p.nextToken() // consume :=

			expr, err := p.parseExpression(LOWEST)
			if err != nil {
				return nil, err
			}

			stmts = append(stmts, &domainAST.Assignment{Name: name, Value: expr})
			p.skipSemicolons()
			continue
		}

		// Plain expression statement
		expr, err := p.parseExpression(LOWEST)
		if err != nil {
			return nil, err
		}
		stmts = append(stmts, expr)
		p.skipSemicolons()
	}

	if len(stmts) == 0 {
		return nil, domainErrors.NewError(domainErrors.ErrSyntaxError, "empty program")
	}

	if len(stmts) == 1 {
		return stmts[0], nil
	}

	return &domainAST.Program{Statements: stmts}, nil
}

func (p *parser) skipSemicolons() {
	for p.currentIs(domainLexer.TOKEN_SEMICOLON) {
		p.nextToken()
	}
}

func (p *parser) parseExpression(precedence int) (domainAST.Node, error) {
	prefix, err := p.parsePrefix()
	if err != nil {
		return nil, err
	}

	for !p.currentIs(domainLexer.TOKEN_EOF) && precedence < p.peekPrecedence() {
		infix, err := p.parseInfix(prefix)
		if err != nil {
			return nil, err
		}
		prefix = infix
	}

	return prefix, nil
}

func (p *parser) parsePrefix() (domainAST.Node, error) {
	tok := p.current()

	switch tok.Type {
	case domainLexer.TOKEN_IDENTIFIER:
		p.nextToken()
		if p.currentIs(domainLexer.TOKEN_LPAREN) {
			return p.parseFunctionCall(tok.Literal)
		}
		return &domainAST.Identifier{Name: tok.Literal}, nil

	case domainLexer.TOKEN_NUMBER:
		p.nextToken()
		val, err := strconv.ParseFloat(tok.Literal, 64)
		if err != nil {
			return nil, domainErrors.NewErrorWithPos(domainErrors.ErrSyntaxError, "invalid number: "+tok.Literal, tok.Pos)
		}
		return &domainAST.NumberLiteral{Value: val}, nil

	case domainLexer.TOKEN_STRING:
		p.nextToken()
		return &domainAST.StringLiteral{Value: tok.Literal}, nil

	case domainLexer.TOKEN_MINUS:
		return p.parsePrefixExpression("-")

	case domainLexer.TOKEN_NOT:
		return p.parsePrefixExpression("NOT")

	case domainLexer.TOKEN_LPAREN:
		p.nextToken()
		node, err := p.parseExpression(LOWEST)
		if err != nil {
			return nil, err
		}
		if !p.expectPeek(domainLexer.TOKEN_RPAREN) {
			return nil, domainErrors.NewErrorWithPos(domainErrors.ErrSyntaxError, "missing closing parenthesis", tok.Pos)
		}
		return node, nil

	case domainLexer.TOKEN_PLUS:
		p.nextToken()
		return p.parseExpression(PREFIX)

	default:
		return nil, domainErrors.NewErrorWithPos(
			domainErrors.ErrSyntaxError,
			fmt.Sprintf("unexpected token: %s", tok.Literal),
			tok.Pos,
		)
	}
}

func (p *parser) parsePrefixExpression(operator string) (domainAST.Node, error) {
	p.nextToken()
	operand, err := p.parseExpression(PREFIX)
	if err != nil {
		return nil, err
	}
	return &domainAST.UnaryExpression{Operator: operator, Operand: operand}, nil
}

func (p *parser) parseInfix(left domainAST.Node) (domainAST.Node, error) {
	tok := p.current()
	p.nextToken()

	switch tok.Type {
	case domainLexer.TOKEN_PLUS, domainLexer.TOKEN_MINUS, domainLexer.TOKEN_MULTIPLY,
		domainLexer.TOKEN_DIVIDE, domainLexer.TOKEN_MODULO,
		domainLexer.TOKEN_GT, domainLexer.TOKEN_LT, domainLexer.TOKEN_GTE, domainLexer.TOKEN_LTE,
		domainLexer.TOKEN_EQ, domainLexer.TOKEN_NEQ,
		domainLexer.TOKEN_AND, domainLexer.TOKEN_OR:
		prec := precedences[tok.Type]
		right, err := p.parseExpression(prec)
		if err != nil {
			return nil, err
		}
		return &domainAST.BinaryExpression{
			Left:     left,
			Operator: tok.Literal,
			Right:    right,
		}, nil

	default:
		return nil, domainErrors.NewErrorWithPos(
			domainErrors.ErrSyntaxError,
			fmt.Sprintf("unexpected operator: %s", tok.Literal),
			tok.Pos,
		)
	}
}

func (p *parser) parseFunctionCall(name string) (domainAST.Node, error) {
	canonicalName, exists := p.funcRegistry.ResolveName(name)
	if !exists {
		return nil, domainErrors.NewErrorWithPos(
			domainErrors.ErrUnknownFunction,
			fmt.Sprintf("unknown function: %s", name),
			p.current().Pos,
		)
	}

	p.nextToken() // consume '('

	var args []domainAST.Node
	if !p.currentIs(domainLexer.TOKEN_RPAREN) {
		arg, err := p.parseExpression(LOWEST)
		if err != nil {
			return nil, err
		}
		args = append(args, arg)

		for p.currentIs(domainLexer.TOKEN_COMMA) {
			p.nextToken()
			arg, err := p.parseExpression(LOWEST)
			if err != nil {
				return nil, err
			}
			args = append(args, arg)
		}
	}

	if !p.expectPeek(domainLexer.TOKEN_RPAREN) {
		return nil, domainErrors.NewErrorWithPos(
			domainErrors.ErrSyntaxError,
			fmt.Sprintf("missing closing parenthesis after function %s", name),
			p.current().Pos,
		)
	}

	def, _ := p.funcRegistry.GetFunction(canonicalName)
	if len(def.Args) > 0 {
		requiredCount := 0
		for _, arg := range def.Args {
			if arg.Required {
				requiredCount++
			}
		}
		if len(args) < requiredCount {
			return nil, domainErrors.NewErrorWithPos(
				domainErrors.ErrInvalidArgCount,
				fmt.Sprintf("function %s expects at least %d arguments, got %d", canonicalName, requiredCount, len(args)),
				p.current().Pos,
			)
		}
		if len(args) > len(def.Args) {
			return nil, domainErrors.NewErrorWithPos(
				domainErrors.ErrInvalidArgCount,
				fmt.Sprintf("function %s expects at most %d arguments, got %d", canonicalName, len(def.Args), len(args)),
				p.current().Pos,
			)
		}
	}

	return &domainAST.FunctionCall{Name: canonicalName, Args: args}, nil
}

func (p *parser) current() domainLexer.Token {
	if p.pos >= len(p.tokens) {
		return domainLexer.Token{Type: domainLexer.TOKEN_EOF}
	}
	return p.tokens[p.pos]
}

func (p *parser) currentIs(t domainLexer.TokenType) bool {
	return p.current().Type == t
}

func (p *parser) nextToken() {
	if p.pos < len(p.tokens) {
		p.pos++
	}
}

func (p *parser) expectPeek(t domainLexer.TokenType) bool {
	if p.currentIs(t) {
		p.nextToken()
		return true
	}
	return false
}

func (p *parser) peekPrecedence() int {
	if p.pos >= len(p.tokens) {
		return LOWEST
	}
	if prec, ok := precedences[p.tokens[p.pos].Type]; ok {
		return prec
	}
	return LOWEST
}

func (p *parser) peekIs(t domainLexer.TokenType) bool {
	nextPos := p.pos + 1
	if nextPos >= len(p.tokens) {
		return false
	}
	return p.tokens[nextPos].Type == t
}
