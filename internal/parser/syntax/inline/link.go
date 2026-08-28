package inline

import (
	"orische/internal/ast"
	"orische/internal/parser/feature"
)

type linkDefinition struct{}

func (*linkDefinition) InlineType() string {
	return "link"
}

func (*linkDefinition) ContentPolicy() feature.InlineContentPolicy {
	return feature.InlineContentNested
}

func (*linkDefinition) ValidateAttribute(attribute string) (bool, error) {
	return attribute != "", nil
}

func (*linkDefinition) BuildInline(candidate feature.InlineDirectiveCandidate) (ast.Inline, error) {
	return &ast.Link{
		URI:     candidate.Attribute,
		Content: candidate.NestedContent,
		Range:   candidate.Range,
	}, nil
}
