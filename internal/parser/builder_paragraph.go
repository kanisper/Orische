package parser

import (
	"fmt"

	"orische/internal/ast"
)

type paragraphBuilder struct{}

func (*paragraphBuilder) build(node parsedBlockNode) (ast.Block, error) {
	block, ok := node.(*parsedBlock)
	if !ok {
		return nil, fmt.Errorf("expected *parsedBlock, got %T", node)
	}

	contents, err := parseInlines(block.Text)
	if err != nil {
		return nil, err
	}

	return &ast.Paragraph{
		Content: contents,
		Range:   block.Range,
	}, nil
}
