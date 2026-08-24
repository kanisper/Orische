package parser

import "orische/internal/ast"

const inlineTypeLink = "link"

type linkInlineDefinition struct{}

func (*linkInlineDefinition) inlineType() string {
	return inlineTypeLink
}

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
