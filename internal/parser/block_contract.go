package parser

import "orische/internal/ast"

// parsedBlock is the private handoff between block parsing and AST building.
// Concrete node types make invalid reader/IR combinations unrepresentable
// inside the parser.
type parsedBlock interface {
	blockRange() ast.Range
}

type paragraphNode struct {
	text          string
	contentOrigin ast.Position
	rng           ast.Range
}

func (n *paragraphNode) blockRange() ast.Range {
	return n.rng
}

type blockDirectiveNode struct {
	dirtype       string
	attribute     string
	text          string
	contentOrigin ast.Position
	rng           ast.Range
}

func (n *blockDirectiveNode) blockRange() ast.Range {
	return n.rng
}
