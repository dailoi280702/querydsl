package querydsl

import "github.com/dailoi280702/querydsl/parser/ast"

// CombineAnd joins multiple AST nodes with the logical AND operator.
// It handles nil nodes gracefully.
func CombineAnd(nodes ...ast.Node) ast.Node {
	var result ast.Node
	for _, node := range nodes {
		if node == nil {
			continue
		}
		if result == nil {
			result = node
			continue
		}
		result = &ast.InfixExpression{
			Left:     result.(ast.Expression),
			Operator: "&&",
			Right:    node.(ast.Expression),
		}
	}
	return result
}
