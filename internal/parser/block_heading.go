package parser

import (
	"strings"
	"unicode/utf8"

	"orische/internal/ast"
)

const typeHeading = "heading"

type headingNode struct {
	level         int
	text          string
	contentOrigin ast.Position
	rng           ast.Range
}

func (n *headingNode) blockRange() ast.Range {
	return n.rng
}

func readHeading(input *blockContext) (parsedBlock, int) {
	line, ok := input.line(0)
	if !ok {
		return nil, 0
	}

	level, content := parseHeadingLine(line.text)
	if level < 1 || level > 6 {
		return nil, 0
	}

	return &headingNode{
		level:         level,
		text:          content,
		contentOrigin: ast.Position{Line: line.number, Column: level + 2},
		rng: ast.Range{
			Start: ast.Position{Line: line.number, Column: 1},
			End:   ast.Position{Line: line.number, Column: level + 1 + utf8.RuneCountInString(content)},
		},
	}, 1
}

func parseHeadingLine(line string) (int, string) {
	idx := strings.IndexByte(line, ' ')
	if idx <= 0 {
		return 0, ""
	}
	if strings.Trim(line[:idx], "=") != "" {
		return 0, ""
	}
	return idx, line[idx+1:]
}

func (p *Parser) buildHeading(block *headingNode) (ast.Block, error) {
	content, err := p.parseInlines(block.text, block.contentOrigin)
	if err != nil {
		return nil, err
	}

	return &ast.Heading{
		Level:   block.level,
		Content: content,
		Range:   block.rng,
	}, nil
}
