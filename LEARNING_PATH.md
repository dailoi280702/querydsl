# Understanding QueryDSL: A Reading Roadmap

This guide explains the flow of data through the QueryDSL library. To understand the project, read the files in this specific order.

## 1. The Blueprint: `parser/ast/ast.go`
**Start here.** This file defines what a query looks like once it's been "understood" by the computer.
*   **Concepts:** Nodes and Trees.
*   **Key Node:** `InfixExpression`. It represents `Left Operator Right` (e.g., `age > 18`).
*   **Key Node:** `PrefixExpression`. It handles operators that come *before* a value, such as negative numbers (`-5`) or logical negation (`!true`).
*   **Enums:** Notice `LiteralType`. We use an enum (Integer, Float, String, etc.) so the computer doesn't have to "guess" the data type later in the pipeline.

## 2. The Slicer: `lexer/lexer.go`
The Lexer turns a raw string into a list of "Tokens" (words).
*   **The Job:** Scanning. It looks at the string character by character.
*   **Key Function:** `NextToken()`. It ignores whitespace and groups characters together (like turning `=` and `=` into a single `EQ` token).
*   **Note:** The Lexer doesn't care about grammar; it just identifies words.

## 3. The Brain: `parser/parser.go`
This is the most complex part. It uses the **Pratt Parsing** algorithm.
*   **The Job:** Turning a flat list of words into the Tree (AST) defined in Step 1.
*   **Precedence:** This is where we define that `&&` happens before `||`.
*   **Recursion:** If the parser sees a `(`, it pauses and starts a new parser "inside" itself to handle the grouped logic.

## 4. The Translator: `compiler/sql/sql.go`
Now we transform the Tree into SQL.
*   **The Job:** Generation. We "walk" the tree from bottom to top.
*   **Security:** This is where SQL Injection is prevented. Notice that literals return a `?` and the actual value is added to an `Args` slice.
*   **Dialects:** This is where we handle the difference between MySQL (`?`) and Postgres (`$1`).

## 5. The Gatekeeper: `validation.go`
This is the business logic layer.
*   **The Job:** Rules enforcement.
*   **Logic:** It walks the tree just like the compiler, but instead of strings, it checks if your input matches your `Schema`.
*   **Flexible Types:** This is where we allow an `Integer` to pass for a `Float` field, but reject a `String` for an `Integer` field.

## 6. The Real World: `examples/layered/main.go`
This shows how to use the library in a professional application.
*   **Service Layer:** Focuses on *what* data is allowed (Schema).
*   **Repository Layer:** Focuses on *how* that data is stored (Mapping names and SQL dialect).

---

### How to add a new feature
- **Infix (e.g., `BETWEEN`):** Requires updating AST, Lexer, Parser (`registerInfix`), and Compiler.
- **Prefix (e.g., `EXISTS`):** Requires updating AST, Lexer, Parser (`registerPrefix`), and Compiler.
- **Literals (e.g., `DATE`):** Add a new `LiteralType` enum and update the Lexer to recognize the format.
