package parser

import (
	"strings"
	"unicode/utf8"

	"orische/internal/ast"
)

const typeParagraph = "paragraph"

type blockContext struct {
	lines []string
	start int
}

func (i *blockContext) len() int {
	return len(i.lines) - i.start
}

type blockLine struct {
	number int
	text   string
}

func (i *blockContext) line(offset int) (blockLine, bool) {
	index := i.start + offset
	if offset < 0 || index < i.start || index >= len(i.lines) {
		return blockLine{}, false
	}
	return blockLine{number: index + 1, text: i.lines[index]}, true
}

func readBlockDirective(input *blockContext) (*blockDirectiveNode, int) {
	opener, ok := input.line(0)
	if !ok {
		return nil, 0
	}
	dirtype, attr, ok := parseBlockDirectiveHeader(opener.text)
	if !ok {
		return nil, 0
	}

	content := make([]string, 0)
	for offset := 1; offset < input.len(); offset++ {
		line, ok := input.line(offset)
		if !ok {
			break
		}
		if line.text == ":::" {
			return &blockDirectiveNode{
				dirtype:       dirtype,
				attribute:     attr,
				text:          strings.Join(content, "\n"),
				contentOrigin: ast.Position{Line: opener.number + 1, Column: 1},
				rng: ast.Range{
					Start: ast.Position{Line: opener.number, Column: 1},
					End:   ast.Position{Line: line.number, Column: 3},
				},
			}, offset + 1
		}
		content = append(content, line.text)
	}

	return nil, 0
}

func parseBlockDirectiveHeader(line string) (string, string, bool) {
	if !strings.HasPrefix(line, ":::[") || !strings.HasSuffix(line, "]") {
		return "", "", false
	}

	dirtype, attr, _ := strings.Cut(line[4:len(line)-1], ":")
	return dirtype, attr, dirtype != ""
}

func readParagraph(input *blockContext) (*paragraphNode, int) {
	first, ok := input.line(0)
	if !ok || strings.TrimSpace(first.text) == "" {
		return nil, 0
	}

	content := make([]string, 0, input.len())
	last := first
	for offset := 0; offset < input.len(); offset++ {
		line, ok := input.line(offset)
		if !ok || strings.TrimSpace(line.text) == "" {
			break
		}
		content = append(content, line.text)
		last = line
	}

	return &paragraphNode{
		text:          strings.Join(content, "\n"),
		contentOrigin: ast.Position{Line: first.number, Column: 1},
		rng: ast.Range{
			Start: ast.Position{Line: first.number, Column: 1},
			End:   ast.Position{Line: last.number, Column: utf8.RuneCountInString(last.text)},
		},
	}, len(content)
}

func (p *Parser) buildParagraph(block *paragraphNode) (ast.Block, error) {
	content, err := p.parseInlines(block.text, block.contentOrigin)
	if err != nil {
		return nil, err
	}

	return &ast.Paragraph{Content: content, Range: block.rng}, nil
}

func buildParagraphDirective(p *Parser, block *blockDirectiveNode) (ast.Block, error) {
	return p.buildParagraph(&paragraphNode{
		text:          block.text,
		contentOrigin: block.contentOrigin,
		rng:           block.rng,
	})
}
