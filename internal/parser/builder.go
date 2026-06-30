package parser

import "medoc/internal/ast"

type blockBuilder interface {
	build(block parsedBlock, p *Parser) (ast.Block, error)
}
