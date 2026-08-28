package block

import (
	"fmt"

	"orische/internal/ast"
	"orische/internal/parser/feature"
)

const typeCode = "code"

type codeDefinition struct{}

func (*codeDefinition) BlockType() string {
	return typeCode
}

func (*codeDefinition) BuildBlock(_ feature.BuildContext, node feature.BlockNode) (ast.Block, error) {
	block, ok := node.(*feature.TextBlock)
	if !ok {
		return nil, fmt.Errorf("expected *feature.TextBlock, got %T", node)
	}

	return &ast.CodeBlock{
		Language: block.Attr,
		Text:     block.Text,
		Range:    block.Range,
	}, nil
}
