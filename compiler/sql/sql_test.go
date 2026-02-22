//nolint:revive
package sql

import (
	"testing"

	"querydsl/lexer"
	"querydsl/parser"
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
