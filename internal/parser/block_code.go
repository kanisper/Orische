package parser

import (
	"fmt"

	"orische/internal/ast"
)

const blockTypeCode = "code"

type codeBlockDefinition struct{}

func (*codeBlockDefinition) blockType() string {
	return blockTypeCode
}

func (*codeBlockDefinition) build(_ *Parser, node parsedBlockNode) (ast.Block, error) {
	block, ok := node.(*parsedBlock)
	if !ok {
		return nil, fmt.Errorf("expected *parsedBlock, got %T", node)
	}

	return &ast.CodeBlock{
		Language: block.Attr,
		Text:     block.Text,
		Range:    block.Range,
	}, nil
}
