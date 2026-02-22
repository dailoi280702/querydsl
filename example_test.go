package querydsl_test

import (
	"errors"
	"fmt"
	"querydsl"
	"querydsl/parser/ast"
)

func ExampleQueryTranspiler() {
	// In a real app, this would be injected into your Service
	var transpiler querydsl.QueryTranspiler = querydsl.NewSQLBackend()

	// 1. Service layer defines business validation
	userSchema := querydsl.Schema{
		"name": querydsl.FieldRule{Type: "string", Required: true},
	}

	// 2. Parse into AST Node
	input := `name == "John"`
	node, err := transpiler.Parse(input)
	if err != nil {
		return
	}

	// 3. Extra behavior: Add mandatory filter
	orgIDNode := &ast.InfixExpression{
		Left:     node.(ast.Expression),
		Operator: "&&",
		Right: &ast.InfixExpression{
			Left:     &ast.Identifier{Value: "org_id"},
			Operator: "==",
			Right:    &ast.Literal{Value: "1", Type: ast.IntegerLiteral},
		},
	}

	// 4. Validate against schema
	if err := transpiler.Validate(orgIDNode, userSchema); err != nil {
		fmt.Printf("Validation failed: %v\n", err)
		return
	}

	// 5. Repository layer adds DB mapping
	cfg := querydsl.NewConfig().WithMapping("name", "full_name").WithPostgres()

	// 6. Transpile the modified AST
	query, args, err := transpiler.Transpile(orgIDNode, cfg)
	if err != nil {
		return
	}

	fmt.Println("Query:", query)
	fmt.Printf("Args: %v\n", args)

	// Output:
	// Query: ((full_name = $1) AND (org_id = $2))
	// Args: [John 1]
}

func ExampleToSQL_errorHandling() {
	cfg := querydsl.NewConfig().WithAllowedFields([]string{"id"})

	// Trying to access a field that isn't allowed
	_, _, err := querydsl.ToSQL(`password == "123456"`, cfg)

	if errors.Is(err, querydsl.ErrFieldNotAllowed) {
		fmt.Println("Caught: Field not allowed")
	}

	// Output:
	// Caught: Field not allowed
}
