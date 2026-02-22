// Package lexer implements the lexical analysis for the QueryDSL.
package lexer

// TokenType represents the type of a token.
type TokenType string

// Token constants for lexer symbols and keywords.
const (
	ILLEGAL TokenType = "ILLEGAL"
	EOF     TokenType = "EOF"

	IDENT  TokenType = "IDENT"
	STRING TokenType = "STRING"
	NUMBER TokenType = "NUMBER"
	TRUE   TokenType = "TRUE"
	FALSE  TokenType = "FALSE"
	NULL   TokenType = "NULL"

	EQ    TokenType = "=="
	NEQ   TokenType = "!="
	GT    TokenType = ">"
	GTE   TokenType = ">="
	LT    TokenType = "<"
	LTE   TokenType = "<="
	AND   TokenType = "&&"
	OR    TokenType = "||"
	IN    TokenType = "IN"
	LIKE  TokenType = "LIKE"
	ILIKE TokenType = "ILIKE"
	// SIMILAR represents the similarity operator (%) in Postgres.
	SIMILAR TokenType = "%"

	// Prefix
	MINUS TokenType = "-"
	NOT   TokenType = "!"

	LPAREN   TokenType = "("
	RPAREN   TokenType = ")"
	LBRACKET TokenType = "["
	RBRACKET TokenType = "]"
	COMMA    TokenType = ","
)

// Token represents a single lexical token.
type Token struct {
	Type    TokenType
	Literal string
}
