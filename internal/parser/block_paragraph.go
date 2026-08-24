package parser

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"orische/internal/ast"
)

const blockTypeParagraph = "paragraph"

type paragraphDefinition struct{}

func (*paragraphDefinition) blockType() string {
	return blockTypeParagraph
}

func (*paragraphDefinition) read(ctx *blockContext) (parsedBlockNode, bool, error) {
	startLine := ctx.pos + 1
	var content []string

	for !ctx.isEOF() {
		line := ctx.line()
		if strings.TrimSpace(line) == "" {
			break
		}
		content = append(content, line)
		ctx.pos++
	}

	ctx.pos--

	return &parsedBlock{
		Type: blockTypeParagraph,
		Attr: "",
		Text: strings.Join(content, "\n"),
		Range: ast.Range{
			Start: ast.Position{Line: startLine, Column: 1},
			End:   ast.Position{Line: ctx.pos + 1, Column: utf8.RuneCountInString(content[len(content)-1])},
		},
	}, true, nil
}

func (*paragraphDefinition) build(parser *Parser, node parsedBlockNode) (ast.Block, error) {
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
