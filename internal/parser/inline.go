package parser

import (
	"strings"

	"orische/internal/ast"
)

type inlineParseState struct {
	parser *Parser
	ctx    *inlineContext
}

func (p *Parser) parseInlines(text string, origin ast.Position) ([]ast.Inline, error) {
	state := &inlineParseState{
		parser: p,
		ctx:    newInlineContext(text, origin),
	}

	nodes, _, _, err := state.parseSeq(0, false)
	return nodes, err
}

// parseSeq parses from start until EOF or, when requested, the first closing
// brace. next is the byte offset after the consumed input; closed reports
// whether a closing brace ended the sequence.
func (p *inlineParseState) parseSeq(
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

		if strings.HasPrefix(p.ctx.text[pos:], ":[") {
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
