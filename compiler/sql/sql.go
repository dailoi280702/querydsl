// Package sql implements the SQL compiler for the QueryDSL.
//
//nolint:revive
package sql

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/dailoi280702/querydsl/parser/ast"
)

// CustomInfix is a function that can override the default compilation of an infix expression.
type CustomInfix func(n *ast.InfixExpression, walk func(ast.Node, string) (string, error)) (sql string, handled bool, err error)

// Compiler represents the SQL compiler.
type Compiler struct {
	Args          []any
	fieldMap      map[string]string
	allowedFields map[string]bool
	fieldTypes    map[string]string // field -> type
	placeholder   string            // "?" or "$"
	argCount      int
	customInfixes []CustomInfix
}

// Config represents the configuration for the SQL compiler.
type Config struct {
	FieldMap      map[string]string
	AllowedFields []string
	FieldTypes    map[string]string
	Placeholder   string // "?" or "$"
	CustomInfixes []CustomInfix
}

// New creates a new Compiler instance.
func New(cfg ...Config) *Compiler {
	c := &Compiler{
		Args:          []any{},
		fieldMap:      make(map[string]string),
		allowedFields: make(map[string]bool),
		fieldTypes:    make(map[string]string),
		placeholder:   "?",
	}

	if len(cfg) > 0 {
		c.fieldMap = cfg[0].FieldMap
		for _, f := range cfg[0].AllowedFields {
			c.allowedFields[f] = true
		}
		c.fieldTypes = cfg[0].FieldTypes
		if cfg[0].Placeholder != "" {
			c.placeholder = cfg[0].Placeholder
		}
		c.customInfixes = cfg[0].CustomInfixes
	}

	return c
}

// Compile compiles the given AST node into a SQL WHERE clause and arguments.
func (c *Compiler) Compile(node ast.Node) (string, []any, error) {
	c.Args = []any{} // reset args for fresh compile
	c.argCount = 0   // reset counter
	sql, err := c.walk(node, "")
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

func (c *Compiler) walk(node ast.Node, currentFieldType string) (string, error) {
	switch n := node.(type) {
	case *ast.InfixExpression:
		return c.compileInfix(n)
	case *ast.PrefixExpression:
		// Optimization: Handle negative numbers as a single argument
		if n.Operator == "-" {
			if lit, ok := n.Right.(*ast.Literal); ok {
				if lit.Type == ast.IntegerLiteral || lit.Type == ast.FloatLiteral {
					val, err := c.parseLiteralValue(lit)
					if err != nil {
						return "", err
					}
					// Negate the value based on type
					switch v := val.(type) {
					case int64:
						c.Args = append(c.Args, -v)
					case float64:
						c.Args = append(c.Args, -v)
					default:
						return "", fmt.Errorf("unexpected number type: %T", v)
					}
					return c.nextPlaceholder(), nil
				}
			}
		}
		return c.compilePrefix(n, currentFieldType)
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

		// Smart Conversion: If field is a date/datetime, convert string to time.Time
		if n.Type == ast.StringLiteral {
			if s, ok := val.(string); ok {
				switch currentFieldType {
				case "date":
					if t, err := time.Parse("2006-01-02", s); err == nil {
						val = t
					}
				case "datetime":
					if t, err := time.Parse(time.RFC3339, s); err == nil {
						val = t
					}
				}
			}
		}

		c.Args = append(c.Args, val)
		return c.nextPlaceholder(), nil
	case *ast.ArrayLiteral:
		return c.compileArray(n, currentFieldType)
	default:
		return "", fmt.Errorf("unknown node type: %T", node)
	}
}

func (c *Compiler) compilePrefix(n *ast.PrefixExpression, currentFieldType string) (string, error) {
	operator := n.Operator
	if operator == "!" {
		operator = "NOT "
	}

	right, err := c.walk(n.Right, currentFieldType)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("(%s%s)", operator, right), nil
}

func (c *Compiler) compileArray(n *ast.ArrayLiteral, currentFieldType string) (string, error) {
	var sb strings.Builder
	sb.WriteString("(")
	for i, e := range n.Elements {
		ph, err := c.walk(e, currentFieldType)
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
	for _, hook := range c.customInfixes {
		sqlStr, handled, err := hook(n, c.walk)
		if err != nil {
			return "", err
		}
		if handled {
			return sqlStr, nil
		}
	}

	leftType := ""
	if ident, ok := n.Left.(*ast.Identifier); ok {
		leftType = c.fieldTypes[ident.Value]
	}

	left, err := c.walk(n.Left, "")
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

	right, err := c.walk(n.Right, leftType)
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
