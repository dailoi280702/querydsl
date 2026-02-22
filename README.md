# QueryDSL for Go

A robust, high-performance library for transpiling a human-readable Query DSL into parameterized SQL `WHERE` clauses. It is designed for use in HTTP APIs where users need to filter data dynamically and securely.

## Features

- **Lexer/Parser:** Hand-written Pratt Parser for accurate operator precedence.
- **SQL Injection Protection:** Generates parameterized SQL (`?` or `$1`) with a separate arguments slice.
- **JSON Type Support:** Handles `true`, `false`, `null`, and array literals.
- **Multi-Dialect:** Supports standard SQL (`?`) and PostgreSQL (`$1`, `$2`, ...) placeholders.
- **Security:** Built-in field whitelisting, schema validation, and mapping.
- **Builder Pattern:** Easily layer configurations across service and repository layers.
- **QOL:** Case-insensitive keywords (`IN`, `NULL`, `TRUE`, `FALSE`) and UTF-8 support (Vietnamese).

## Installation

```bash
go get querydsl
```

## Quick Start

```go
import "querydsl"

// Simple one-liner
where, args, err := querydsl.ToSQL(`status == "active" && age >= 18`)
// Result: (status = ?) AND (age >= ?)
// Args: ["active", 18]
```

## Advanced Configuration (Builder Pattern)

The builder pattern allows you to define validation in your service layer and database-specific details in your repository.

```go
// 1. Service layer adds validation rules
cfg := querydsl.NewConfig().WithSchema(querydsl.Schema{
    "name": querydsl.FieldRule{Type: "string", Required: true},
    "age":  querydsl.FieldRule{Type: "int"},
})

// 2. Repository layer adds mapping and Postgres dialect
cfg = cfg.WithMapping("name", "full_name").WithPostgres()

where, args, err := querydsl.ToSQL(`name == "John" && age >= 18`, cfg)
// Result: ((full_name = $1) AND (age >= $2))
// Args: ["John", 18]
```

## Error Handling

The library provides exported sentinel errors for precise checking:

```go
where, args, err := querydsl.ToSQL(input, cfg)
if err != nil {
    if errors.Is(err, querydsl.ErrFieldNotAllowed) {
        // Handle unauthorized field access
    } else if errors.Is(err, querydsl.ErrTypeMismatch) {
        // Handle type errors (e.g., string passed to an int field)
    }
}
```

## Supported Syntax

| Operator | DSL | SQL Equivalent |
| :--- | :--- | :--- |
| Equality | `==`, `=` | `=` |
| Inequality | `!=` | `!=` |
| Comparison | `>`, `<`, `>=`, `<=` | `>`, `<`, `>=`, `<=` |
| Logical | `&&`, `||` | `AND`, `OR` |
| Inclusion | `in` | `IN (...)` |
| Similarity | `%` | `%` (Postgres Trigram) |
| Pattern | `like`, `ilike` | `LIKE`, `ILIKE` |
| Null Check | `== null`, `!= null` | `IS NULL`, `IS NOT NULL` |

## Development

This project uses `mise` for tool management.

```bash
export GITHUB_TOKEN=$(gh auth token)
mise install
go test ./...
```

## Security

**Never** concatenate the returned `where` clause directly into a string. Always use the database driver's parameter passing:

```go
where, args, err := querydsl.ToSQL(input, cfg)
db.Query("SELECT * FROM users WHERE " + where, args...)
```
