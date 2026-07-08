package parser

import (
	"strings"

	"medoc/internal/ast"
)

type listParser struct{}

func (*listParser) parse(ctx *blockContext) (parsedBlockNode, bool, error) {
	_, level, _ := parseListLine(ctx.getLine())
	if level <= 0 {
		return nil, false, nil
	}

	items := parseListBlocks(ctx)
	return items, true, nil
}

func parseListLine(line string) (bool, int, string) {
	idx := strings.Index(line, " ")
	if idx <= 0 {
		return false, 0, ""
	}
	if strings.Count(line[:idx], "*") == len(line[:idx]) {
		return false, len(line[:idx]), line[idx+1:]
	}
	if strings.Count(line[:idx], "#") == len(line[:idx]) {
		return true, len(line[:idx]), line[idx+1:]
	}
	return false, 0, ""
}

func parseListBlocks(ctx *blockContext) *parsedList {
	ordered, currentLevel, _ := parseListLine(ctx.getLine())

	startLine := ctx.getPos() + 1
	items := []parsedListItem{}
	item := parsedListItem{}

	for !ctx.isEOF() && strings.TrimSpace(ctx.getLine()) != "" {
		_, level, text := parseListLine(ctx.getLine())
		if level < currentLevel {
			break
		} else if level > currentLevel {
			item.Range = ast.Range{StartLine: ctx.getPos() + 1}
			item.Blocks = append(item.Blocks, parseListBlocks(ctx))
			item.Range.EndLine = ctx.getPos() + 1
		} else {
			item.Blocks = append(item.Blocks, &parsedBlock{
				Type:  "Paragraph",
				Attr:  "",
				Text:  text,
				Range: ast.Range{StartLine: ctx.getPos() + 1, EndLine: ctx.getPos() + 1},
			})
			item.Range = ast.Range{StartLine: ctx.getPos() + 1, EndLine: ctx.getPos() + 1}
		}
		items = append(items, item)
		item = parsedListItem{}
		ctx.advance(1)
	}

	ctx.advance(-1)
	return &parsedList{
		Ordered: ordered,
		Items:   items,
		Range:   ast.Range{StartLine: startLine, EndLine: ctx.getPos() + 1},
	}
}
