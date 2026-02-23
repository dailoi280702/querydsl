//nolint:revive
package elastic

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/dailoi280702/querydsl/lexer"
	"github.com/dailoi280702/querydsl/parser"
)

func TestCompile(t *testing.T) {
	tests := []struct {
		input    string
		expected string // JSON string for easy comparison
	}{
		{
			input:    `name == "john"`,
			expected: `{"term":{"name":"john"}}`,
		},
		{
			input:    `age >= 18`,
			expected: `{"range":{"age":{"gte":18}}}`,
		},
		{
			input:    `name == "john" && age >= 18`,
			expected: `{"bool":{"must":[{"term":{"name":"john"}},{"range":{"age":{"gte":18}}}]}}`,
		},
		{
			input:    `status == "active" || status == "pending"`,
			expected: `{"bool":{"minimum_should_match":1,"should":[{"term":{"status":"active"}},{"term":{"status":"pending"}}]}}`,
		},
		{
			input:    `!(status == "deleted")`,
			expected: `{"bool":{"must_not":{"term":{"status":"deleted"}}}}`,
		},
		{
			input:    `name % "john"`,
			expected: `{"match":{"name":{"fuzziness":"AUTO","query":"john"}}}`,
		},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := parser.New(l)
		node := p.ParseExpression(parser.LOWEST)

		c := New(Config{})
		result, err := c.Compile(node)
		if err != nil {
			t.Fatalf("Compile(%q) unexpected error: %v", tt.input, err)
		}

		jsonBytes, _ := json.Marshal(result)
		got := string(jsonBytes)

		if !jsonEqual(got, tt.expected) {
			t.Errorf("Compile(%q)\ngot:  %s\nwant: %s", tt.input, got, tt.expected)
		}
	}
}

func jsonEqual(a, b string) bool {
	var j1, j2 interface{}
	if err := json.Unmarshal([]byte(a), &j1); err != nil {
		return false
	}
	if err := json.Unmarshal([]byte(b), &j2); err != nil {
		return false
	}
	return reflect.DeepEqual(j1, j2)
}

func TestFieldMapping(t *testing.T) {
	l := lexer.New(`name == "john"`)
	p := parser.New(l)
	node := p.ParseExpression(parser.LOWEST)

	cfg := Config{
		FieldMap: map[string]string{"name": "full_name"},
	}
	c := New(cfg)
	result, err := c.Compile(node)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := `{"term":{"full_name":"john"}}`
	jsonBytes, _ := json.Marshal(result)
	if !jsonEqual(string(jsonBytes), expected) {
		t.Errorf("got %s, want %s", string(jsonBytes), expected)
	}
}

func TestGetValueTypes(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`a == 10`, `{"term":{"a":10}}`},
		{`b == 10.5`, `{"term":{"b":10.5}}`},
		{`c == true`, `{"term":{"c":true}}`},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := parser.New(l)
		node := p.ParseExpression(parser.LOWEST)

		c := New(Config{})
		result, err := c.Compile(node)
		if err != nil {
			t.Fatalf("Compile(%q) error: %v", tt.input, err)
		}

		jsonBytes, _ := json.Marshal(result)
		if !jsonEqual(string(jsonBytes), tt.expected) {
			t.Errorf("Compile(%q)\ngot:  %s\nwant: %s", tt.input, string(jsonBytes), tt.expected)
		}
	}
}
