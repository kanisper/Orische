package parser

import "orische/internal/ast"

type codeInlineDefinition struct{}

func (*codeInlineDefinition) contentPolicy() inlineContentPolicy {
	return inlineContentLiteral
}

func (*codeInlineDefinition) validateAttribute(string) (bool, error) {
	return true, nil
}

func (*codeInlineDefinition) buildInline(candidate inlineDirectiveCandidate) (ast.Inline, error) {
	return &ast.CodeSpan{
		Value: candidate.LiteralContent,
		Range: candidate.Range,
	}, nil
}
