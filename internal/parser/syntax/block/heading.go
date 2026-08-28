package block

import (
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"

	"orische/internal/ast"
	"orische/internal/parser/feature"
)

const typeHeading = "heading"

type headingDefinition struct{}

func (*headingDefinition) BlockType() string {
	return typeHeading
}

func (*headingDefinition) ReadBlock(input feature.BlockInput) (feature.BlockReadResult, error) {
	line, ok := input.Line(0)
	if !ok {
		return feature.BlockReadResult{}, nil
	}

	level, content := parseHeadingLine(line.Text)
	if level < 1 || level > 6 {
		return feature.BlockReadResult{}, nil
	}

	return feature.BlockReadResult{
		Matched:  true,
		Consumed: 1,
		Node: &feature.TextBlock{
			Type:          typeHeading,
			Attr:          "level" + strconv.Itoa(level),
			Text:          content,
			ContentOrigin: ast.Position{Line: line.Number, Column: level + 2},
			Range: ast.Range{
				Start: ast.Position{Line: line.Number, Column: 1},
				End:   ast.Position{Line: line.Number, Column: level + 1 + utf8.RuneCountInString(content)},
			},
		},
	}, nil
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

func (*headingDefinition) BuildBlock(ctx feature.BuildContext, node feature.BlockNode) (ast.Block, error) {
	block, ok := node.(*feature.TextBlock)
	if !ok {
		return nil, fmt.Errorf("expected *feature.TextBlock, got %T", node)
	}

	const levelPrefix = "level"
	if !strings.HasPrefix(block.Attr, levelPrefix) || len(block.Attr) == len(levelPrefix) {
		return nil, fmt.Errorf("invalid heading attribute %q", block.Attr)
	}
	level, err := strconv.Atoi(block.Attr[len(levelPrefix):])
	if err != nil {
		return nil, fmt.Errorf("invalid heading attribute %q: %w", block.Attr, err)
	}
	if level < 1 || level > 6 {
		return nil, fmt.Errorf("invalid heading attribute %q", block.Attr)
	}

	content, err := ctx.ParseInlines(block.Text, block.ContentOrigin)
	if err != nil {
		return nil, err
	}

	return &ast.Heading{
		Level:   level,
		Content: content,
		Range:   block.Range,
	}, nil
}
