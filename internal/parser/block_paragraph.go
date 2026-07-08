package parser

import (
	"strings"

	"medoc/internal/ast"
)

type paragraphParser struct{}

func (*paragraphParser) parse(ctx *blockContext) (parsedBlockNode, bool, error) {
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
		Type: "Paragraph",
		Attr: "",
		Text: strings.Join(content, "\n"),
		Range: ast.Range{
			StartLine: startPos,
			EndLine:   ctx.getPos() + 1,
		},
	}, true, nil
}
