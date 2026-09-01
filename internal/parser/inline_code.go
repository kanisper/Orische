package parser

import "orische/internal/ast"

var codeDefinition = inlineDefinition{
	policy: inlineContentLiteral,
	build: func(candidate inlineCandidate) ast.Inline {
		return &ast.CodeSpan{
			Value: candidate.literalContent,
			Range: candidate.rng,
		}
	},
}
