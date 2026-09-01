package parser

import (
	"testing"
)

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
