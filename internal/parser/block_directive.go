package parser

import (
	"strings"

	"medoc/internal/ast"
)

type blockDirectiveParser struct{}

func (*blockDirectiveParser) parse(ctx *blockContext) (parsedBlockNode, bool, error) {
	if !isBlockDirective(ctx.getLine()) {
		return &parsedBlock{}, false, nil
	}

	dirtype, attr, ok := parseBlockDirective(ctx.getLine())
	if !ok {
		return &parsedBlock{}, false, nil
	}

	startLine := ctx.getPos()
	ctx.advance(1)

	content := []string{}
	for !ctx.isEOF() {
		line := ctx.getLine()
		if line == ":::" {
			return &parsedBlock{
				Type: dirtype,
				Attr: attr,
				Text: strings.Join(content, "\n"),
				Range: ast.Range{
					StartLine: startLine,
					EndLine:   ctx.getPos(),
				},
			}, true, nil
		}
		content = append(content, line)
		ctx.advance(1)
	}
	return &parsedBlock{}, false, nil
}

func isBlockDirective(line string) bool {
	return strings.HasPrefix(line, ":::[") && strings.HasSuffix(line, "]")
}

func parseBlockDirective(line string) (string, string, bool) {
	trimmed := strings.TrimPrefix(line, ":::[")
	trimmed = strings.TrimSuffix(trimmed, "]")
	idx := strings.Index(trimmed, ":")
	if idx < 0 {
		return trimmed, "", true
	} else if idx == 0 {
		return "", "", false
	} else {
		return trimmed[:idx], trimmed[idx+1:], true
	}
}
