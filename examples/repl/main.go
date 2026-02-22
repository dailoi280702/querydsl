// Package main provides a REPL for testing and debugging QueryDSL strings.
package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/dailoi280702/querydsl"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("QueryDSL REPL")
	fmt.Println("Type 'exit' to quit")
	fmt.Println("-----------------")

	// Standard config for debugging
	cfg := querydsl.NewConfig().WithPostgres()

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
