package parser

import (
	"fmt"

	"orische/internal/ast"
)

type paragraphBuilder struct{}

func (*paragraphBuilder) build(parser *Parser, node parsedBlockNode) (ast.Block, error) {
	block, ok := node.(*parsedBlock)
	if !ok {
		return nil, fmt.Errorf("expected *parsedBlock, got %T", node)
	}

	origin := ast.Position{
		Line:   block.Range.Start.Line,
		Column: block.Range.Start.Column,
	}
	contents, err := parser.parseInlines(block.Text, origin)
	if err != nil {
		return nil, err
	}

	return &ast.Paragraph{
		Content: contents,
		Range:   block.Range,
	}, nil
}
