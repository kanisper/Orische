package parser

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"orische/internal/ast"
)

type listParser struct{}

type listLine struct {
	Ordered      bool
	RawLevel     int
	LogicalLevel int
	Text         string
	Line         int
}

func (*listParser) parse(ctx *blockContext) (parsedBlockNode, bool, error) {
	_, rawLevel, _ := parseListLine(ctx.getLine())
	if rawLevel <= 0 {
		return nil, false, nil
	}

	lines := collectListLines(ctx)
	if len(lines) == 0 {
		return nil, false, nil
	}

	index := 0

	list := buildParsedList(lines, &index, 1)

	if index != len(lines) {
		return nil, false, fmt.Errorf(
			"list parser: %d unconsumed list lines",
			len(lines)-index,
		)
	}

	return list, true, nil
}

func parseListLine(line string) (ordered bool, level int, text string) {
	separator_idx := strings.IndexByte(line, ' ')
	if separator_idx <= 0 {
		return false, 0, ""
	}

	markers := line[:separator_idx]

	for i := range markers {
		if markers[i] != '*' && markers[i] != '#' {
			return false, 0, ""
		}
	}

	return markers[0] == '#', len(markers), line[separator_idx+1:]
}

func collectListLines(ctx *blockContext) []listLine {
	var lines []listLine

	previousRawLevel := 0
	previousLogicalLevel := 0

	for !ctx.isEOF() {
		if strings.TrimSpace(ctx.getLine()) == "" {
			break
		}

		ordered, rawLevel, text := parseListLine(ctx.getLine())
		if rawLevel == 0 {
			break
		}

		logicalLevel := normalizeListLevel(previousRawLevel, previousLogicalLevel, rawLevel)

		lines = append(lines, listLine{
			Ordered:      ordered,
			RawLevel:     rawLevel,
			LogicalLevel: logicalLevel,
			Text:         text,
			Line:         ctx.getPos() + 1,
		})

		previousRawLevel = rawLevel
		previousLogicalLevel = logicalLevel

		ctx.advance(1)
	}

	ctx.advance(-1)
	return lines
}

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

			parent := &list.Items[len(list.Items)-1]
			parent.Blocks = append(parent.Blocks, nested)
			parent.Range.End = nested.Range.End

			list.Range.End = nested.Range.End

		default:
			endColumn := line.RawLevel + 1 + utf8.RuneCountInString(line.Text)

			list.Items = append(list.Items, parsedListItem{
				Blocks: []parsedBlockNode{
					&parsedBlock{
						Type: "Paragraph",
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
