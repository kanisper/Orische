package parser

import "orische/internal/ast"

func readInlineEscape(p *inlineParseState, start int) (ast.Inline, int, bool, error) {
	if p.ctx.text[start] != '\\' || start+1 >= len(p.ctx.text) {
		return nil, start, false, nil
	}

	escaped := p.ctx.text[start+1]
	if !isASCIIPunctuation(escaped) {
		return nil, start, false, nil
	}

	next := start + 2
	return &ast.Text{
		Value: string(escaped),
		Range: p.ctx.rangeOf(start, next),
	}, next, true, nil
}

func isASCIIPunctuation(char byte) bool {
	return char >= '!' && char <= '/' ||
		char >= ':' && char <= '@' ||
		char >= '[' && char <= '`' ||
		char >= '{' && char <= '~'
}
