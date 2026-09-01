package parser

import (
	"strings"
	"unicode/utf8"

	"orische/internal/ast"
)

const typeList = "list"

type listLine struct {
	ordered      bool
	rawLevel     int
	logicalLevel int
	text         string
	line         int
}

type listNode struct {
	ordered bool
	items   []listItemNode
	rng     ast.Range
}

func (n *listNode) blockRange() ast.Range {
	return n.rng
}

type listItemNode struct {
	blocks []parsedBlock
	rng    ast.Range
}

func readList(input *blockContext) (parsedBlock, int) {
	first, ok := input.line(0)
	if !ok {
		return nil, 0
	}
	_, rawLevel, _ := parseListLine(first.text)
	if rawLevel <= 0 {
		return nil, 0
	}

	lines := collectListLines(input)
	index := 0
	list := buildListNode(lines, &index, 1)
	return list, len(lines)
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

func collectListLines(input *blockContext) []listLine {
	var lines []listLine
	previousRawLevel := 0
	previousLogicalLevel := 0

	for offset := 0; offset < input.len(); offset++ {
		line, ok := input.line(offset)
		if !ok {
			break
		}
		ordered, rawLevel, text := parseListLine(line.text)
		if rawLevel == 0 {
			break
		}

		logicalLevel := normalizeListLevel(previousRawLevel, previousLogicalLevel, rawLevel)
		lines = append(lines, listLine{
			ordered:      ordered,
			rawLevel:     rawLevel,
			logicalLevel: logicalLevel,
			text:         text,
			line:         line.number,
		})
		previousRawLevel = rawLevel
		previousLogicalLevel = logicalLevel
	}

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

func buildListNode(lines []listLine, index *int, level int) *listNode {
	first := lines[*index]
	list := &listNode{
		ordered: first.ordered,
		items:   make([]listItemNode, 0),
		rng: ast.Range{
			Start: ast.Position{Line: first.line, Column: 1},
			End:   ast.Position{Line: first.line},
		},
	}

	for *index < len(lines) {
		line := lines[*index]
		switch {
		case line.logicalLevel < level:
			return list

		case line.logicalLevel > level:
			nested := buildListNode(lines, index, level+1)
			parent := &list.items[len(list.items)-1]
			parent.blocks = append(parent.blocks, nested)
			parent.rng.End = nested.rng.End
			list.rng.End = nested.rng.End

		default:
			endColumn := line.rawLevel + 1 + utf8.RuneCountInString(line.text)
			list.items = append(list.items, listItemNode{
				blocks: []parsedBlock{
					&paragraphNode{
						text:          line.text,
						contentOrigin: ast.Position{Line: line.line, Column: line.rawLevel + 2},
						rng: ast.Range{
							Start: ast.Position{Line: line.line, Column: line.rawLevel + 2},
							End:   ast.Position{Line: line.line, Column: endColumn},
						},
					},
				},
				rng: ast.Range{
					Start: ast.Position{Line: line.line, Column: 1},
					End:   ast.Position{Line: line.line, Column: endColumn},
				},
			})
			list.rng.End = ast.Position{Line: line.line, Column: endColumn}
			*index++
		}
	}

	return list
}

func (p *Parser) buildList(list *listNode) (ast.Block, error) {
	result := &ast.List{
		Ordered: list.ordered,
		Items:   make([]*ast.ListItem, 0, len(list.items)),
		Range:   list.rng,
	}

	for _, parsedItem := range list.items {
		blocks := make([]ast.Block, 0, len(parsedItem.blocks))
		for _, child := range parsedItem.blocks {
			built, err := p.buildBlock(child)
			if err != nil {
				return nil, err
			}
			blocks = append(blocks, built)
		}
		result.Items = append(result.Items, &ast.ListItem{Blocks: blocks, Range: parsedItem.rng})
	}

	return result, nil
}
