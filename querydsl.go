// Package querydsl provides a high-level API to convert Query DSL to various database formats.
package querydsl

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/dailoi280702/querydsl/compiler/sql"
	"github.com/dailoi280702/querydsl/lexer"
	"github.com/dailoi280702/querydsl/parser"
	"github.com/dailoi280702/querydsl/parser/ast"
)

// QueryTranspiler defines the abstract interface for converting DSL
// to a database-specific query format.
type QueryTranspiler interface {
	Parse(input string) (ast.Node, error)
	Validate(node ast.Node, schema Schema) error
	Transpile(node ast.Node, cfg Config) (string, []any, error)
}

// SQLBackend implements the QueryTranspiler interface for SQL databases.
type SQLBackend struct{}

// NewSQLBackend creates a new SQL transpiler backend.
func NewSQLBackend() *SQLBackend {
	return &SQLBackend{}
}

// Parse converts a string into an AST Node.
func (b *SQLBackend) Parse(input string) (ast.Node, error) {
	return Parse(input)
}

// Validate checks an AST Node against a schema.
func (b *SQLBackend) Validate(node ast.Node, schema Schema) error {
	return Validate(node, schema)
}

// Transpile converts an AST Node into a SQL WHERE clause and arguments.
func (b *SQLBackend) Transpile(node ast.Node, cfg Config) (string, []any, error) {
	fieldTypes := make(map[string]string)
	for k, v := range cfg.Schema {
		fieldTypes[k] = v.Type
	}

	sqlCfg := sql.Config{
		FieldMap:      cfg.FieldMap,
		AllowedFields: cfg.AllowedFields,
		FieldTypes:    fieldTypes,
		Placeholder:   cfg.Placeholder,
		CustomInfixes: cfg.CustomInfixes,
	}
	compiler := sql.New(sqlCfg)
	sqlStr, args, err := compiler.Compile(node)
	if err != nil {
		if strings.Contains(err.Error(), "field not allowed") {
			return "", nil, fmt.Errorf("%w: %s", ErrFieldNotAllowed, err.Error())
		}
		return "", nil, err
	}
	return sqlStr, args, nil
}

// Parse converts a DSL string into an AST Node.
func Parse(input string) (ast.Node, error) {
	l := lexer.New(input)
	p := parser.New(l)

	expr := p.ParseExpression(parser.LOWEST)
	if len(p.Errors()) > 0 {
		return nil, &ParserError{Errors: p.Errors()}
	}

	if !p.IsEOF() {
		return nil, fmt.Errorf("%w", ErrUnexpectedTokens)
	}

	return expr, nil
}

// ToSQL is a helper that performs the full pipeline from string to SQL.
func ToSQL(input string, cfg ...Config) (string, []any, error) {
	var activeCfg Config
	if len(cfg) > 0 {
		activeCfg = cfg[0]
	} else {
		activeCfg = NewConfig()
	}

	logger := activeCfg.Logger
	if logger == nil {
		logger = slog.Default()
	}

	logger.Debug("parsing QueryDSL input", "input", input)
	node, err := Parse(input)
	if err != nil {
		logger.Error("failed to parse QueryDSL", "input", input, "err", err)
		return "", nil, err
	}

	if activeCfg.Schema != nil {
		logger.Debug("validating QueryDSL against schema")
		if err := Validate(node, activeCfg.Schema); err != nil {
			logger.Warn("validation failed", "err", err)
			return "", nil, err
		}
	}

	backend := NewSQLBackend()
	sqlStr, args, err := backend.Transpile(node, activeCfg)
	if err != nil {
		logger.Error("transpilation failed", "err", err)
		return "", nil, err
	}

	logger.Debug("successfully transpiled QueryDSL to SQL", "sql", sqlStr, "args_count", len(args))
	return sqlStr, args, nil
}
