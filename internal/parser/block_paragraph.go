package parser

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"orische/internal/ast"
)

const blockBuilderKeyParagraph = "paragraph"

type paragraphReader struct{}

func (*paragraphReader) read(ctx *blockContext) (parsedBlockNode, bool, error) {
	var line string
	startPos := ctx.getPos() + 1
	content := []string{}

	for !ctx.isEOF() {
		line = ctx.getLine()
		if strings.TrimSpace(line) == "" {
			break
		}
		content = append(content, line)
		ctx.advance(1)
	}

	ctx.advance(-1)

	return &parsedBlock{
		Type: blockBuilderKeyParagraph,
		Attr: "",
		Text: strings.Join(content, "\n"),
		Range: ast.Range{
			Start: ast.Position{Line: startPos, Column: 1},
			End:   ast.Position{Line: ctx.getPos() + 1, Column: utf8.RuneCountInString(content[len(content)-1])},
		},
	}, true, nil
}

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
