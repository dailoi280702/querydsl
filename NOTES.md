# Developer Notes & Advanced Patterns

This document contains implementation recipes and architectural advice for the QueryDSL library.

## 1. Handling Pagination

The library provides `CombineAnd` to merge user queries with system-level filters like cursors.

### Cursor Pagination (Recommended)
```go
// 1. Convert your cursor to an AST node
cursorNode := &ast.InfixExpression{
    Left:     &ast.Identifier{Value: "id"},
    Operator: ">",
    Right:    &ast.Literal{Value: lastID, Type: ast.IntegerLiteral},
}

// 2. Combine with user query
finalNode := querydsl.CombineAnd(userNode, cursorNode)

// 3. Transpile and execute
where, args, _ := transpiler.Transpile(finalNode, cfg)
query := fmt.Sprintf("SELECT * FROM table WHERE %s ORDER BY id ASC LIMIT 10", where)
```

### Offset Pagination
```go
where, args, _ := querydsl.ToSQL(userInput, cfg)
query := fmt.Sprintf("SELECT * FROM table WHERE %s LIMIT ? OFFSET ?", where)
args = append(args, limit, offset)
```

## 2. Multi-Tenancy (Force Filters)

You should always wrap user queries in a system-enforced filter to prevent data leaking between tenants.

```go
func (s *Service) List(ctx context.Context, tenantID int, userDSL string) {
    userNode, _ := querydsl.Parse(userDSL)
    
    tenantNode := &ast.InfixExpression{
        Left:     &ast.Identifier{Value: "tenant_id"},
        Operator: "==",
        Right:    &ast.Literal{Value: strconv.Itoa(tenantID), Type: ast.IntegerLiteral},
    }

    // This ensures the user can NEVER see data outside their tenant
    finalNode := querydsl.CombineAnd(tenantNode, userNode)
    
    where, args, _ := s.repo.Find(ctx, finalNode)
}
```

## 3. Customizing SQL Dialects

### PostgreSQL ($1, $2)
```go
cfg := querydsl.NewConfig().WithPostgres()
```

### Field Mapping (Security Aliasing)
Always map your JSON field names to DB column names in the Repository layer to hide your schema details.
```go
cfg := querydsl.NewConfig().
    WithMapping("userName", "full_name").
    WithMapping("userEmail", "email_address")
```

## 4. Using with Elasticsearch

The `ElasticBackend` returns a `map[string]any` which can be encoded directly for the official Go client.

```go
backend := querydsl.NewElasticBackend()
query, _ := backend.Transpile(node, cfg)

body := map[string]any{
    "query": query,
}

var buf bytes.Buffer
json.NewEncoder(&buf).Encode(body)

res, _ := es.Search(
    es.Search.WithIndex("my_index"),
    es.Search.WithBody(&buf),
)
```

## 5. Testing Strategy

### Mocking the Transpiler
Since `QueryTranspiler` is an interface, you can mock it in your Service layer tests to verify that the service is passing the correct `Schema` or adding the correct `tenant_id` filter.

### Validating the AST
Instead of checking the final SQL string (which can be brittle), check the structure of the AST node returned by `Parse`.

## 5. Performance Note
- The library uses `strings.Builder` for efficient string generation.
- The Parser uses a Pratt algorithm (O(n) complexity).
- For extremely high-throughput APIs, consider caching the result of `Parse(input)` if the same query string is used frequently.
