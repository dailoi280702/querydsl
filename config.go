package querydsl

import (
	"github.com/dailoi280702/querydsl/compiler/sql"
	"log/slog"
)

// Config represents the configuration for the QueryDSL conversion.
type Config struct {
	FieldMap      map[string]string
	AllowedFields []string
	Placeholder   string // "?" or "$"
	Schema        Schema
	Logger        *slog.Logger
	CustomInfixes []sql.CustomInfix
}

// NewConfig creates a new empty configuration.
func NewConfig() Config {
	return Config{
		FieldMap: make(map[string]string),
	}
}

// WithSchema adds a validation schema to the config.
func (c Config) WithSchema(s Schema) Config {
	c.Schema = s
	return c
}

// WithMapping adds a single field mapping to the config.
func (c Config) WithMapping(from, to string) Config {
	c.FieldMap[from] = to
	return c
}

// WithAllowedFields sets the list of allowed fields.
func (c Config) WithAllowedFields(fields []string) Config {
	c.AllowedFields = fields
	return c
}

// WithPostgres switches placeholders to PostgreSQL style ($1, $2, ...).
func (c Config) WithPostgres() Config {
	c.Placeholder = "$"
	return c
}

// WithLogger sets the logger for the config.
func (c Config) WithLogger(l *slog.Logger) Config {
	c.Logger = l
	return c
}

// WithCustomInfix adds a custom compiler hook for infix expressions.
func (c Config) WithCustomInfix(fn sql.CustomInfix) Config {
	c.CustomInfixes = append(c.CustomInfixes, fn)
	return c
}

// Schema defines the validation rules for fields.
type Schema map[string]FieldRule

// FieldRule defines validation for a single field.
type FieldRule struct {
	Type     string // "string", "int", "float", "bool"
	Required bool
	Error    string // Custom error message
}
