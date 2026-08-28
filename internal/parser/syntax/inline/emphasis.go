package inline

import (
	"orische/internal/ast"
	"orische/internal/parser/feature"
)

type emphasisDefinition struct{}

func (*emphasisDefinition) InlineType() string {
	return "em"
}

func (*emphasisDefinition) ContentPolicy() feature.InlineContentPolicy {
	return feature.InlineContentNested
}

func (*emphasisDefinition) ValidateAttribute(string) (bool, error) {
	return true, nil
}

func (*emphasisDefinition) BuildInline(candidate feature.InlineDirectiveCandidate) (ast.Inline, error) {
	return &ast.Emphasis{
		Content: candidate.NestedContent,
		Range:   candidate.Range,
	}, nil
}
