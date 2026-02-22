package querydsl

import (
	"errors"
	"querydsl/parser/ast"
	"strings"
	"testing"
)

func TestToSQL(t *testing.T) {
	tests := []struct {
		input       string
		expectedSQL string
		expectedLen int
		wantErr     bool
	}{
		{
			input:       `(a=="abc"||(c=1&&d>=100.100))`,
			expectedSQL: "((a = ?) OR ((c = ?) AND (d >= ?)))",
			expectedLen: 3,
			wantErr:     false,
		},
		{
			input:       `status != "deleted" && price < 1000`,
			expectedSQL: "((status != ?) AND (price < ?))",
			expectedLen: 2,
			wantErr:     false,
		},
		{
			input:       `age > 18 || name == "admin"`,
			expectedSQL: "((age > ?) OR (name = ?))",
			expectedLen: 2,
			wantErr:     false,
		},
		{
			input:       `x <= 5.5`,
			expectedSQL: "(x <= ?)",
			expectedLen: 1,
			wantErr:     false,
		},
		{
			input:       `is_active == true`,
			expectedSQL: "(is_active = ?)",
			expectedLen: 1,
			wantErr:     false,
		},
		{
			input:       `deleted_at == null`,
			expectedSQL: "(deleted_at IS NULL)",
			expectedLen: 0,
			wantErr:     false,
		},
		{
			input:       `status in ["active", "pending"]`,
			expectedSQL: "(status IN (?, ?))",
			expectedLen: 2,
			wantErr:     false,
		},
		{
			input:       `name like "john%"`,
			expectedSQL: "(name LIKE ?)",
			expectedLen: 1,
			wantErr:     false,
		},
		{
			input:       `email ilike "%@gmail.com"`,
			expectedSQL: "(email ILIKE ?)",
			expectedLen: 1,
			wantErr:     false,
		},
		{
			input:       `name % "john"`,
			expectedSQL: "(name % ?)",
			expectedLen: 1,
			wantErr:     false,
		},
		{
			input:       `tên == "Nguyễn Văn A"`,
			expectedSQL: "(tên = ?)",
			expectedLen: 1,
			wantErr:     false,
		},
		{
			input:       `mô_tả % "tiếng việt"`,
			expectedSQL: "(mô_tả % ?)",
			expectedLen: 1,
			wantErr:     false,
		},
		// Error cases
		{
			input:   `a == "abc" && (b == 1`,
			wantErr: true,
		},
		{
			input:   `a == "abc" & b == 1`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		sql, args, err := ToSQL(tt.input)
		if (err != nil) != tt.wantErr {
			t.Errorf("ToSQL(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			continue
		}

		if !tt.wantErr {
			if sql != tt.expectedSQL {
				t.Errorf("ToSQL(%q) SQL = %q, want %q", tt.input, sql, tt.expectedSQL)
			}
			if len(args) != tt.expectedLen {
				t.Errorf("ToSQL(%q) args length = %d, want %d", tt.input, len(args), tt.expectedLen)
			}
		}
	}
}

func TestToSQLWithOptions(t *testing.T) {
	cfg := Config{
		FieldMap: map[string]string{
			"user": "user_id",
			"name": "full_name",
		},
		AllowedFields: []string{"user", "name", "age"},
	}

	tests := []struct {
		input       string
		expectedSQL string
		wantErr     bool
	}{
		{
			input:       `user == 1`,
			expectedSQL: "(user_id = ?)",
			wantErr:     false,
		},
		{
			input:       `name == "abc"`,
			expectedSQL: "(full_name = ?)",
			wantErr:     false,
		},
		{
			input:       `age > 18`,
			expectedSQL: "(age > ?)",
			wantErr:     false,
		},
		{
			input:   `password == "123"`, // Not allowed
			wantErr: true,
		},
	}

	for _, tt := range tests {
		sql, _, err := ToSQL(tt.input, cfg)
		if (err != nil) != tt.wantErr {
			t.Errorf("ToSQL(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			continue
		}

		if !tt.wantErr && sql != tt.expectedSQL {
			t.Errorf("ToSQL(%q) SQL = %q, want %q", tt.input, sql, tt.expectedSQL)
		}
	}
}

func TestPostgresPlaceholders(t *testing.T) {
	cfg := Config{Placeholder: "$"}
	input := `a == 1 && b == 2 || c in [3, 4]`
	// Expected: ((a = $1) AND (b = $2)) OR (c IN ($3, $4))
	expected := "(((a = $1) AND (b = $2)) OR (c IN ($3, $4)))"

	sql, args, err := ToSQL(input, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if sql != expected {
		t.Errorf("expected SQL=%q, got=%q", expected, sql)
	}

	if len(args) != 4 {
		t.Errorf("expected 4 args, got %d", len(args))
	}
}

func TestCaseInsensitivity(t *testing.T) {
	inputs := []string{
		`is_active == TRUE`,
		`is_active == True`,
		`is_active == true`,
		`id IN [1, 2]`,
		`id in [1, 2]`,
		`deleted_at == NULL`,
	}

	for _, input := range inputs {
		_, _, err := ToSQL(input)
		if err != nil {
			t.Errorf("input %q failed: %v", input, err)
		}
	}
}

func TestSchemaValidation(t *testing.T) {
	cfg := Config{
		Schema: Schema{
			"name":  FieldRule{Type: "string", Required: true},
			"age":   FieldRule{Type: "int", Error: "Age must be a whole number"},
			"price": FieldRule{Type: "float"},
		},
	}

	tests := []struct {
		input   string
		wantErr string
	}{
		{
			input:   `name == "John" && age == 25`,
			wantErr: "",
		},
		{
			input:   `name == "John" && price == 10.5`,
			wantErr: "",
		},
		{
			input:   `name == "John" && price == 10`, // Int matching float rule
			wantErr: "",
		},
		{
			input:   `age == 25`, // Missing required "name"
			wantErr: "field name is required",
		},
		{
			input:   `name == "John" && age == 25.5`, // Wrong type for age
			wantErr: "Age must be a whole number",
		},
		{
			input:   `name == 123`, // Wrong type for name
			wantErr: "expected type string, got int",
		},
	}

	for _, tt := range tests {
		_, _, err := ToSQL(tt.input, cfg)
		if tt.wantErr == "" {
			if err != nil {
				t.Errorf("ToSQL(%q) unexpected error: %v", tt.input, err)
			}
		} else {
			if err == nil {
				t.Errorf("ToSQL(%q) expected error %q, got nil", tt.input, tt.wantErr)
			} else if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("ToSQL(%q) expected error %q, got %q", tt.input, tt.wantErr, err.Error())
			}
		}
	}
}

func TestConfigBuilder(t *testing.T) {
	// 1. Service layer adds validation
	cfg := NewConfig().WithSchema(Schema{
		"user": FieldRule{Type: "int", Required: true},
	})

	// 2. Repository layer adds mapping and Postgres dialect
	cfg = cfg.WithMapping("user", "user_id").WithPostgres()

	sql, args, err := ToSQL("user == 100", cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if sql != "(user_id = $1)" {
		t.Errorf("expected SQL=(user_id = $1), got %q", sql)
	}

	if args[0] != int64(100) {
		t.Errorf("expected args[0]=100, got %v", args[0])
	}
}

func TestErrorChecking(t *testing.T) {
	cfg := NewConfig().WithSchema(Schema{
		"age": FieldRule{Type: "int", Required: true},
	})

	t.Run("Check ErrRequiredField", func(t *testing.T) {
		_, _, err := ToSQL(`name == "John"`, cfg)
		if !errors.Is(err, ErrRequiredField) {
			t.Errorf("expected error to be ErrRequiredField, got %v", err)
		}
	})

	t.Run("Check ErrTypeMismatch", func(t *testing.T) {
		_, _, err := ToSQL(`age == "young"`, cfg)
		if !errors.Is(err, ErrTypeMismatch) {
			t.Errorf("expected error to be ErrTypeMismatch, got %v", err)
		}
	})

	t.Run("Check ErrFieldNotAllowed", func(t *testing.T) {
		restrictedCfg := NewConfig().WithAllowedFields([]string{"age"})
		_, _, err := ToSQL(`name == "John"`, restrictedCfg)
		if !errors.Is(err, ErrFieldNotAllowed) {
			t.Errorf("expected error to be ErrFieldNotAllowed, got %v", err)
		}
	})

	t.Run("Check ErrUnexpectedTokens", func(t *testing.T) {
		_, _, err := ToSQL(`age == 18 & name == "John"`, cfg)
		if !errors.Is(err, ErrUnexpectedTokens) {
			t.Errorf("expected error to be ErrUnexpectedTokens, got %v", err)
		}
	})
}

func TestQueryTranspiler(t *testing.T) {
	var _ QueryTranspiler = NewSQLBackend()

	backend := NewSQLBackend()
	node, err := backend.Parse("a == 1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	query, args, err := backend.Transpile(node, NewConfig())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if query != "(a = ?)" {
		t.Errorf("expected query=(a = ?), got %q", query)
	}

	if len(args) != 1 || args[0] != int64(1) {
		t.Errorf("expected arg 1, got %v", args)
	}
}

func TestExtraBehavior(t *testing.T) {
	backend := NewSQLBackend()
	node, _ := backend.Parse(`name == "Alice"`)

	// Service layer adds extra behavior: AND org_id == 123
	orgNode := &ast.InfixExpression{
		Left:     node.(ast.Expression),
		Operator: "&&",
		Right: &ast.InfixExpression{
			Left:     &ast.Identifier{Value: "org_id"},
			Operator: "==",
			Right:    &ast.Literal{Value: "123", Type: ast.IntegerLiteral},
		},
	}

	query, args, _ := backend.Transpile(orgNode, NewConfig())
	// Expected: ((name = ?) AND (org_id = ?))
	if query != "((name = ?) AND (org_id = ?))" {
		t.Errorf("expected query with extra behavior, got %q", query)
	}
	if len(args) != 2 || args[1] != int64(123) {
		t.Errorf("expected 2 args, second being 123, got %v", args)
	}
}
