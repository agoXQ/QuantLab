package lexer

// TokenType represents the type of a lexical token.
type TokenType int

// Token types for the Formula DSL.
const (
	// Special tokens
	TOKEN_ILLEGAL TokenType = iota
	TOKEN_EOF

	// Identifiers and literals
	TOKEN_IDENTIFIER
	TOKEN_NUMBER
	TOKEN_STRING

	// Arithmetic operators
	TOKEN_PLUS     // +
	TOKEN_MINUS    // -
	TOKEN_MULTIPLY // *
	TOKEN_DIVIDE   // /
	TOKEN_MODULO   // %

	// Comparison operators
	TOKEN_GT  // >
	TOKEN_LT  // <
	TOKEN_GTE // >=
	TOKEN_LTE // <=
	TOKEN_EQ  // ==
	TOKEN_NEQ // !=

	// Logical operators (also recognized as keywords)
	TOKEN_AND
	TOKEN_OR
	TOKEN_NOT

	// Delimiters
	TOKEN_LPAREN // (
	TOKEN_RPAREN // )
	TOKEN_COMMA  // ,

	// Assignment and statement
	TOKEN_ASSIGN  // :=
	TOKEN_COLON   // :
	TOKEN_SEMICOLON // ;
)

// Token represents a single lexical token with position information.
type Token struct {
	Type    TokenType
	Literal string
	Pos     int
	Line    int
	Column  int
}

// String returns a human-readable representation of the token type.
func (t TokenType) String() string {
	switch t {
	case TOKEN_ILLEGAL:
		return "ILLEGAL"
	case TOKEN_EOF:
		return "EOF"
	case TOKEN_IDENTIFIER:
		return "IDENTIFIER"
	case TOKEN_NUMBER:
		return "NUMBER"
	case TOKEN_STRING:
		return "STRING"
	case TOKEN_PLUS:
		return "PLUS"
	case TOKEN_MINUS:
		return "MINUS"
	case TOKEN_MULTIPLY:
		return "MULTIPLY"
	case TOKEN_DIVIDE:
		return "DIVIDE"
	case TOKEN_MODULO:
		return "MODULO"
	case TOKEN_GT:
		return "GT"
	case TOKEN_LT:
		return "LT"
	case TOKEN_GTE:
		return "GTE"
	case TOKEN_LTE:
		return "LTE"
	case TOKEN_EQ:
		return "EQ"
	case TOKEN_NEQ:
		return "NEQ"
	case TOKEN_AND:
		return "AND"
	case TOKEN_OR:
		return "OR"
	case TOKEN_NOT:
		return "NOT"
	case TOKEN_LPAREN:
		return "LPAREN"
	case TOKEN_RPAREN:
		return "RPAREN"
	case TOKEN_COMMA:
		return "COMMA"
	case TOKEN_ASSIGN:
		return "ASSIGN"
	case TOKEN_COLON:
		return "COLON"
	case TOKEN_SEMICOLON:
		return "SEMICOLON"
	default:
		return "UNKNOWN"
	}
}
