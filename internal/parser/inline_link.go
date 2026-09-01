package parser

import "orische/internal/ast"

var linkDefinition = inlineDefinition{
	policy: inlineContentNested,
	validate: func(attribute string) bool {
		return attribute != ""
	},
	build: func(candidate inlineCandidate) ast.Inline {
		return &ast.Link{
			URI:     candidate.attribute,
			Content: candidate.nestedContent,
			Range:   candidate.rng,
		}
	},
}
