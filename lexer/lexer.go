// Package lexer implements the lexical analysis for the QueryDSL.
package lexer

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// Lexer represents a lexical scanner.
type Lexer struct {
	input        string
	position     int  // current position in input (points to current char)
	readPosition int  // current reading position in input (after current char)
	ch           rune // current char under examination
	line         int
	column       int
}

// New creates a new Lexer instance for the given input.
func New(input string) *Lexer {
	l := &Lexer{input: input, line: 1, column: 0}
	l.readChar()
	return l
}

func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0
	} else {
		r, size := utf8.DecodeRuneInString(l.input[l.readPosition:])
		l.ch = r
		l.position = l.readPosition
		l.readPosition += size

		if r == '\n' {
			l.line++
			l.column = 0
		} else {
			l.column++
		}
		return
	}
	l.position = l.readPosition
	l.readPosition++
}

func (l *Lexer) peekChar() rune {
	if l.readPosition >= len(l.input) {
		return 0
	}
	r, _ := utf8.DecodeRuneInString(l.input[l.readPosition:])
	return r
}

// NextToken returns the next token from the input.
func (l *Lexer) NextToken() Token {
	var tok Token

	l.skipWhitespace()

	switch l.ch {
	case '=':
		if l.peekChar() == '=' {
			line, col := l.line, l.column
			l.readChar()
			tok = Token{Type: EQ, Literal: "==", Line: line, Column: col}
		} else {
			tok = l.newToken(EQ, "=")
		}
	case '!':
		if l.peekChar() == '=' {
			line, col := l.line, l.column
			l.readChar()
			tok = Token{Type: NEQ, Literal: "!=", Line: line, Column: col}
		} else {
			tok = l.newToken(NOT, "!")
		}
	case '-':
		tok = l.newToken(MINUS, "-")
	case '>':
		if l.peekChar() == '=' {
			line, col := l.line, l.column
			l.readChar()
			tok = Token{Type: GTE, Literal: ">=", Line: line, Column: col}
		} else {
			tok = l.newToken(GT, ">")
		}
	case '<':
		if l.peekChar() == '=' {
			line, col := l.line, l.column
			l.readChar()
			tok = Token{Type: LTE, Literal: "<=", Line: line, Column: col}
		} else {
			tok = l.newToken(LT, "<")
		}
	case '&':
		if l.peekChar() == '&' {
			line, col := l.line, l.column
			l.readChar()
			tok = Token{Type: AND, Literal: "&&", Line: line, Column: col}
		} else {
			tok = l.newToken(ILLEGAL, "&")
		}
	case '|':
		if l.peekChar() == '|' {
			line, col := l.line, l.column
			l.readChar()
			tok = Token{Type: OR, Literal: "||", Line: line, Column: col}
		} else {
			tok = l.newToken(ILLEGAL, "|")
		}
	case '%':
		tok = l.newToken(SIMILAR, "%")
	case '(':
		tok = l.newToken(LPAREN, "(")
	case ')':
		tok = l.newToken(RPAREN, ")")
	case '[':
		tok = l.newToken(LBRACKET, "[")
	case ']':
		tok = l.newToken(RBRACKET, "]")
	case ',':
		tok = l.newToken(COMMA, ",")
	case '"':
		line, col := l.line, l.column
		tok.Literal = l.readString()
		tok.Type = STRING
		tok.Line = line
		tok.Column = col
	case 0:
		tok.Literal = ""
		tok.Type = EOF
		tok.Line = l.line
		tok.Column = l.column
	default:
		if isLetter(l.ch) {
			line, col := l.line, l.column
			tok.Literal = l.readIdentifier()
			tok.Type = lookupIdent(tok.Literal)
			tok.Line = line
			tok.Column = col
			return tok
		}
		if isDigit(l.ch) {
			line, col := l.line, l.column
			tok.Literal = l.readNumber()
			tok.Type = NUMBER
			tok.Line = line
			tok.Column = col
			return tok
		}
		tok = l.newToken(ILLEGAL, string(l.ch))
	}

	l.readChar()
	return tok
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

func (l *Lexer) readIdentifier() string {
	start := l.position
	for isLetter(l.ch) || isDigit(l.ch) {
		l.readChar()
	}
	return l.input[start:l.position]
}

func (l *Lexer) readNumber() string {
	start := l.position
	for isDigit(l.ch) || l.ch == '.' {
		l.readChar()
	}
	return l.input[start:l.position]
}

func (l *Lexer) readString() string {
	start := l.position + 1
	for {
		l.readChar()
		if l.ch == '"' || l.ch == 0 {
			break
		}
	}
	return l.input[start:l.position]
}

func (l *Lexer) newToken(tokenType TokenType, literal string) Token {
	return Token{Type: tokenType, Literal: literal, Line: l.line, Column: l.column}
}

var keywords = map[string]TokenType{
	"true":  TRUE,
	"false": FALSE,
	"null":  NULL,
	"in":    IN,
	"like":  LIKE,
	"ilike": ILIKE,
}

func lookupIdent(ident string) TokenType {
	if tok, ok := keywords[strings.ToLower(ident)]; ok {
		return tok
	}
	return IDENT
}

func isLetter(ch rune) bool {
	return unicode.IsLetter(ch) || ch == '_'
}

func isDigit(ch rune) bool {
	return unicode.IsDigit(ch)
}
