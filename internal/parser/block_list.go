package parser

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"orische/internal/ast"
)

const blockBuilderKeyList = "list"

type listReader struct{}

func (*listReader) builderKey() string {
	return blockBuilderKeyList
}

type listLine struct {
	Ordered      bool
	RawLevel     int
	LogicalLevel int
	Text         string
	Line         int
}

func (*listReader) read(ctx *blockContext) (parsedBlockNode, bool, error) {
	_, rawLevel, _ := parseListLine(ctx.line())
	if rawLevel <= 0 {
		return nil, false, nil
	}

	lines := collectListLines(ctx)
	index := 0
	list := buildParsedList(lines, &index, 1)

	if index != len(lines) {
		return nil, false, fmt.Errorf(
			"list reader: %d unconsumed list lines",
			len(lines)-index,
		)
	}

	return list, true, nil
}

func parseListLine(line string) (ordered bool, level int, text string) {
	separator := strings.IndexByte(line, ' ')
	if separator <= 0 {
		return false, 0, ""
	}

	markers := line[:separator]

	for i := range markers {
		if markers[i] != '*' && markers[i] != '#' {
			return false, 0, ""
		}
	}

	return markers[0] == '#', len(markers), line[separator+1:]
}

func collectListLines(ctx *blockContext) []listLine {
	var lines []listLine

	previousRawLevel := 0
	previousLogicalLevel := 0

	for !ctx.isEOF() {
		ordered, rawLevel, text := parseListLine(ctx.line())
		if rawLevel == 0 {
			break
		}

		logicalLevel := normalizeListLevel(previousRawLevel, previousLogicalLevel, rawLevel)

		lines = append(lines, listLine{
			Ordered:      ordered,
			RawLevel:     rawLevel,
			LogicalLevel: logicalLevel,
			Text:         text,
			Line:         ctx.pos + 1,
		})

		previousRawLevel = rawLevel
		previousLogicalLevel = logicalLevel

		ctx.pos++
	}

	// The document loop advances once after a successful reader.
	ctx.pos--
	return lines
}

// normalizeListLevel maps marker-run changes to logical nesting. Any increase,
// regardless of its raw size, introduces exactly one logical level.
func normalizeListLevel(previousRaw int, previousLogical int, currentRaw int) int {
	if previousRaw == 0 {
		return 1
	}

	switch {
	case previousRaw < currentRaw:
		return previousLogical + 1

	case previousRaw > currentRaw:
		level := previousLogical - (previousRaw - currentRaw)
		if level < 1 {
			return 1
		}
		return level

	default:
		return previousLogical
	}
}

func buildParsedList(lines []listLine, index *int, level int) *parsedList {
	first := lines[*index]

	list := &parsedList{
		Ordered: first.Ordered,
		Items:   make([]parsedListItem, 0),
		Range: ast.Range{
			Start: ast.Position{Line: first.Line, Column: 1},
			End:   ast.Position{Line: first.Line},
		},
	}

	for *index < len(lines) {
		line := lines[*index]

		switch {
		case line.LogicalLevel < level:
			goto done

		case line.LogicalLevel > level:
			nested := buildParsedList(lines, index, level+1)

			// A deeper run always belongs to the immediately preceding item.
			parent := &list.Items[len(list.Items)-1]
			parent.Blocks = append(parent.Blocks, nested)
			parent.Range.End = nested.Range.End

			list.Range.End = nested.Range.End

		default:
			endColumn := line.RawLevel + 1 + utf8.RuneCountInString(line.Text)

			list.Items = append(list.Items, parsedListItem{
				Blocks: []parsedBlockNode{
					&parsedBlock{
						Type: blockBuilderKeyParagraph,
						Attr: "",
						Text: line.Text,
						Range: ast.Range{
							Start: ast.Position{Line: line.Line, Column: line.RawLevel + 2},
							End:   ast.Position{Line: line.Line, Column: endColumn},
						},
					},
				},
				Range: ast.Range{
					Start: ast.Position{Line: line.Line, Column: 1},
					End:   ast.Position{Line: line.Line, Column: endColumn},
				},
			})

			list.Range.End.Line = line.Line
			list.Range.End.Column = endColumn
			(*index)++
		}
	}

done:
	return list
}

type listBuilder struct{}

func (*listBuilder) build(parser *Parser, node parsedBlockNode) (ast.Block, error) {
	list, ok := node.(*parsedList)
	if !ok {
		return nil, fmt.Errorf("expected *parsedList, got %T", node)
	}

	result := &ast.List{
		Ordered: list.Ordered,
		Items:   make([]*ast.ListItem, 0, len(list.Items)),
		Range:   list.Range,
	}

	for _, parsedItem := range list.Items {
		blocks := make([]ast.Block, 0, len(parsedItem.Blocks))
		for _, node := range parsedItem.Blocks {
			block, err := parser.buildBlock(node)
			if err != nil {
				return nil, err
			}
			blocks = append(blocks, block)
		}

		result.Items = append(result.Items, &ast.ListItem{
			Blocks: blocks,
			Range:  parsedItem.Range,
		})
	}

	return result, nil
}
