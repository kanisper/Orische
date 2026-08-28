package block

import (
	"fmt"

	"orische/internal/ast"
	"orische/internal/parser/feature"
)

type paragraphDefinition struct{}

func (*paragraphDefinition) BlockType() string {
	return feature.ParagraphBlockType
}

func (*paragraphDefinition) BuildParagraph(ctx feature.BuildContext, node feature.BlockNode) (*ast.Paragraph, error) {
	block, ok := node.(*feature.TextBlock)
	if !ok {
		return nil, fmt.Errorf("expected *feature.TextBlock, got %T", node)
	}

	content, err := ctx.ParseInlines(block.Text, block.ContentOrigin)
	if err != nil {
		return nil, err
	}

	return &ast.Paragraph{
		Content: content,
		Range:   block.Range,
	}, nil
}
