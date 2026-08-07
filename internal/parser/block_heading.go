package parser

import (
	"strconv"
	"strings"

	"orische/internal/ast"
)

type headingParser struct{}

func (*headingParser) parse(ctx *blockContext) (parsedBlockNode, bool, error) {
	level, content := parseHeadingLine(ctx.getLine())
	if level < 1 || level > 6 {
		return nil, false, nil
	} else {
		return &parsedBlock{
			Type: "Heading",
			Attr: "level" + strconv.Itoa(level),
			Text: content,
			Range: ast.Range{
				Start: ast.Position{Line: ctx.getPos() + 1, Column: 1},
				End:   ast.Position{Line: ctx.getPos() + 1, Column: level + 1 + len(content)},
			},
		}, true, nil
	}
}

func parseHeadingLine(line string) (int, string) {
	idx := strings.Index(line, " ")
	if idx <= 0 {
		return 0, ""
	} else if strings.Count(line[:idx], "=") != len(line[:idx]) {
		return 0, ""
	} else {
		return strings.Count(line[:idx], "="), line[idx+1:]
	}
}
