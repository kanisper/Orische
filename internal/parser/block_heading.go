package parser

import (
	"strconv"
	"strings"

	"medoc/internal/ast"
)

type headingParser struct{}

func (*headingParser) parse(ctx *blockContext) (parsedBlockNode, bool, error) {
	level, content := isHeadingLine(ctx.getLine())
	if level <= 0 {
		return &parsedBlock{}, false, nil
	} else {
		return &parsedBlock{
			Type: "Heading",
			Attr: "level" + strconv.Itoa(level),
			Text: content,
			Range: ast.Range{
				StartLine: ctx.getPos(),
				EndLine:   ctx.getPos(),
			},
		}, true, nil
	}
}

func isHeadingLine(line string) (int, string) {
	idx := strings.Index(line, " ")
	if idx <= 0 {
		return 0, ""
	} else if strings.Count(line[:idx], "=") != len(line[:idx]) {
		return 0, ""
	} else {
		return strings.Count(line[:idx], "="), line[idx+1:]
	}
}
