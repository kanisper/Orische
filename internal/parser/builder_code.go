package parser

import (
	"fmt"

	"medoc/internal/ast"
)

type codeBlockBuilder struct{}

func (*codeBlockBuilder) build(node parsedBlockNode) (ast.Block, error) {
	block, ok := node.(*parsedBlock)
	if !ok {
		return nil, fmt.Errorf("codeBlock Builder: expected *parsedBlock, got %T", node)
	}

	return &ast.CodeBlock{
		Language: block.Attr,
		Text:     block.Text,
		Range:    block.Range,
	}, nil
}
