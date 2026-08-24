package parser

import "orische/internal/ast"

const inlineTypeCode = "code"

type codeInlineDefinition struct{}

func (*codeInlineDefinition) inlineType() string {
	return inlineTypeCode
}

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
