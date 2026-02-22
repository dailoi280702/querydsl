package querydsl

// Config represents the configuration for the QueryDSL conversion.
type Config struct {
	FieldMap      map[string]string
	AllowedFields []string
	Placeholder   string // "?" or "$"
	Schema        Schema
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

// Schema defines the validation rules for fields.
type Schema map[string]FieldRule

// FieldRule defines validation for a single field.
type FieldRule struct {
	Type     string // "string", "int", "float", "bool"
	Required bool
	Error    string // Custom error message
}
