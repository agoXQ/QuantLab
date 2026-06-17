package compiler

import (
	"context"
	"encoding/json"

	"github.com/agoXQ/QuantLab/app/formula/domain/ast"
)

// PlanType represents the type of execution plan.
type PlanType string

const (
	PlanTypeFilter PlanType = "FILTER"
	PlanTypeSignal PlanType = "SIGNAL"
	PlanTypeSort   PlanType = "SORT"
	PlanTypeValue  PlanType = "VALUE"
)

// ExecutionPlan represents the compiled execution plan for a formula.
type ExecutionPlan struct {
	PlanType  PlanType  `json:"plan_type"`
	Root      ast.Node  `json:"root"`
	Optimized bool      `json:"optimized"`
}

// MarshalJSON implements json.Marshaler for ExecutionPlan.
func (p *ExecutionPlan) MarshalJSON() ([]byte, error) {
	rootJSON, err := json.Marshal(p.Root)
	if err != nil {
		return nil, err
	}
	return json.Marshal(map[string]interface{}{
		"plan_type": p.PlanType,
		"root":      json.RawMessage(rootJSON),
		"optimized": p.Optimized,
	})
}

// UnmarshalJSON implements json.Unmarshaler for ExecutionPlan.
func (p *ExecutionPlan) UnmarshalJSON(data []byte) error {
	var raw struct {
		PlanType  PlanType         `json:"plan_type"`
		Root      json.RawMessage  `json:"root"`
		Optimized bool             `json:"optimized"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	p.PlanType = raw.PlanType
	p.Optimized = raw.Optimized

	// Determine the root node type from the JSON
	var typeOnly struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(raw.Root, &typeOnly); err != nil {
		return err
	}
	p.Root = unmarshalASTNode(typeOnly.Type, raw.Root)
	return nil
}

// unmarshalASTNode unmarshals JSON into the correct AST node type.
func unmarshalASTNode(typeName string, data []byte) ast.Node {
	var node ast.Node
	switch typeName {
	case "BinaryExpression":
		node = &ast.BinaryExpression{}
	case "UnaryExpression":
		node = &ast.UnaryExpression{}
	case "FunctionCall":
		node = &ast.FunctionCall{}
	case "Identifier":
		node = &ast.Identifier{}
	case "NumberLiteral":
		node = &ast.NumberLiteral{}
	case "StringLiteral":
		node = &ast.StringLiteral{}
	case "BoolLiteral":
		node = &ast.BoolLiteral{}
	case "Assignment":
		node = &ast.Assignment{}
	case "Program":
		node = &ast.Program{}
	default:
		return nil
	}
	if err := json.Unmarshal(data, node); err != nil {
		return nil
	}
	return node
}

// Planner defines the interface for generating execution plans from AST.
// Implementations must be safe for concurrent use.
type Planner interface {
	// Plan generates an execution plan from an AST node.
	Plan(ctx context.Context, node ast.Node) (*ExecutionPlan, error)
}
