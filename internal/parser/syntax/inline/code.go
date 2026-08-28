package inline

import (
	"orische/internal/ast"
	"orische/internal/parser/feature"
)

type codeDefinition struct{}

func (*codeDefinition) InlineType() string {
	return "code"
}

func (*codeDefinition) ContentPolicy() feature.InlineContentPolicy {
	return feature.InlineContentLiteral
}

func (*codeDefinition) ValidateAttribute(string) (bool, error) {
	return true, nil
}

func (*codeDefinition) BuildInline(candidate feature.InlineDirectiveCandidate) (ast.Inline, error) {
	return &ast.CodeSpan{
		Value: candidate.LiteralContent,
		Range: candidate.Range,
	}, nil
}
