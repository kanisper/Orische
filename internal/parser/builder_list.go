package parser

import (
	"fmt"
	"strings"

	"orische/internal/ast"
)

type listBuilder struct{}

func (*listBuilder) build(node parsedBlockNode) (ast.Block, error) {
	list, ok := node.(*parsedList)
	if !ok {
		return nil, fmt.Errorf("expected *parsedList, got %T", node)
	}

	return buildList(list)
}

func buildList(pl *parsedList) (*ast.List, error) {
	list := &ast.List{
		Ordered: pl.Ordered,
		Items:   make([]*ast.ListItem, 0, len(pl.Items)),
		Range:   pl.Range,
	}

	for _, parsedItem := range pl.Items {
		item, err := buildListItem(parsedItem)
		if err != nil {
			return nil, err
		}

		list.Items = append(list.Items, item)
	}

	return list, nil
}

func buildListItem(parsedItem parsedListItem) (*ast.ListItem, error) {
	blocks := make([]ast.Block, 0, len(parsedItem.Blocks))

	for _, node := range parsedItem.Blocks {
		block, err := buildListItemBlock(node, parsedItem.RawLevel)
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

func buildListItemBlock(node parsedBlockNode, itemRawLevel int) (ast.Block, error) {
	switch block := node.(type) {
	case *parsedBlock:
		if !strings.EqualFold(block.Type, "paragraph") {
			return nil, fmt.Errorf("unexpected block type in list item: %q", block.Type)
		}

		origin := block.getBlockRange().Start
		inlines, err := parseInlines(block.Text, origin)
		if err != nil {
			return nil, err
		}

		return &ast.Paragraph{
			Content: inlines,
			Range:   block.Range,
		}, nil

	case *parsedList:
		nested, err := buildList(block)
		if err != nil {
			return nil, err
		}

		return nested, nil

	default:
		return nil, fmt.Errorf("unsupported type of node in list item: %T", node)
	}
}
