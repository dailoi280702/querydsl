package querydsl

import (
	"fmt"
	"time"

	"github.com/dailoi280702/querydsl/parser/ast"
)

// Validate checks the given AST node against the provided schema.
func Validate(node ast.Node, schema Schema) error {
	foundFields := make(map[string]bool)

	// Walk to check types and collect found fields
	err := walkValidate(node, schema, foundFields)
	if err != nil {
		return err
	}

	// Check required fields
	for field, rule := range schema {
		if rule.Required && !foundFields[field] {
			if rule.Error != "" {
				return &ValidationError{Message: rule.Error, Code: ErrRequiredField}
			}
			return &ValidationError{Message: fmt.Sprintf("field %s is required", field), Code: ErrRequiredField}
		}
	}

	return nil
}

func walkValidate(node ast.Node, schema Schema, found map[string]bool) error {
	switch n := node.(type) {
	case *ast.InfixExpression:
		if ident, ok := n.Left.(*ast.Identifier); ok {
			found[ident.Value] = true
			if rule, ok := schema[ident.Value]; ok {
				if err := validateType(n.Right, rule); err != nil {
					return err
				}
			}
		}
		if err := walkValidate(n.Left, schema, found); err != nil {
			return err
		}
		return walkValidate(n.Right, schema, found)
	case *ast.PrefixExpression:
		return walkValidate(n.Right, schema, found)
	case *ast.ArrayLiteral:
		for _, e := range n.Elements {
			if err := walkValidate(e, schema, found); err != nil {
				return err
			}
		}
	}
	return nil
}

func validateType(right ast.Expression, rule FieldRule) error {
	var actualType string
	var line, col int
	switch r := right.(type) {
	case *ast.Literal:
		line, col = r.Line, r.Column
		switch r.Type {
		case ast.StringLiteral:
			actualType = "string"
		case ast.IntegerLiteral:
			actualType = "int"
		case ast.FloatLiteral:
			actualType = "float"
		case ast.BooleanLiteral:
			actualType = "bool"
		case ast.NullLiteral:
			actualType = "null"
		}
	}

	if actualType == "string" {
		if rule.Type == "date" {
			if _, err := time.Parse("2006-01-02", right.(*ast.Literal).Value); err != nil {
				return &ValidationError{Message: fmt.Sprintf("[%d:%d] invalid date format, expected YYYY-MM-DD", line, col), Code: ErrTypeMismatch}
			}
			return nil
		}
		if rule.Type == "datetime" {
			if _, err := time.Parse(time.RFC3339, right.(*ast.Literal).Value); err != nil {
				return &ValidationError{Message: fmt.Sprintf("[%d:%d] invalid datetime format, expected RFC3339", line, col), Code: ErrTypeMismatch}
			}
			return nil
		}
	}

	if actualType != "" && actualType != rule.Type {
		// Special case: allow int to be used where float is expected
		if actualType == "int" && rule.Type == "float" {
			return nil
		}

		if rule.Error != "" {
			return &ValidationError{Message: fmt.Sprintf("[%d:%d] %s", line, col, rule.Error), Code: ErrTypeMismatch}
		}
		return &ValidationError{Message: fmt.Sprintf("[%d:%d] expected type %s, got %s", line, col, rule.Type, actualType), Code: ErrTypeMismatch}
	}
	return nil
}
