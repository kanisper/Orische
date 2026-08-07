package parser

import (
	"strings"

	"orische/internal/ast"
)

type blockDirectiveParser struct{}

func (*blockDirectiveParser) parse(ctx *blockContext) (parsedBlockNode, bool, error) {
	if !isBlockDirective(ctx.getLine()) {
		return nil, false, nil
	}

	dirtype, attr, ok := parseBlockDirective(ctx.getLine())
	if !ok {
		return nil, false, nil
	}

	startPos := ctx.getPos()
	startLine := startPos + 1
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
					Start: ast.Position{Line: startLine, Column: 1},
					End:   ast.Position{Line: ctx.getPos() + 1, Column: 3},
				},
			}, true, nil
		}
		content = append(content, line)
		ctx.advance(1)
	}

	ctx.setPos(startPos)
	return nil, false, nil
}

func isBlockDirective(line string) bool {
	return strings.HasPrefix(line, ":::[") && strings.HasSuffix(line, "]")
}

func parseBlockDirective(line string) (string, string, bool) {
	trimmed := strings.TrimPrefix(line, ":::[")
	trimmed = strings.TrimSuffix(trimmed, "]")
	idx := strings.Index(trimmed, ":")
	if idx < 0 {
		if trimmed == "" {
			return "", "", false
		}
		return trimmed, "", true
	} else if idx == 0 {
		return "", "", false
	} else {
		return trimmed[:idx], trimmed[idx+1:], true
	}
}
