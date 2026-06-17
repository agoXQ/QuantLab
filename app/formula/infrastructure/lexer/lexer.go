package lexer

import (
	"context"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	domainLexer "github.com/agoXQ/QuantLab/app/formula/domain/lexer"
)

// lexer implements the domain Lexer interface.
type lexer struct {
	input   string
	pos     int
	readPos int
	ch      rune
	line    int
	column  int
}

// NewLexer creates a new lexer instance.
func NewLexer() domainLexer.Lexer {
	return &lexer{}
}

func (l *lexer) Tokenize(_ context.Context, input string) ([]domainLexer.Token, error) {
	l.input = input
	l.pos = 0
	l.readPos = 0
	l.line = 1
	l.column = 1
	l.readChar()

	var tokens []domainLexer.Token
	for l.ch != 0 {
		l.skipWhitespace()

		if l.ch == 0 {
			break
		}

		pos := l.pos
		col := l.column

		switch {
		case l.ch == '(':
			tokens = append(tokens, l.makeToken(domainLexer.TOKEN_LPAREN, "(", pos, col))
			l.readChar()
		case l.ch == ')':
			tokens = append(tokens, l.makeToken(domainLexer.TOKEN_RPAREN, ")", pos, col))
			l.readChar()
		case l.ch == ',':
			tokens = append(tokens, l.makeToken(domainLexer.TOKEN_COMMA, ",", pos, col))
			l.readChar()
		case l.ch == '+':
			tokens = append(tokens, l.makeToken(domainLexer.TOKEN_PLUS, "+", pos, col))
			l.readChar()
		case l.ch == '-':
			tokens = append(tokens, l.makeToken(domainLexer.TOKEN_MINUS, "-", pos, col))
			l.readChar()
		case l.ch == '*':
			tokens = append(tokens, l.makeToken(domainLexer.TOKEN_MULTIPLY, "*", pos, col))
			l.readChar()
		case l.ch == '/':
			tokens = append(tokens, l.makeToken(domainLexer.TOKEN_DIVIDE, "/", pos, col))
			l.readChar()
		case l.ch == '%':
			if l.peekChar() == '=' {
				l.readChar()
				l.readChar()
				continue
			}
			tokens = append(tokens, l.makeToken(domainLexer.TOKEN_MODULO, "%", pos, col))
			l.readChar()
		case l.ch == ':':
			if l.peekChar() == '=' {
				tokens = append(tokens, l.makeToken(domainLexer.TOKEN_ASSIGN, ":=", pos, col))
				l.readChar()
				l.readChar()
			} else {
				tokens = append(tokens, l.makeToken(domainLexer.TOKEN_COLON, ":", pos, col))
				l.readChar()
			}
		case l.ch == ';':
			tokens = append(tokens, l.makeToken(domainLexer.TOKEN_SEMICOLON, ";", pos, col))
			l.readChar()
		case l.ch == '{':
			// Block comment: skip everything until matching }
			l.readChar()
			for l.ch != 0 && l.ch != '}' {
				l.readChar()
			}
			if l.ch == '}' {
				l.readChar()
			}
		case l.ch == '>':
			if l.peekChar() == '=' {
				tokens = append(tokens, l.makeToken(domainLexer.TOKEN_GTE, ">=", pos, col))
				l.readChar()
				l.readChar()
			} else {
				tokens = append(tokens, l.makeToken(domainLexer.TOKEN_GT, ">", pos, col))
				l.readChar()
			}
		case l.ch == '<':
			if l.peekChar() == '=' {
				tokens = append(tokens, l.makeToken(domainLexer.TOKEN_LTE, "<=", pos, col))
				l.readChar()
				l.readChar()
			} else {
				tokens = append(tokens, l.makeToken(domainLexer.TOKEN_LT, "<", pos, col))
				l.readChar()
			}
		case l.ch == '=':
			if l.peekChar() == '=' {
				tokens = append(tokens, l.makeToken(domainLexer.TOKEN_EQ, "==", pos, col))
				l.readChar()
				l.readChar()
			} else {
				tokens = append(tokens, l.makeToken(domainLexer.TOKEN_ILLEGAL, "=", pos, col))
				l.readChar()
			}
		case l.ch == '!':
			if l.peekChar() == '=' {
				tokens = append(tokens, l.makeToken(domainLexer.TOKEN_NEQ, "!=", pos, col))
				l.readChar()
				l.readChar()
			} else {
				tokens = append(tokens, l.makeToken(domainLexer.TOKEN_ILLEGAL, "!", pos, col))
				l.readChar()
			}
		case l.ch == '"':
			tok, err := l.readString(pos, col)
			if err != nil {
				return nil, err
			}
			tokens = append(tokens, tok)
		case unicode.IsLetter(l.ch):
			tok := l.readIdentifier(pos, col)
			tokens = append(tokens, tok)
		case unicode.IsDigit(l.ch):
			tok := l.readNumber(pos, col)
			tokens = append(tokens, tok)
		default:
			tokens = append(tokens, l.makeToken(domainLexer.TOKEN_ILLEGAL, string(l.ch), pos, col))
			l.readChar()
		}
	}

	tokens = append(tokens, l.makeToken(domainLexer.TOKEN_EOF, "", l.pos, l.column))
	return tokens, nil
}

func (l *lexer) readChar() {
	if l.readPos >= len(l.input) {
		l.pos = l.readPos
		l.ch = 0
		return
	}
	r, size := utf8.DecodeRuneInString(l.input[l.readPos:])
	l.ch = r
	l.pos = l.readPos
	l.readPos += size
	if l.ch == '\n' {
		l.line++
		l.column = 1
	} else {
		l.column++
	}
}

func (l *lexer) peekChar() rune {
	if l.readPos >= len(l.input) {
		return 0
	}
	r, _ := utf8.DecodeRuneInString(l.input[l.readPos:])
	return r
}

func (l *lexer) skipWhitespace() {
	for l.ch != 0 && (l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r') {
		l.readChar()
	}
}

func (l *lexer) readIdentifier(pos, col int) domainLexer.Token {
	start := l.pos
	for l.ch != 0 && (unicode.IsLetter(l.ch) || unicode.IsDigit(l.ch) || l.ch == '_') {
		l.readChar()
	}
	literal := l.input[start:l.pos]

	// Check for keywords
	upper := strings.ToUpper(literal)
	switch upper {
	case "AND":
		return l.makeToken(domainLexer.TOKEN_AND, literal, pos, col)
	case "OR":
		return l.makeToken(domainLexer.TOKEN_OR, literal, pos, col)
	case "NOT":
		return l.makeToken(domainLexer.TOKEN_NOT, literal, pos, col)
	case "TRUE":
		return l.makeToken(domainLexer.TOKEN_IDENTIFIER, literal, pos, col)
	case "FALSE":
		return l.makeToken(domainLexer.TOKEN_IDENTIFIER, literal, pos, col)
	default:
		return l.makeToken(domainLexer.TOKEN_IDENTIFIER, literal, pos, col)
	}
}

func (l *lexer) readNumber(pos, col int) domainLexer.Token {
	start := l.pos
	for l.ch != 0 && unicode.IsDigit(l.ch) {
		l.readChar()
	}
	// Check for decimal part
	if l.ch == '.' && l.peekChar() != 0 && unicode.IsDigit(l.peekChar()) {
		l.readChar() // consume '.'
		for l.ch != 0 && unicode.IsDigit(l.ch) {
			l.readChar()
		}
	}
	literal := l.input[start:l.pos]

	// Validate the number format
	_, err := strconv.ParseFloat(literal, 64)
	if err != nil {
		return l.makeToken(domainLexer.TOKEN_ILLEGAL, literal, pos, col)
	}

	return l.makeToken(domainLexer.TOKEN_NUMBER, literal, pos, col)
}

func (l *lexer) readString(pos, col int) (domainLexer.Token, error) {
	l.readChar() // consume opening "
	start := l.pos
	for l.ch != 0 && l.ch != '"' {
		l.readChar()
	}
	if l.ch == 0 {
		return l.makeToken(domainLexer.TOKEN_ILLEGAL, l.input[pos:l.readPos], pos, col), nil
	}
	literal := l.input[start:l.pos]
	l.readChar() // consume closing "
	return l.makeToken(domainLexer.TOKEN_STRING, literal, pos, col), nil
}

func (l *lexer) makeToken(t domainLexer.TokenType, literal string, pos, col int) domainLexer.Token {
	return domainLexer.Token{
		Type:    t,
		Literal: literal,
		Pos:     pos,
		Line:    l.line,
		Column:  col,
	}
}

// isFollowingOperator checks if the current char is an operator that follows a value/expression.
func (l *lexer) isFollowingOperator() bool {
	return true
}
