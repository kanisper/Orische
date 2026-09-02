package parser

import "orische/internal/ast"

var strongDefinition = inlineDefinition{
	policy: inlineContentNested,
	build: func(candidate inlineCandidate) ast.Inline {
		return &ast.Strong{Content: candidate.nestedContent, Range: candidate.rng}
	},
}

var italicDefinition = inlineDefinition{
	policy: inlineContentNested,
	build: func(candidate inlineCandidate) ast.Inline {
		return &ast.Italic{Content: candidate.nestedContent, Range: candidate.rng}
	},
}

var boldDefinition = inlineDefinition{
	policy: inlineContentNested,
	build: func(candidate inlineCandidate) ast.Inline {
		return &ast.Bold{Content: candidate.nestedContent, Range: candidate.rng}
	},
}

var deletedDefinition = inlineDefinition{
	policy: inlineContentNested,
	build: func(candidate inlineCandidate) ast.Inline {
		return &ast.Deleted{Content: candidate.nestedContent, Range: candidate.rng}
	},
}

var outdatedDefinition = inlineDefinition{
	policy: inlineContentNested,
	build: func(candidate inlineCandidate) ast.Inline {
		return &ast.Outdated{Content: candidate.nestedContent, Range: candidate.rng}
	},
}
