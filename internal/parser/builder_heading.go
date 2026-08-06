package parser

import (
	"fmt"
	"strconv"

	"orische/internal/ast"
)

type headingBuilder struct{}

func (*headingBuilder) build(node parsedBlockNode) (ast.Block, error) {
	block, ok := node.(*parsedBlock)
	if !ok {
		return nil, fmt.Errorf("expected *parsedBlock, got %T", node)
	}

	level, err := strconv.Atoi(block.Attr[len(block.Attr)-1:])
	if err != nil {
		return nil, fmt.Errorf("invalid attribute %s: %w", block.Attr, err)
	}

	content, err := parseInlines(block.Text)
	if err != nil {
		return nil, err
	}

	return &ast.Heading{
		Level:   level,
		Content: content,
		Range:   block.Range,
	}, nil
}
