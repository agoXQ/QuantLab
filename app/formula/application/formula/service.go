package formula

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"

	domainAST "github.com/agoXQ/QuantLab/app/formula/domain/ast"
	domainCompiler "github.com/agoXQ/QuantLab/app/formula/domain/compiler"
	domainErrors "github.com/agoXQ/QuantLab/app/formula/domain/errors"
	domainFunc "github.com/agoXQ/QuantLab/app/formula/domain/function"
	domainLexer "github.com/agoXQ/QuantLab/app/formula/domain/lexer"
	domainOptimizer "github.com/agoXQ/QuantLab/app/formula/domain/optimizer"
	domainParser "github.com/agoXQ/QuantLab/app/formula/domain/parser"
	domainValidator "github.com/agoXQ/QuantLab/app/formula/domain/validator"
)

// CompileResult contains the full result of compiling a formula.
type CompileResult struct {
	AST       domainAST.Node                `json:"-"`
	Plan      *domainCompiler.ExecutionPlan `json:"plan"`
	Valid     bool                          `json:"valid"`
	ErrorCode int                           `json:"error_code,omitempty"`
	ErrorMsg  string                        `json:"error_msg,omitempty"`
}

// MarshalJSON implements json.Marshaler for CompileResult.
func (r *CompileResult) MarshalJSON() ([]byte, error) {
	planJSON, _ := json.Marshal(r.Plan)
	astJSON, _ := json.Marshal(r.AST)
	return json.Marshal(map[string]interface{}{
		"ast":        json.RawMessage(astJSON),
		"plan":       json.RawMessage(planJSON),
		"valid":      r.Valid,
		"error_code": r.ErrorCode,
		"error_msg":  r.ErrorMsg,
	})
}

// Service defines the application-level interface for the Formula Engine.
type Service interface {
	Validate(ctx context.Context, formula string) (*domainValidator.ValidationResult, error)
	Compile(ctx context.Context, formula string) (*CompileResult, error)
	GetAST(ctx context.Context, formula string) (domainAST.Node, error)
	ListFunctions(ctx context.Context) ([]domainFunc.FunctionDefinition, error)
	GetFunction(ctx context.Context, name string) (*domainFunc.FunctionDefinition, error)
	FormulaHash(formula string) string
}

// service implements the Service interface.
type service struct {
	lexer     domainLexer.Lexer
	parser    domainParser.Parser
	validator domainValidator.Validator
	optimizer domainOptimizer.Optimizer
	planner   domainCompiler.Planner
	funcReg   domainFunc.Registry
}

// NewService creates a new Formula Engine application service.
func NewService(
	lexer domainLexer.Lexer,
	parser domainParser.Parser,
	validator domainValidator.Validator,
	optimizer domainOptimizer.Optimizer,
	planner domainCompiler.Planner,
	funcReg domainFunc.Registry,
) Service {
	return &service{
		lexer:     lexer,
		parser:    parser,
		validator: validator,
		optimizer: optimizer,
		planner:   planner,
		funcReg:   funcReg,
	}
}

func (s *service) Validate(ctx context.Context, formula string) (*domainValidator.ValidationResult, error) {
	tokens, err := s.lexer.Tokenize(ctx, formula)
	if err != nil {
		return &domainValidator.ValidationResult{
			Valid: false,
			Errors: []domainValidator.ValidationError{{
				Code:     domainErrors.ErrLexerError,
				CodeStr:  "LEXER_ERROR",
				Message:  err.Error(),
				Severity: domainValidator.SeverityError,
			}},
		}, nil
	}

	node, err := s.parser.Parse(ctx, tokens)
	if err != nil {
		return &domainValidator.ValidationResult{
			Valid: false,
			Errors: []domainValidator.ValidationError{{
				Code:     domainErrors.ErrParseError,
				CodeStr:  "PARSE_ERROR",
				Message:  err.Error(),
				Severity: domainValidator.SeverityError,
			}},
		}, nil
	}

	return s.validator.Validate(ctx, node)
}

func (s *service) Compile(ctx context.Context, formula string) (*CompileResult, error) {
	tokens, err := s.lexer.Tokenize(ctx, formula)
	if err != nil {
		return nil, err
	}

	node, err := s.parser.Parse(ctx, tokens)
	if err != nil {
		return nil, err
	}

	valResult, err := s.validator.Validate(ctx, node)
	if err != nil {
		return nil, err
	}
	if !valResult.Valid {
		firstCode := 0
		firstMsg := "formula validation failed"
		if len(valResult.Errors) > 0 {
			firstCode = valResult.Errors[0].Code
			firstMsg = valResult.Errors[0].Message
		}
		return &CompileResult{
			AST:       node,
			Valid:     false,
			ErrorCode: firstCode,
			ErrorMsg:  firstMsg,
		}, nil
	}

	optimized, err := s.optimizer.Optimize(ctx, node)
	if err != nil {
		return nil, err
	}

	plan, err := s.planner.Plan(ctx, optimized)
	if err != nil {
		return nil, err
	}

	return &CompileResult{
		AST:   optimized,
		Plan:  plan,
		Valid: true,
	}, nil
}

func (s *service) GetAST(ctx context.Context, formula string) (domainAST.Node, error) {
	tokens, err := s.lexer.Tokenize(ctx, formula)
	if err != nil {
		return nil, err
	}

	return s.parser.Parse(ctx, tokens)
}

func (s *service) ListFunctions(_ context.Context) ([]domainFunc.FunctionDefinition, error) {
	return s.funcReg.ListFunctions(), nil
}

func (s *service) GetFunction(_ context.Context, name string) (*domainFunc.FunctionDefinition, error) {
	def, ok := s.funcReg.GetFunction(name)
	if !ok {
		return nil, nil
	}
	return &def, nil
}

func (s *service) FormulaHash(formula string) string {
	h := sha256.Sum256([]byte(formula))
	return hex.EncodeToString(h[:])
}
