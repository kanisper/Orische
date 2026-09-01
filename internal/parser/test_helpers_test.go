package parser

import (
	"testing"

	"orische/internal/ast"
)

func testRange(startLine, startColumn, endLine, endColumn int) ast.Range {
	return ast.Range{
		Start: ast.Position{Line: startLine, Column: startColumn},
		End:   ast.Position{Line: endLine, Column: endColumn},
	}
}

func testText(value string, startLine, startColumn, endLine, endColumn int) ast.Inline {
	return &ast.Text{
		Value: value,
		Range: testRange(startLine, startColumn, endLine, endColumn),
	}
}

func mustCoreParser(t testing.TB) *Parser {
	t.Helper()
	return NewParser()
}

func parserWithInlineDefinitions(t testing.TB, definitions map[string]inlineDefinition) *Parser {
	t.Helper()
	p := NewParser()
	for typ, definition := range definitions {
		p.spec.inlineDefinitions[normalizeSyntaxType(typ)] = definition
	}
	return p
}
