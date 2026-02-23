//nolint:revive
package sql

import (
	"testing"

	"github.com/dailoi280702/querydsl/lexer"
	"github.com/dailoi280702/querydsl/parser"
	"github.com/dailoi280702/querydsl/parser/ast"
)

func TestCompile(t *testing.T) {
	input := `(a=="abc"||(c=1&&d>=100.100))`

	l := lexer.New(input)
	p := parser.New(l)
	exp := p.ParseExpression(parser.LOWEST)

	if len(p.Errors()) != 0 {
		for _, err := range p.Errors() {
			t.Errorf("parser error: %q", err)
		}
		t.FailNow()
	}

	c := New()
	sql, args, err := c.Compile(exp)
	if err != nil {
		t.Fatalf("compiler error: %v", err)
	}

	expectedSQL := "((a = ?) OR ((c = ?) AND (d >= ?)))"
	if sql != expectedSQL {
		t.Errorf("expected SQL=%q, got=%q", expectedSQL, sql)
	}

	if len(args) != 3 {
		t.Fatalf("expected 3 args, got %d", len(args))
	}

	if args[0] != "abc" {
		t.Errorf("expected args[0]=%q, got %v", "abc", args[0])
	}

	if args[1] != int64(1) {
		t.Errorf("expected args[1]=%v, got %v", 1, args[1])
	}

	if args[2] != 100.100 {
		t.Errorf("expected args[2]=%v, got %v", 100.100, args[2])
	}
}

func TestCompileCall(t *testing.T) {
	c := New(Config{
		AllowedFunctions: []string{"lower", "trim"},
	})

	t.Run("Allowed function", func(t *testing.T) {
		call := &ast.CallExpression{
			Function: "lower",
			Arguments: []ast.Expression{
				&ast.Identifier{Value: "name"},
			},
		}
		sql, _, err := c.Compile(call)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if sql != "lower(name)" {
			t.Errorf("expected lower(name), got %q", sql)
		}
	})

	t.Run("Not allowed function", func(t *testing.T) {
		call := &ast.CallExpression{
			Function: "danger",
			Arguments: []ast.Expression{
				&ast.Identifier{Value: "name"},
			},
		}
		_, _, err := c.Compile(call)
		if err == nil {
			t.Fatal("expected error for unauthorized function, got nil")
		}
	})

	t.Run("Nested allowed functions", func(t *testing.T) {
		call := &ast.CallExpression{
			Function: "lower",
			Arguments: []ast.Expression{
				&ast.CallExpression{
					Function: "trim",
					Arguments: []ast.Expression{
						&ast.Identifier{Value: "name"},
					},
				},
			},
		}
		sql, _, err := c.Compile(call)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if sql != "lower(trim(name))" {
			t.Errorf("expected lower(trim(name)), got %q", sql)
		}
	})
}
