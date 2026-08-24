package parser

import (
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"

	"orische/internal/ast"
)

const blockBuilderKeyHeading = "heading"

type headingReader struct{}

func (*headingReader) builderKey() string {
	return blockBuilderKeyHeading
}

func (*headingReader) read(ctx *blockContext) (parsedBlockNode, bool, error) {
	level, content := parseHeadingLine(ctx.line())
	if level < 1 || level > 6 {
		return nil, false, nil
	}

	return &parsedBlock{
		Type: blockBuilderKeyHeading,
		Attr: "level" + strconv.Itoa(level),
		Text: content,
		Range: ast.Range{
			Start: ast.Position{Line: ctx.pos + 1, Column: 1},
			End:   ast.Position{Line: ctx.pos + 1, Column: level + 1 + utf8.RuneCountInString(content)},
		},
	}, true, nil
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

type headingBuilder struct{}

func (*headingBuilder) build(parser *Parser, node parsedBlockNode) (ast.Block, error) {
	block, ok := node.(*parsedBlock)
	if !ok {
		return nil, fmt.Errorf("expected *parsedBlock, got %T", node)
	}

	level, err := strconv.Atoi(block.Attr[len(block.Attr)-1:])
	if err != nil {
		return nil, fmt.Errorf("invalid attribute %s: %w", block.Attr, err)
	}

	origin := ast.Position{
		Line:   block.Range.Start.Line,
		Column: block.Range.Start.Column + level + 1,
	}
	content, err := parser.parseInlines(block.Text, origin)
	if err != nil {
		return nil, err
	}

	return &ast.Heading{
		Level:   level,
		Content: content,
		Range:   block.Range,
	}, nil
}
