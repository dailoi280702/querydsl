// Package ast defines the abstract syntax tree for the QueryDSL.
//nolint:revive
package ast

import (
	"strings"
)

// LiteralType represents the data type of a literal.
type LiteralType string

const (
	// StringLiteral represents a string value.
	StringLiteral LiteralType = "STRING"
	// IntegerLiteral represents a whole number.
	IntegerLiteral LiteralType = "INTEGER"
	// FloatLiteral represents a decimal number.
	FloatLiteral LiteralType = "FLOAT"
	// BooleanLiteral represents a boolean value (true or false).
	BooleanLiteral LiteralType = "BOOLEAN"
	// NullLiteral represents a null value.
	NullLiteral LiteralType = "NULL"
)

// Node represents a node in the AST.
type Node interface {
	String() string
}

// Expression represents an expression node in the AST.
type Expression interface {
	Node
	expressionNode()
}

// Identifier represents an identifier in an expression.
type Identifier struct {
	Value string
}

func (i *Identifier) expressionNode() {}

// String returns the string representation of the identifier.
func (i *Identifier) String() string { return i.Value }

// Literal represents a literal value.
type Literal struct {
	Value string
	Type  LiteralType
}

func (l *Literal) expressionNode() {}

// String returns the string representation.
func (l *Literal) String() string { return l.Value }

// ArrayLiteral represents an array.
type ArrayLiteral struct {
	Elements []Expression
}

func (al *ArrayLiteral) expressionNode() {}

// String returns the string representation.
func (al *ArrayLiteral) String() string {
	var sb strings.Builder
	sb.WriteString("[")
	for i, e := range al.Elements {
		sb.WriteString(e.String())
		if i < len(al.Elements)-1 {
			sb.WriteString(", ")
		}
	}
	sb.WriteString("]")
	return sb.String()
}

// PrefixExpression represents an expression with a prefix operator (e.g., -5, !true).
type PrefixExpression struct {
	Operator string
	Right    Expression
}

func (pe *PrefixExpression) expressionNode() {}

// String returns the string representation.
func (pe *PrefixExpression) String() string {
	return "(" + pe.Operator + pe.Right.String() + ")"
}

// InfixExpression represents an infix operation.
type InfixExpression struct {
	Left     Expression
	Operator string
	Right    Expression
}

func (ie *InfixExpression) expressionNode() {}

// String returns the string representation.
func (ie *InfixExpression) String() string {
	return "(" + ie.Left.String() + " " + ie.Operator + " " + ie.Right.String() + ")"
}
