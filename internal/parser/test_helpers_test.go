package parser

import (
	"testing"

	"orische/internal/parser/syntax"
)

func mustCoreParser(t testing.TB) *Parser {
	t.Helper()

	p, err := NewParser(syntax.Core())
	if err != nil {
		t.Fatalf("NewParser(syntax.Core()) returned an error: %v", err)
	}
	return p
}
