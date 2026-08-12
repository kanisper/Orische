package parser

import (
	"strings"

	"orische/internal/ast"
)

func (p *inlineParser) parseDirective(start int) (ast.Inline, int, bool, error) {
	headerStart := start + 2

	headerEnd := strings.Index(p.ctx.text[headerStart:], "]{")
	if headerEnd < 0 {
		return nil, start, false, nil
	}
	headerEnd += headerStart

	contentStart := headerEnd + 2
	if contentStart >= len(p.ctx.text) {
		return nil, start, false, nil
	}

	literalEnd := strings.IndexByte(p.ctx.text[contentStart:], '}')
	literalNext := start
	if literalEnd >= 0 {
		literalNext = contentStart + literalEnd + 1
	}

	dirtype, attr, ok := parseInlineHeader(p.ctx.text[headerStart:headerEnd])
	if !ok {
		return nil, literalNext, false, nil
	}

	switch dirtype {
	case "em":
		content, next, closed, err := p.parseSeq(contentStart, true)
		if err != nil {
			return nil, start, false, nil
		}
		if !closed {
			return nil, start, false, nil
		}

		return &ast.Emphasis{
			Content: content,
			Range:   p.ctx.rangeOf(start, next),
		}, next, true, nil

	case "link":
		if attr == "" {
			return nil, literalNext, false, nil
		}

		content, next, closed, err := p.parseSeq(contentStart, true)
		if err != nil {
			return nil, start, false, err
		}
		if !closed {
			return nil, start, false, nil
		}

		return &ast.Link{
			URI:     attr,
			Content: content,
			Range:   p.ctx.rangeOf(start, next),
		}, next, true, nil

	case "code":
		contentEnd := strings.IndexByte(p.ctx.text[contentStart:], '}')
		if contentEnd < 0 {
			return nil, start, false, nil
		}
		contentEnd += contentStart

		return &ast.CodeSpan{
			Value: p.ctx.text[contentStart:contentEnd],
			Range: p.ctx.rangeOf(start, contentEnd+1),
		}, contentEnd + 1, true, nil

	default:
		return nil, literalNext, false, nil
	}
}

func parseInlineHeader(header string) (dirtype string, attr string, ok bool) {
	if sep := strings.IndexByte(header, ':'); sep >= 0 {
		dirtype = header[:sep]
		attr = header[sep+1:]
	} else {
		dirtype = header
		attr = ""
	}

	if dirtype == "" {
		return "", "", false
	}
	return dirtype, attr, true
}
