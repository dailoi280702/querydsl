//nolint:revive
package parser

import (
	"testing"

	"github.com/dailoi280702/querydsl/lexer"
)

func TestParseExpression(t *testing.T) {
	input := `(a=="abc"||(c=1&&d>=100.100))`

	l := lexer.New(input)
	p := New(l)
	exp := p.ParseExpression(LOWEST)

	if len(p.Errors()) != 0 {
		for _, err := range p.Errors() {
			t.Errorf("parser error: %q", err)
		}
		t.FailNow()
	}

	expected := "((a == abc) || ((c = 1) && (d >= 100.100)))"
	if exp.String() != expected {
		t.Errorf("expected=%q, got=%q", expected, exp.String())
	}
}
