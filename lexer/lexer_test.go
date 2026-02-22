package lexer

import (
	"testing"
)

func TestNextToken(t *testing.T) {
	input := `(a=="abc"||(c=1&&d>=100.100))`

	tests := []struct {
		expectedType    TokenType
		expectedLiteral string
	}{
		{LPAREN, "("},
		{IDENT, "a"},
		{EQ, "=="},
		{STRING, "abc"},
		{OR, "||"},
		{LPAREN, "("},
		{IDENT, "c"},
		{EQ, "="},
		{NUMBER, "1"},
		{AND, "&&"},
		{IDENT, "d"},
		{GTE, ">="},
		{NUMBER, "100.100"},
		{RPAREN, ")"},
		{RPAREN, ")"},
		{EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}
