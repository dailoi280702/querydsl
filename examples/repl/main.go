// Package main provides a REPL for testing and debugging QueryDSL strings.
package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/dailoi280702/querydsl"
	"github.com/dailoi280702/querydsl/parser/ast"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("QueryDSL REPL (Postgres + Similarity Hook on 'name')")
	fmt.Println("Type 'exit' to quit")
	fmt.Println("-----------------")

	// Custom hook for similarity on 'name' field
	customHook := func(n *ast.InfixExpression, walk func(ast.Node, string) (string, error)) (string, bool, error) {
		if n.Operator == "%" {
			if ident, ok := n.Left.(*ast.Identifier); ok && ident.Value == "name" {
				left, _ := walk(n.Left, "")
				right, _ := walk(n.Right, "")
				return fmt.Sprintf("unaccent(lower(%s)) %% unaccent(lower(%s))", left, right), true, nil
			}
		}
		return "", false, nil
	}

	cfg := querydsl.NewConfig().WithPostgres().WithCustomInfix(customHook)

	for {
		fmt.Print("dsl> ")
		if !scanner.Scan() {
			break
		}

		line := scanner.Text()
		if line == "exit" {
			break
		}
		if line == "" {
			continue
		}

		// 1. Parse
		node, err := querydsl.Parse(line)
		if err != nil {
			fmt.Printf("\033[31mParse Error:\033[0m %v\n", err)
			continue
		}

		// 2. Show AST
		fmt.Printf("\033[34mAST:\033[0m %s\n", node.String())

		// 3. Transpile to SQL
		sql, args, err := querydsl.ToSQL(line, cfg)
		if err != nil {
			fmt.Printf("\033[31mTranspile Error:\033[0m %v\n", err)
			continue
		}

		fmt.Printf("\033[32mSQL:\033[0m %s\n", sql)
		fmt.Printf("\033[32mArgs:\033[0m %v\n", args)
		fmt.Println()
	}
}
