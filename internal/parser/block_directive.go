package parser

import (
	"strings"

	"orische/internal/ast"
)

type blockDirectiveReader struct{}

func (*blockDirectiveReader) read(ctx *blockContext) (parsedBlockNode, bool, error) {
	dirtype, attr, ok := parseBlockDirective(ctx.line())
	if !ok {
		return nil, false, nil
	}

	startPos := ctx.pos
	startLine := startPos + 1
	ctx.pos++

	var content []string
	for !ctx.isEOF() {
		line := ctx.line()
		if line == ":::" {
			return &parsedBlock{
				Type: dirtype,
				Attr: attr,
				Text: strings.Join(content, "\n"),
				Range: ast.Range{
					Start: ast.Position{Line: startLine, Column: 1},
					End:   ast.Position{Line: ctx.pos + 1, Column: 3},
				},
			}, true, nil
		}
		content = append(content, line)
		ctx.pos++
	}

	// Rejected readers must not consume input.
	ctx.pos = startPos
	return nil, false, nil
}

func parseBlockDirective(line string) (string, string, bool) {
	if !strings.HasPrefix(line, ":::[") || !strings.HasSuffix(line, "]") {
		return "", "", false
	}

	dirtype, attr, _ := strings.Cut(line[4:len(line)-1], ":")
	return dirtype, attr, dirtype != ""
}
