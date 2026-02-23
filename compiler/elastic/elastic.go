// Package elastic implements the Elasticsearch compiler for the QueryDSL.
package elastic

import (
	"fmt"
	"github.com/dailoi280702/querydsl/parser/ast"
	"strconv"
)

// Compiler compiles the AST into an Elasticsearch query map.
type Compiler struct {
	fieldMap map[string]string
}

// Config represents the configuration for the Elasticsearch compiler.
type Config struct {
	FieldMap map[string]string
}

// New creates a new Compiler instance.
func New(cfg Config) *Compiler {
	return &Compiler{
		fieldMap: cfg.FieldMap,
	}
}

// Compile transforms the AST into an Elasticsearch query map.
func (c *Compiler) Compile(node ast.Node) (map[string]any, error) {
	return c.walk(node)
}

func (c *Compiler) walk(node ast.Node) (map[string]any, error) {
	switch n := node.(type) {
	case *ast.InfixExpression:
		return c.compileInfix(n)
	case *ast.PrefixExpression:
		return c.compilePrefix(n)
	default:
		return nil, fmt.Errorf("unsupported node type: %T", node)
	}
}

func (c *Compiler) compileInfix(n *ast.InfixExpression) (map[string]any, error) {
	switch n.Operator {
	case "&&":
		left, err := c.walk(n.Left)
		if err != nil {
			return nil, err
		}
		right, err := c.walk(n.Right)
		if err != nil {
			return nil, err
		}
		return map[string]any{
			"bool": map[string]any{
				"must": []any{left, right},
			},
		}, nil
	case "||":
		left, err := c.walk(n.Left)
		if err != nil {
			return nil, err
		}
		right, err := c.walk(n.Right)
		if err != nil {
			return nil, err
		}
		return map[string]any{
			"bool": map[string]any{
				"should":               []any{left, right},
				"minimum_should_match": 1,
			},
		}, nil
	case "==":
		return c.compileTerm(n.Left, n.Right)
	case "!=":
		term, err := c.compileTerm(n.Left, n.Right)
		if err != nil {
			return nil, err
		}
		return map[string]any{
			"bool": map[string]any{
				"must_not": term,
			},
		}, nil
	case ">", ">=", "<", "<=":
		return c.compileRange(n)
	case "%":
		return c.compileFuzzy(n.Left, n.Right)
	default:
		return nil, fmt.Errorf("unsupported operator: %s", n.Operator)
	}
}

func (c *Compiler) compileFuzzy(left, right ast.Expression) (map[string]any, error) {
	field, err := c.getFieldName(left)
	if err != nil {
		return nil, err
	}
	val, err := c.getValue(right)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"match": map[string]any{
			field: map[string]any{
				"query":     val,
				"fuzziness": "AUTO",
			},
		},
	}, nil
}

func (c *Compiler) compilePrefix(n *ast.PrefixExpression) (map[string]any, error) {
	if n.Operator == "!" {
		right, err := c.walk(n.Right)
		if err != nil {
			return nil, err
		}
		return map[string]any{
			"bool": map[string]any{
				"must_not": right,
			},
		}, nil
	}
	return nil, fmt.Errorf("unsupported prefix operator: %s", n.Operator)
}

func (c *Compiler) compileTerm(left, right ast.Expression) (map[string]any, error) {
	field, err := c.getFieldName(left)
	if err != nil {
		return nil, err
	}
	val, err := c.getValue(right)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"term": map[string]any{
			field: val,
		},
	}, nil
}

func (c *Compiler) compileRange(n *ast.InfixExpression) (map[string]any, error) {
	field, err := c.getFieldName(n.Left)
	if err != nil {
		return nil, err
	}
	val, err := c.getValue(n.Right)
	if err != nil {
		return nil, err
	}

	opMap := map[string]string{
		">":  "gt",
		">=": "gte",
		"<":  "lt",
		"<=": "lte",
	}

	return map[string]any{
		"range": map[string]any{
			field: map[string]any{
				opMap[n.Operator]: val,
			},
		},
	}, nil
}

func (c *Compiler) getFieldName(expr ast.Expression) (string, error) {
	if ident, ok := expr.(*ast.Identifier); ok {
		if mapped, ok := c.fieldMap[ident.Value]; ok {
			return mapped, nil
		}
		return ident.Value, nil
	}
	return "", fmt.Errorf("expected identifier, got %T", expr)
}

func (c *Compiler) getValue(expr ast.Expression) (any, error) {
	if lit, ok := expr.(*ast.Literal); ok {
		switch lit.Type {
		case ast.IntegerLiteral:
			return strconv.ParseInt(lit.Value, 10, 64)
		case ast.FloatLiteral:
			return strconv.ParseFloat(lit.Value, 64)
		case ast.BooleanLiteral:
			return strconv.ParseBool(lit.Value)
		case ast.StringLiteral:
			return lit.Value, nil
		}
	}
	return nil, fmt.Errorf("expected literal, got %T", expr)
}
