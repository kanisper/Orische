package parser

import "orische/internal/ast"

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
		nodes = p.appendText(nodes, textStart, end)
		textStart = end
	}

	for pos < len(p.ctx.text) {
		if stopAtClosingBrace && p.ctx.text[pos] == '}' {
			flushText(pos)
			return nodes, pos + 1, true, nil
		}

		matched := false
		for _, reader := range p.parser.spec.inlineReaders {
			node, next, ok, err := reader(p, pos)
			if err != nil {
				return nil, 0, false, err
			}

			if ok {
				flushText(pos)
				nodes = append(nodes, node)
				pos = next
				textStart = next
				matched = true
				break
			}
			if next > pos {
				pos = next
				matched = true
				break
			}
		}
		if matched {
			continue
		}

		pos++
	}

	flushText(pos)
	return nodes, pos, false, nil
}

func (p *inlineParseState) appendText(nodes []ast.Inline, start, end int) []ast.Inline {
	segmentStart := start
	for pos := start; pos < end; {
		next, newline := p.ctx.logicalNewlineEnd(pos)
		if !newline {
			pos++
			continue
		}

		if segmentStart < pos {
			nodes = append(nodes, p.textNode(segmentStart, pos))
		}
		pos = next
		segmentStart = next
	}
	if segmentStart < end {
		nodes = append(nodes, p.textNode(segmentStart, end))
	}
	return nodes
}

func (p *inlineParseState) textNode(start, end int) ast.Inline {
	return &ast.Text{
		Value: p.ctx.text[start:end],
		Range: p.ctx.rangeOf(start, end),
	}
}
