package parser

import "orische/internal/ast"

type emphasisInlineDefinition struct{}

func (*emphasisInlineDefinition) contentPolicy() inlineContentPolicy {
	return inlineContentNested
}

func (*emphasisInlineDefinition) validateAttribute(string) (bool, error) {
	return true, nil
}

func (*emphasisInlineDefinition) buildInline(candidate inlineDirectiveCandidate) (ast.Inline, error) {
	return &ast.Emphasis{
		Content: candidate.NestedContent,
		Range:   candidate.Range,
	}, nil
}
