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

var underlineDefinition = inlineDefinition{
	policy: inlineContentNested,
	build: func(candidate inlineCandidate) ast.Inline {
		return &ast.Underline{Content: candidate.nestedContent, Range: candidate.rng}
	},
}

var strikethroughDefinition = inlineDefinition{
	policy: inlineContentNested,
	build: func(candidate inlineCandidate) ast.Inline {
		return &ast.Strikethrough{Content: candidate.nestedContent, Range: candidate.rng}
	},
}
