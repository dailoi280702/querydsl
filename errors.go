package querydsl

import (
	"errors"
	"strings"
)

var (
	// ErrUnexpectedTokens is returned when there are tokens left after parsing an expression.
	ErrUnexpectedTokens = errors.New("unexpected tokens after expression")
	// ErrRequiredField is returned when a required field is missing from the query.
	ErrRequiredField = errors.New("required field missing")
	// ErrTypeMismatch is returned when a field value does not match the schema type.
	ErrTypeMismatch = errors.New("type mismatch")
	// ErrFieldNotAllowed is returned when a field is not in the allowed list or schema.
	ErrFieldNotAllowed = errors.New("field not allowed")
	// ErrFunctionNotAllowed is returned when a function is not in the allowed list.
	ErrFunctionNotAllowed = errors.New("function not allowed")
)

// ParserError represents errors that occur during parsing.
type ParserError struct {
	Errors []string
}

func (e *ParserError) Error() string {
	return "parser errors: " + strings.Join(e.Errors, "; ")
}

// ValidationError represents errors that occur during schema validation.
type ValidationError struct {
	Message string
	Code    error
}

func (e *ValidationError) Error() string {
	return "validation error: " + e.Message
}

// Unwrap returns the underlying error code.
func (e *ValidationError) Unwrap() error {
	return e.Code
}
