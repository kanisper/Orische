package parser

import (
	"fmt"

	"orische/internal/ast"
)

type listBuilder struct{}

func (*listBuilder) build(parser *Parser, node parsedBlockNode) (ast.Block, error) {
	list, ok := node.(*parsedList)
	if !ok {
		return nil, fmt.Errorf("expected *parsedList, got %T", node)
	}

	return buildList(parser, list)
}

func buildList(parser *Parser, pl *parsedList) (*ast.List, error) {
	list := &ast.List{
		Ordered: pl.Ordered,
		Items:   make([]*ast.ListItem, 0, len(pl.Items)),
		Range:   pl.Range,
	}

	for _, parsedItem := range pl.Items {
		item, err := buildListItem(parser, parsedItem)
		if err != nil {
			return nil, err
		}

		list.Items = append(list.Items, item)
	}

	return list, nil
}

func buildListItem(parser *Parser, parsedItem parsedListItem) (*ast.ListItem, error) {
	blocks := make([]ast.Block, 0, len(parsedItem.Blocks))

	for _, node := range parsedItem.Blocks {
		block, err := parser.buildBlock(node)
		if err != nil {
			return nil, err
		}

		blocks = append(blocks, block)
	}

	return &ast.ListItem{
		Blocks: blocks,
		Range:  parsedItem.Range,
	}, nil
}
