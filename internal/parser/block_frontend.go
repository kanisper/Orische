package parser

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"orische/internal/ast"
	"orische/internal/parser/feature"
)

type blockInput struct {
	lines []string
	start int
}

func (i *blockInput) Len() int {
	return len(i.lines) - i.start
}

func (i *blockInput) Line(offset int) (feature.BlockLine, bool) {
	index := i.start + offset
	if offset < 0 || index < i.start || index >= len(i.lines) {
		return feature.BlockLine{}, false
	}
	return feature.BlockLine{Number: index + 1, Text: i.lines[index]}, true
}

type blockDirectiveReader struct{}

func (*blockDirectiveReader) ReadBlock(input feature.BlockInput) (feature.BlockReadResult, error) {
	opener, ok := input.Line(0)
	if !ok {
		return feature.BlockReadResult{}, nil
	}
	dirtype, attr, ok := parseBlockDirective(opener.Text)
	if !ok {
		return feature.BlockReadResult{}, nil
	}

	content := make([]string, 0)
	for offset := 1; offset < input.Len(); offset++ {
		line, ok := input.Line(offset)
		if !ok {
			break
		}
		if line.Text == ":::" {
			return feature.BlockReadResult{
				Matched:  true,
				Consumed: offset + 1,
				Node: &feature.TextBlock{
					Type:          dirtype,
					Attr:          attr,
					Text:          strings.Join(content, "\n"),
					ContentOrigin: ast.Position{Line: opener.Number + 1, Column: 1},
					Range: ast.Range{
						Start: ast.Position{Line: opener.Number, Column: 1},
						End:   ast.Position{Line: line.Number, Column: 3},
					},
				},
			}, nil
		}
		content = append(content, line.Text)
	}

	return feature.BlockReadResult{}, nil
}

func parseBlockDirective(line string) (string, string, bool) {
	if !strings.HasPrefix(line, ":::[") || !strings.HasSuffix(line, "]") {
		return "", "", false
	}

	dirtype, attr, _ := strings.Cut(line[4:len(line)-1], ":")
	return dirtype, attr, dirtype != ""
}

type paragraphReader struct{}

func (*paragraphReader) ReadBlock(input feature.BlockInput) (feature.BlockReadResult, error) {
	first, ok := input.Line(0)
	if !ok || strings.TrimSpace(first.Text) == "" {
		return feature.BlockReadResult{}, nil
	}

	content := make([]string, 0, input.Len())
	last := first
	for offset := 0; offset < input.Len(); offset++ {
		line, ok := input.Line(offset)
		if !ok || strings.TrimSpace(line.Text) == "" {
			break
		}
		content = append(content, line.Text)
		last = line
	}

	return feature.BlockReadResult{
		Matched:  true,
		Consumed: len(content),
		Node: &feature.TextBlock{
			Type:          feature.ParagraphBlockType,
			Text:          strings.Join(content, "\n"),
			ContentOrigin: ast.Position{Line: first.Number, Column: 1},
			Range: ast.Range{
				Start: ast.Position{Line: first.Number, Column: 1},
				End:   ast.Position{Line: last.Number, Column: utf8.RuneCountInString(last.Text)},
			},
		},
	}, nil
}

type paragraphDefinition struct{}

func (*paragraphDefinition) BlockType() string {
	return feature.ParagraphBlockType
}

func (*paragraphDefinition) BuildBlock(ctx feature.BuildContext, node feature.BlockNode) (ast.Block, error) {
	block, ok := node.(*feature.TextBlock)
	if !ok {
		return nil, fmt.Errorf("expected *feature.TextBlock, got %T", node)
	}

	content, err := ctx.ParseInlines(block.Text, block.ContentOrigin)
	if err != nil {
		return nil, err
	}

	return &ast.Paragraph{Content: content, Range: block.Range}, nil
}
