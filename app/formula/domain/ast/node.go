package ast

import (
	"encoding/json"
	"fmt"
)

// Node is the base interface for all AST nodes.
type Node interface {
	String() string
	Type() string
}

// nodeJSON is a helper struct for JSON serialization that includes the type field.
type nodeJSON struct {
	Type     string            `json:"type"`
	Left     json.RawMessage   `json:"left,omitempty"`
	Right    json.RawMessage   `json:"right,omitempty"`
	Operator string            `json:"operator,omitempty"`
	Operand  json.RawMessage   `json:"operand,omitempty"`
	Name     string            `json:"name,omitempty"`
	Args     []json.RawMessage `json:"args,omitempty"`
	Value    interface{}       `json:"value,omitempty"`
	Statements []json.RawMessage `json:"statements,omitempty"`
}

// --- BinaryExpression ---

type BinaryExpression struct {
	Left     Node   `json:"-"`
	Operator string `json:"-"`
	Right    Node   `json:"-"`
}

func (e *BinaryExpression) String() string {
	return fmt.Sprintf("(%s %s %s)", e.Left.String(), e.Operator, e.Right.String())
}

func (e *BinaryExpression) Type() string { return "BinaryExpression" }

func (e *BinaryExpression) MarshalJSON() ([]byte, error) {
	left, _ := json.Marshal(e.Left)
	right, _ := json.Marshal(e.Right)
	return json.Marshal(nodeJSON{
		Type:     e.Type(),
		Left:     left,
		Right:    right,
		Operator: e.Operator,
	})
}

func (e *BinaryExpression) UnmarshalJSON(data []byte) error {
	var j nodeJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	e.Operator = j.Operator
	if len(j.Left) > 0 {
		e.Left = unmarshalNodeJSON(j.Left)
	}
	if len(j.Right) > 0 {
		e.Right = unmarshalNodeJSON(j.Right)
	}
	return nil
}

// --- UnaryExpression ---

type UnaryExpression struct {
	Operator string `json:"-"`
	Operand  Node   `json:"-"`
}

func (e *UnaryExpression) String() string {
	return fmt.Sprintf("(%s%s)", e.Operator, e.Operand.String())
}

func (e *UnaryExpression) Type() string { return "UnaryExpression" }

func (e *UnaryExpression) MarshalJSON() ([]byte, error) {
	operand, _ := json.Marshal(e.Operand)
	return json.Marshal(nodeJSON{
		Type:     e.Type(),
		Operator: e.Operator,
		Operand:  operand,
	})
}

func (e *UnaryExpression) UnmarshalJSON(data []byte) error {
	var j nodeJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	e.Operator = j.Operator
	if len(j.Operand) > 0 {
		e.Operand = unmarshalNodeJSON(j.Operand)
	}
	return nil
}

// --- FunctionCall ---

type FunctionCall struct {
	Name string `json:"-"`
	Args []Node `json:"-"`
}

func (f *FunctionCall) String() string {
	s := f.Name + "("
	for i, arg := range f.Args {
		if i > 0 {
			s += ", "
		}
		s += arg.String()
	}
	s += ")"
	return s
}

func (f *FunctionCall) Type() string { return "FunctionCall" }

func (f *FunctionCall) MarshalJSON() ([]byte, error) {
	args := make([]json.RawMessage, len(f.Args))
	for i, arg := range f.Args {
		args[i], _ = json.Marshal(arg)
	}
	return json.Marshal(nodeJSON{
		Type: f.Type(),
		Name: f.Name,
		Args: args,
	})
}

func (f *FunctionCall) UnmarshalJSON(data []byte) error {
	var j nodeJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	f.Name = j.Name
	f.Args = make([]Node, len(j.Args))
	for i, arg := range j.Args {
		f.Args[i] = unmarshalNodeJSON(arg)
	}
	return nil
}

// --- Identifier ---

type Identifier struct {
	Name string `json:"-"`
}

func (i *Identifier) String() string { return i.Name }

func (i *Identifier) Type() string { return "Identifier" }

func (i *Identifier) MarshalJSON() ([]byte, error) {
	return json.Marshal(nodeJSON{
		Type: i.Type(),
		Name: i.Name,
	})
}

func (i *Identifier) UnmarshalJSON(data []byte) error {
	var j nodeJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	i.Name = j.Name
	return nil
}

// --- NumberLiteral ---

type NumberLiteral struct {
	Value float64 `json:"-"`
}

func (n *NumberLiteral) String() string {
	return fmt.Sprintf("%v", n.Value)
}

func (n *NumberLiteral) Type() string { return "NumberLiteral" }

func (n *NumberLiteral) MarshalJSON() ([]byte, error) {
	return json.Marshal(nodeJSON{
		Type:  n.Type(),
		Value: n.Value,
	})
}

func (n *NumberLiteral) UnmarshalJSON(data []byte) error {
	var j nodeJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if v, ok := j.Value.(float64); ok {
		n.Value = v
	}
	return nil
}

// --- StringLiteral ---

type StringLiteral struct {
	Value string `json:"-"`
}

func (s *StringLiteral) String() string {
	return fmt.Sprintf("%q", s.Value)
}

func (s *StringLiteral) Type() string { return "StringLiteral" }

func (s *StringLiteral) MarshalJSON() ([]byte, error) {
	return json.Marshal(nodeJSON{
		Type:  s.Type(),
		Value: s.Value,
	})
}

func (s *StringLiteral) UnmarshalJSON(data []byte) error {
	var j nodeJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if v, ok := j.Value.(string); ok {
		s.Value = v
	}
	return nil
}

// --- BoolLiteral ---

type BoolLiteral struct {
	Value bool `json:"-"`
}

func (b *BoolLiteral) String() string {
	if b.Value {
		return "TRUE"
	}
	return "FALSE"
}

func (b *BoolLiteral) Type() string { return "BoolLiteral" }

func (b *BoolLiteral) MarshalJSON() ([]byte, error) {
	return json.Marshal(nodeJSON{
		Type:  b.Type(),
		Value: b.Value,
	})
}

func (b *BoolLiteral) UnmarshalJSON(data []byte) error {
	var j nodeJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if v, ok := j.Value.(bool); ok {
		b.Value = v
	}
	return nil
}

// unmarshalNodeJSON unmarshals a JSON byte slice into the appropriate Node type.
func unmarshalNodeJSON(data []byte) Node {
	var typeOnly struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(data, &typeOnly); err != nil {
		return nil
	}

	var node Node
	switch typeOnly.Type {
	case "BinaryExpression":
		node = &BinaryExpression{}
	case "UnaryExpression":
		node = &UnaryExpression{}
	case "FunctionCall":
		node = &FunctionCall{}
	case "Identifier":
		node = &Identifier{}
	case "NumberLiteral":
		node = &NumberLiteral{}
	case "StringLiteral":
		node = &StringLiteral{}
	case "BoolLiteral":
		node = &BoolLiteral{}
	case "Assignment":
		node = &Assignment{}
	case "Program":
		node = &Program{}
	default:
		return nil
	}

	if err := json.Unmarshal(data, node); err != nil {
		return nil
	}
	return node
}

// --- Assignment ---

type Assignment struct {
	Name  string `json:"-"`
	Value Node   `json:"-"`
}

func (a *Assignment) String() string {
	return fmt.Sprintf("%s := %s", a.Name, a.Value.String())
}

func (a *Assignment) Type() string { return "Assignment" }

func (a *Assignment) MarshalJSON() ([]byte, error) {
	val, _ := json.Marshal(a.Value)
	return json.Marshal(nodeJSON{
		Type:  a.Type(),
		Name:  a.Name,
		Value: json.RawMessage(val),
	})
}

func (a *Assignment) UnmarshalJSON(data []byte) error {
	var j nodeJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	a.Name = j.Name
	if j.Value != nil {
		if raw, ok := j.Value.(json.RawMessage); ok {
			a.Value = unmarshalNodeJSON(raw)
		} else {
			// Try to convert
			b, _ := json.Marshal(j.Value)
			a.Value = unmarshalNodeJSON(b)
		}
	}
	return nil
}

// --- Program ---

type Program struct {
	Statements []Node `json:"-"`
}

func (p *Program) String() string {
	s := ""
	for i, stmt := range p.Statements {
		if i > 0 {
			s += "; "
		}
		s += stmt.String()
	}
	return s
}

func (p *Program) Type() string { return "Program" }

func (p *Program) MarshalJSON() ([]byte, error) {
	stmts := make([]json.RawMessage, len(p.Statements))
	for i, stmt := range p.Statements {
		stmts[i], _ = json.Marshal(stmt)
	}
	return json.Marshal(nodeJSON{
		Type:       p.Type(),
		Statements: stmts,
	})
}

func (p *Program) UnmarshalJSON(data []byte) error {
	var j nodeJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	p.Statements = make([]Node, len(j.Statements))
	for i, stmt := range j.Statements {
		p.Statements[i] = unmarshalNodeJSON(stmt)
	}
	return nil
}

// Ensure all types implement Node.
var _ Node = (*BinaryExpression)(nil)
var _ Node = (*UnaryExpression)(nil)
var _ Node = (*FunctionCall)(nil)
var _ Node = (*Identifier)(nil)
var _ Node = (*NumberLiteral)(nil)
var _ Node = (*StringLiteral)(nil)
var _ Node = (*BoolLiteral)(nil)
var _ Node = (*Assignment)(nil)
var _ Node = (*Program)(nil)
