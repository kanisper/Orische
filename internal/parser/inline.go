package parser

import (
	"strings"

	"orische/internal/ast"
)

func parseInlines(text string) ([]ast.Inline, error) {
	nodes, _, _, err := parseInlineSeq(text, 0, false)
	if err != nil {
		return nil, err
	}

	return nodes, nil
}

// return []ast.Inline  the parsed inline nodes
// return int           the position after the parsed inline nodes
// return bool          whether the parsing stopped at a closing brace
// return error
func parseInlineSeq(text string, start int, stopAtClosingBrace bool) ([]ast.Inline, int, bool, error) {
	var nodes []ast.Inline
	var ctx strings.Builder

	flushCtx := func() {
		if ctx.Len() == 0 {
			return
		}
		nodes = append(nodes, &ast.Text{Value: ctx.String()})
		ctx.Reset()
	}

	pos := start

	for pos < len(text) {
		if stopAtClosingBrace && text[pos] == '}' {
			flushCtx()
			return nodes, pos + 1, true, nil
		}

		if hasInlinePrefix(text, pos) {
			node, next, ok, err := parseInline(text, pos)
			if err != nil {
				return nil, 0, false, err
			}
			if ok {
				flushCtx()
				nodes = append(nodes, node)
				pos = next
				continue
			}
		}

		ctx.WriteByte(text[pos])
		pos++
	}

	flushCtx()
	return nodes, pos, false, nil
}

func hasInlinePrefix(text string, pos int) bool {
	return strings.HasPrefix(text[pos:], ":[")
}

func parseInline(text string, start int) (ast.Inline, int, bool, error) {
	headerStart := start + 2

	headerEnd := strings.Index(text[headerStart:], "]{")
	if headerEnd < 0 {
		return nil, start, false, nil
	}
	headerEnd += headerStart

	dirtype, attr, ok := parseInlineHeader(text[headerStart:headerEnd])
	if !ok {
		return nil, start, false, nil
	}

	contentStart := headerEnd + 2
	if contentStart+1 >= len(text) {
		return nil, start, false, nil
	}

	switch dirtype {
	case "em":
		content, next, closed, err := parseInlineSeq(text, contentStart, true)
		if err != nil {
			return nil, start, false, nil
		}
		if !closed {
			return nil, start, false, nil
		}

		return &ast.Emphasis{
			Content: content,
		}, next, true, nil

	case "link":
		if attr == "" {
			return nil, start, false, nil
		}

		content, next, closed, err := parseInlineSeq(text, contentStart, true)
		if err != nil {
			return nil, start, false, err
		}
		if !closed {
			return nil, start, false, nil
		}

		return &ast.Link{
			URI:     attr,
			Content: content,
		}, next, true, nil

	case "code":
		contentEnd := strings.IndexByte(text[contentStart:], '}')
		if contentEnd < 0 {
			return nil, start, false, nil
		}
		contentEnd += contentStart

		return &ast.CodeSpan{
			Value: text[contentStart:contentEnd],
		}, contentEnd + 1, true, nil

	default:
		return nil, start, false, nil
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
