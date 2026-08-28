package feature

import "orische/internal/ast"

// InlineContentPolicy selects recursive or literal content parsing.
type InlineContentPolicy uint8

const (
	// InlineContentNested recursively parses directive content.
	InlineContentNested InlineContentPolicy = iota + 1
	// InlineContentLiteral preserves directive content as source text.
	InlineContentLiteral
)

// InlineDirectiveCandidate is a structurally closed directive with either
// NestedContent or LiteralContent populated according to its definition policy.
type InlineDirectiveCandidate struct {
	Attribute      string
	NestedContent  []ast.Inline
	LiteralContent string
	Range          ast.Range
}

// InlineDirectiveDefinition owns semantic validation, content policy, and AST
// construction. ValidateAttribute returns true to accept, false with a nil error
// for literal fallback, or a non-nil error to abort parsing. The parser frontend
// owns delimiters, recursion, fallback, and source ranges.
type InlineDirectiveDefinition interface {
	InlineType() string
	ContentPolicy() InlineContentPolicy
	ValidateAttribute(attribute string) (bool, error)
	BuildInline(InlineDirectiveCandidate) (ast.Inline, error)
}
