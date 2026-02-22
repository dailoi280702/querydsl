// Package main demonstrates a layered architecture using the QueryDSL library.
package main

import (
	"context"
	"fmt"
	"github.com/dailoi280702/querydsl"
	"github.com/dailoi280702/querydsl/parser/ast"
)

// --- Domain/Service Layer ---

type User struct {
	ID   int
	Name string
	Age  int
}

// UserUsecase handles business logic for users.
type UserUsecase struct {
	repo        UserRepository
	transpiler  querydsl.QueryTranspiler
	querySchema querydsl.Schema
}

func NewUserUsecase(repo UserRepository, t querydsl.QueryTranspiler) *UserUsecase {
	return &UserUsecase{
		repo:       repo,
		transpiler: t,
		// Business rule: define what fields are allowed and their types
		querySchema: querydsl.Schema{
			"name": querydsl.FieldRule{Type: "string", Required: true},
			"age":  querydsl.FieldRule{Type: "int"},
		},
	}
}

func (u *UserUsecase) SearchUsers(ctx context.Context, dslQuery string, orgID string) ([]User, error) {
	// 1. Parse into AST Node
	node, err := u.transpiler.Parse(dslQuery)
	if err != nil {
		return nil, fmt.Errorf("parse error: %w", err)
	}

	// 2. Extra behavior: Inject mandatory org_id filter
	// dslQuery = (dslQuery) && (org_id == orgID)
	node = &ast.InfixExpression{
		Left:     node.(ast.Expression),
		Operator: "&&",
		Right: &ast.InfixExpression{
			Left:     &ast.Identifier{Value: "org_id"},
			Operator: "==",
			Right:    &ast.Literal{Value: orgID, Type: ast.IntegerLiteral},
		},
	}

	// 3. Start with business validation (Schema)
	cfg := querydsl.NewConfig().WithSchema(u.querySchema)

	// 4. Delegate to Repository with the modified AST
	return u.repo.FindAll(ctx, node, cfg)
}

// --- Repository Layer ---

type UserRepository interface {
	FindAll(ctx context.Context, node ast.Node, cfg querydsl.Config) ([]User, error)
}

type UserPostgresRepo struct {
	transpiler querydsl.QueryTranspiler
}

func (r *UserPostgresRepo) FindAll(_ context.Context, node ast.Node, cfg querydsl.Config) ([]User, error) {
	// 1. Add Infrastructure details (Field Mapping + Dialect)
	cfg = cfg.WithMapping("name", "full_name").WithPostgres()

	// 2. Perform the actual transpilation using the node
	where, args, err := r.transpiler.Transpile(node, cfg)
	if err != nil {
		return nil, err
	}

	// 3. Execute the safe query
	fullQuery := fmt.Sprintf("SELECT id, full_name, age FROM users WHERE %s", where)
	fmt.Println("Repository Executing SQL:", fullQuery)
	fmt.Printf("Repository Using Args: %v\n", args)

	// Mocking DB response
	return []User{{ID: 1, Name: "Alice", Age: 25}}, nil
}

// --- Transport/Controller Layer ---

func main() {
	// 1. Dependency Injection
	sqlBackend := querydsl.NewSQLBackend()
	repo := &UserPostgresRepo{transpiler: sqlBackend}
	usecase := NewUserUsecase(repo, sqlBackend)

	// 2. Raw input from user (e.g., from URL ?q=...)
	userInput := `name == "Alice" && age > 20`
	orgID := "123"
	fmt.Printf("User Input DSL: %s (OrgID: %s)\n", userInput, orgID)

	// 3. Call Usecase
	users, err := usecase.SearchUsers(context.Background(), userInput, orgID)
	if err != nil {
		fmt.Printf("Search Error: %v\n", err)
		return
	}

	fmt.Printf("Final Result: %+v\n", users)
}
