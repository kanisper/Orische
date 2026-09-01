package parser

import "orische/internal/ast"

// inlineContentPolicy selects recursive or literal content parsing.
type inlineContentPolicy uint8

const (
	inlineContentNested inlineContentPolicy = iota + 1
	inlineContentLiteral
)

// inlineCandidate is a structurally closed directive. The parser fills only
// the content field selected by the definition's policy before building it.
type inlineCandidate struct {
	attribute      string
	nestedContent  []ast.Inline
	literalContent string
	rng            ast.Range
}

// inlineDefinition keeps the small, type-specific part of inline parsing
// together. The map key in spec is the normalized directive type.
type inlineDefinition struct {
	policy   inlineContentPolicy
	validate func(string) bool
	build    func(inlineCandidate) ast.Inline
}
