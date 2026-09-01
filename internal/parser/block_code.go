package parser

import "orische/internal/ast"

const typeCodeBlock = "code"

func buildCodeBlock(_ *Parser, block *blockDirectiveNode) (ast.Block, error) {
	return &ast.CodeBlock{
		Language: block.attribute,
		Text:     block.text,
		Range:    block.rng,
	}, nil
}
