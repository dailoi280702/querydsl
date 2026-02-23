// Package parser implements the parser for the QueryDSL.
//
//nolint:revive
package parser

import (
	"fmt"
	"strings"

	"github.com/dailoi280702/querydsl/lexer"
	"github.com/dailoi280702/querydsl/parser/ast"
)

const (
	_ int = iota
	// LOWEST represents the lowest operator precedence.
	LOWEST
	// OR represents the precedence of the OR operator.
	OR
	// AND represents the precedence of the AND operator.
	AND
	// EQUALS represents the precedence of comparison operators.
		EQUALS
		// PREFIX represents the precedence of prefix operators.
		PREFIX
		// CALL represents the precedence of function calls.
		CALL
	)
	
	var precedences = map[lexer.TokenType]int{
		lexer.EQ:      EQUALS,
		lexer.NEQ:     EQUALS,
		lexer.LT:      EQUALS,
		lexer.LTE:     EQUALS,
		lexer.GT:      EQUALS,
		lexer.GTE:     EQUALS,
		lexer.IN:      EQUALS,
		lexer.LIKE:    EQUALS,
		lexer.ILIKE:   EQUALS,
		lexer.SIMILAR: EQUALS,
		lexer.AND:     AND,
		lexer.OR:      OR,
		lexer.LPAREN:  CALL,
	}
	

type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(ast.Expression) ast.Expression
)

// Parser represents the QueryDSL parser.
type Parser struct {
	l      *lexer.Lexer
	errors []string

	curToken  lexer.Token
	peekToken lexer.Token

	prefixParseFns map[lexer.TokenType]prefixParseFn
	infixParseFns  map[lexer.TokenType]infixParseFn
}

// New creates a new Parser instance.
func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:      l,
		errors: []string{},
	}

	p.prefixParseFns = make(map[lexer.TokenType]prefixParseFn)
	p.registerPrefix(lexer.IDENT, p.parseIdentifier)
	p.registerPrefix(lexer.NUMBER, p.parseLiteral)
	p.registerPrefix(lexer.STRING, p.parseLiteral)
	p.registerPrefix(lexer.TRUE, p.parseLiteral)
	p.registerPrefix(lexer.FALSE, p.parseLiteral)
	p.registerPrefix(lexer.NULL, p.parseLiteral)
	p.registerPrefix(lexer.LPAREN, p.parseGroupedExpression)
	p.registerPrefix(lexer.LBRACKET, p.parseArrayLiteral)
	p.registerPrefix(lexer.MINUS, p.parsePrefixExpression)
	p.registerPrefix(lexer.NOT, p.parsePrefixExpression)

	p.infixParseFns = make(map[lexer.TokenType]infixParseFn)
	p.registerInfix(lexer.EQ, p.parseInfixExpression)
	p.registerInfix(lexer.NEQ, p.parseInfixExpression)
	p.registerInfix(lexer.LT, p.parseInfixExpression)
	p.registerInfix(lexer.LTE, p.parseInfixExpression)
	p.registerInfix(lexer.GT, p.parseInfixExpression)
	p.registerInfix(lexer.GTE, p.parseInfixExpression)
	p.registerInfix(lexer.IN, p.parseInfixExpression)
	p.registerInfix(lexer.LIKE, p.parseInfixExpression)
	p.registerInfix(lexer.ILIKE, p.parseInfixExpression)
	p.registerInfix(lexer.SIMILAR, p.parseInfixExpression)
	p.registerInfix(lexer.AND, p.parseInfixExpression)
	p.registerInfix(lexer.OR, p.parseInfixExpression)
	p.registerInfix(lexer.LPAREN, p.parseCallExpression)

	// Read two tokens, so curToken and peekToken are both set
	p.nextToken()
	p.nextToken()

	return p
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()

	if p.curToken.Type == lexer.ILLEGAL {
		p.errors = append(p.errors, fmt.Sprintf("[%d:%d] illegal token: %s", p.curToken.Line, p.curToken.Column, p.curToken.Literal))
	}
}

func (p *Parser) peekTokenIs(t lexer.TokenType) bool {
	return p.peekToken.Type == t
}

func (p *Parser) expectPeek(t lexer.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	}
	p.peekError(t)
	return false
}

func (p *Parser) peekError(t lexer.TokenType) {
	msg := fmt.Sprintf("[%d:%d] expected next token to be %s, got %s instead",
		p.peekToken.Line, p.peekToken.Column, t, p.peekToken.Type)
	p.errors = append(p.errors, msg)
}

// Errors returns the list of parser errors.
func (p *Parser) Errors() []string {
	return p.errors
}

// IsEOF checks if the next token is the end of input.
func (p *Parser) IsEOF() bool {
	return p.peekToken.Type == lexer.EOF
}

func (p *Parser) registerPrefix(tokenType lexer.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

func (p *Parser) registerInfix(tokenType lexer.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}

// ParseExpression parses an expression based on the given precedence.
func (p *Parser) ParseExpression(precedence int) ast.Expression {
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}
	leftExp := prefix()

	for !p.peekTokenIs(lexer.EOF) && precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}

		p.nextToken()

		leftExp = infix(leftExp)
	}

	return leftExp
}

func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) curPrecedence() int {
	if p, ok := precedences[p.curToken.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{
		Value:  p.curToken.Literal,
		Line:   p.curToken.Line,
		Column: p.curToken.Column,
	}
}

func (p *Parser) parsePrefixExpression() ast.Expression {
	expression := &ast.PrefixExpression{
		Operator: p.curToken.Literal,
	}

	p.nextToken()

	expression.Right = p.ParseExpression(PREFIX)

	return expression
}

func (p *Parser) parseLiteral() ast.Expression {
	var litType ast.LiteralType
	switch p.curToken.Type {
	case lexer.STRING:
		litType = ast.StringLiteral
	case lexer.NUMBER:
		if strings.Contains(p.curToken.Literal, ".") {
			litType = ast.FloatLiteral
		} else {
			litType = ast.IntegerLiteral
		}
	case lexer.TRUE, lexer.FALSE:
		litType = ast.BooleanLiteral
	case lexer.NULL:
		litType = ast.NullLiteral
	}
	return &ast.Literal{
		Value:  p.curToken.Literal,
		Type:   litType,
		Line:   p.curToken.Line,
		Column: p.curToken.Column,
	}
}

func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	expression := &ast.InfixExpression{
		Left:     left,
		Operator: p.curToken.Literal,
	}

	precedence := p.curPrecedence()
	p.nextToken()
	expression.Right = p.ParseExpression(precedence)

	return expression
}

func (p *Parser) parseCallExpression(function ast.Expression) ast.Expression {
	exp := &ast.CallExpression{Function: function.String()}
	exp.Arguments = p.parseExpressionList(lexer.RPAREN)
	return exp
}

func (p *Parser) parseGroupedExpression() ast.Expression {
	p.nextToken()

	exp := p.ParseExpression(LOWEST)

	if !p.expectPeek(lexer.RPAREN) {
		return nil
	}

	return exp
}

func (p *Parser) parseArrayLiteral() ast.Expression {
	array := &ast.ArrayLiteral{}
	array.Elements = p.parseExpressionList(lexer.RBRACKET)
	return array
}

func (p *Parser) parseExpressionList(end lexer.TokenType) []ast.Expression {
	list := []ast.Expression{}

	if p.peekTokenIs(end) {
		p.nextToken()
		return list
	}

	p.nextToken()
	list = append(list, p.ParseExpression(LOWEST))

	for p.peekTokenIs(lexer.COMMA) {
		p.nextToken()
		p.nextToken()
		list = append(list, p.ParseExpression(LOWEST))
	}

	if !p.expectPeek(end) {
		return nil
	}

	return list
}

func (p *Parser) noPrefixParseFnError(t lexer.TokenType) {
	msg := fmt.Sprintf("[%d:%d] no prefix parse function for %s found",
		p.curToken.Line, p.curToken.Column, t)
	p.errors = append(p.errors, msg)
}
