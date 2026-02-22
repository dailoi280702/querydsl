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

// GetFields extracts all unique field names (identifiers) from an AST node.
func GetFields(node ast.Node) []string {
	fieldMap := make(map[string]bool)
	walkFields(node, fieldMap)

	fields := make([]string, 0, len(fieldMap))
	for field := range fieldMap {
		fields = append(fields, field)
	}
	return fields
}

func walkFields(node ast.Node, fieldMap map[string]bool) {
	switch n := node.(type) {
	case *ast.InfixExpression:
		walkFields(n.Left, fieldMap)
		walkFields(n.Right, fieldMap)
	case *ast.PrefixExpression:
		walkFields(n.Right, fieldMap)
	case *ast.Identifier:
		fieldMap[n.Value] = true
	case *ast.ArrayLiteral:
		for _, e := range n.Elements {
			walkFields(e, fieldMap)
		}
	}
}
