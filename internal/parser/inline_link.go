package parser

import "orische/internal/ast"

type linkInlineDefinition struct{}

func (*linkInlineDefinition) contentPolicy() inlineContentPolicy {
	return inlineContentNested
}

func (*linkInlineDefinition) validateAttribute(attribute string) (bool, error) {
	return attribute != "", nil
}

func (*linkInlineDefinition) buildInline(candidate inlineDirectiveCandidate) (ast.Inline, error) {
	return &ast.Link{
		URI:     candidate.Attribute,
		Content: candidate.NestedContent,
		Range:   candidate.Range,
	}, nil
}
