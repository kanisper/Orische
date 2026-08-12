package parser

import (
	"orische/internal/ast"
)

type inlineParser struct {
	ctx *inlineContext
}

func parseInlines(text string, origin ast.Position) ([]ast.Inline, error) {
	p := &inlineParser{
		ctx: newInlineContext(text, origin),
	}

	nodes, _, _, err := p.parseSeq(0, false)
	if err != nil {
		return nil, err
	}

	return nodes, nil
}

// return []ast.Inline  the parsed inline nodes
// return int           the position after the parsed inline nodes
// return bool          whether the parsing stopped at a closing brace
// return error
func (p *inlineParser) parseSeq(
	start int,
	stopAtClosingBrace bool,
) ([]ast.Inline, int, bool, error) {
	var nodes []ast.Inline

	pos := start
	textStart := start

	flushText := func(end int) {
		if textStart == end {
			return
		}

		nodes = append(nodes, &ast.Text{
			Value: p.ctx.text[textStart:end],
			Range: p.ctx.rangeOf(textStart, end),
		})

		textStart = end
	}

	for pos < len(p.ctx.text) {
		if stopAtClosingBrace && p.ctx.text[pos] == '}' {
			flushText(pos)
			return nodes, pos + 1, true, nil
		}

		if p.ctx.hasInlinePrefix(pos) {
			node, next, ok, err := p.parseDirective(pos)
			if err != nil {
				return nil, 0, false, err
			}

			if ok {
				flushText(pos)
				nodes = append(nodes, node)

				pos = next
				textStart = next
				continue
			}
			if next > pos {
				pos = next
				continue
			}
		}

		pos++
	}

	flushText(pos)
	return nodes, pos, false, nil
}
