package parser

import "orische/internal/ast"

var emphasisDefinition = inlineDefinition{
	policy: inlineContentNested,
	build: func(candidate inlineCandidate) ast.Inline {
		return &ast.Emphasis{
			Content: candidate.nestedContent,
			Range:   candidate.rng,
		}
	},
}
