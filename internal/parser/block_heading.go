package parser

import (
	"strconv"
	"strings"

	"orische/internal/ast"
)

type headingParser struct{}

func (*headingParser) parse(ctx *blockContext) (parsedBlockNode, bool, error) {
	level, content := parseHeadingLine(ctx.getLine())
	if level <= 0 {
		return &parsedBlock{}, false, nil
	} else {
		return &parsedBlock{
			Type: "Heading",
			Attr: "level" + strconv.Itoa(level),
			Text: content,
			Range: ast.Range{
				StartLine: ctx.getPos() + 1,
				EndLine:   ctx.getPos() + 1,
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
