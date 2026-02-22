// Package sql implements the SQL compiler for the QueryDSL.
package sql

import (
	"fmt"
	"querydsl/parser/ast"
	"strconv"
	"strings"
)

// Compiler represents the SQL compiler.
type Compiler struct {
	Args          []any
	fieldMap      map[string]string
	allowedFields map[string]bool
	placeholder   string // "?" or "$"
	argCount      int
}

// Config represents the configuration for the SQL compiler.
type Config struct {
	FieldMap      map[string]string
	AllowedFields []string
	Placeholder   string // "?" or "$"
}

// New creates a new Compiler instance.
func New(cfg ...Config) *Compiler {
	c := &Compiler{
		Args:          []any{},
		fieldMap:      make(map[string]string),
		allowedFields: make(map[string]bool),
		placeholder:   "?",
	}

	if len(cfg) > 0 {
		c.fieldMap = cfg[0].FieldMap
		for _, f := range cfg[0].AllowedFields {
			c.allowedFields[f] = true
		}
		if cfg[0].Placeholder != "" {
			c.placeholder = cfg[0].Placeholder
		}
	}

	return c
}

// Compile compiles the given AST node into a SQL WHERE clause and arguments.
func (c *Compiler) Compile(node ast.Node) (string, []any, error) {
	c.Args = []any{} // reset args for fresh compile
	c.argCount = 0   // reset counter
	sql, err := c.walk(node)
	if err != nil {
		return "", nil, err
	}
	return sql, c.Args, nil
}

func (c *Compiler) nextPlaceholder() string {
	if c.placeholder == "?" {
		return "?"
	}
	c.argCount++
	return fmt.Sprintf("$%d", c.argCount)
}

func (c *Compiler) walk(node ast.Node) (string, error) {
	switch n := node.(type) {
	case *ast.InfixExpression:
		return c.compileInfix(n)
	case *ast.Identifier:
		fieldName := n.Value
		if len(c.allowedFields) > 0 {
			if !c.allowedFields[fieldName] {
				return "", fmt.Errorf("field not allowed: %s", fieldName)
			}
		}
		if mapped, ok := c.fieldMap[fieldName]; ok {
			return mapped, nil
		}
		return fieldName, nil
	case *ast.Literal:
		val, err := c.parseLiteralValue(n)
		if err != nil {
			return "", err
		}
		c.Args = append(c.Args, val)
		return c.nextPlaceholder(), nil
	case *ast.ArrayLiteral:
		return c.compileArray(n)
	default:
		return "", fmt.Errorf("unknown node type: %T", node)
	}
}

func (c *Compiler) compileArray(n *ast.ArrayLiteral) (string, error) {
	var sb strings.Builder
	sb.WriteString("(")
	for i, e := range n.Elements {
		ph, err := c.walk(e)
		if err != nil {
			return "", err
		}
		sb.WriteString(ph)
		if i < len(n.Elements)-1 {
			sb.WriteString(", ")
		}
	}
	sb.WriteString(")")
	return sb.String(), nil
}

func (c *Compiler) compileInfix(n *ast.InfixExpression) (string, error) {
	left, err := c.walk(n.Left)
	if err != nil {
		return "", err
	}

	// Special case for NULL
	if rightLiteral, ok := n.Right.(*ast.Literal); ok && rightLiteral.Type == ast.NullLiteral {
		if n.Operator == "==" || n.Operator == "=" {
			return fmt.Sprintf("(%s IS NULL)", left), nil
		}
		if n.Operator == "!=" {
			return fmt.Sprintf("(%s IS NOT NULL)", left), nil
		}
	}

	right, err := c.walk(n.Right)
	if err != nil {
		return "", err
	}

	operator := n.Operator
	switch operator {
	case "==":
		operator = "="
	case "&&":
		operator = "AND"
	case "||":
		operator = "OR"
	case "in":
		operator = "IN"
	case "like":
		operator = "LIKE"
	case "ilike":
		operator = "ILIKE"
	case "%":
		operator = "%"
	}

	return fmt.Sprintf("(%s %s %s)", left, operator, right), nil
}

func (c *Compiler) parseLiteralValue(l *ast.Literal) (any, error) {
	switch l.Type {
	case ast.BooleanLiteral:
		return strings.ToLower(l.Value) == "true", nil
	case ast.NullLiteral:
		return nil, nil
	case ast.IntegerLiteral:
		return strconv.ParseInt(l.Value, 10, 64)
	case ast.FloatLiteral:
		return strconv.ParseFloat(l.Value, 64)
	case ast.StringLiteral:
		return l.Value, nil
	default:
		return nil, fmt.Errorf("unknown literal type: %s", l.Type)
	}
}
