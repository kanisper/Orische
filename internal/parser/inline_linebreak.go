package parser

import (
	"strings"

	"orische/internal/ast"
)

func readLineBreak(p *inlineParseState, start int) (ast.Inline, int, bool, error) {
	const marker = " +"
	if !strings.HasPrefix(p.ctx.text[start:], marker) {
		return nil, start, false, nil
	}

	markerEnd := start + len(marker)
	next, ok := p.ctx.logicalNewlineEnd(markerEnd)
	if !ok {
		return nil, start, false, nil
	}

	return &ast.LineBreak{
		Range: p.ctx.rangeOf(start, markerEnd),
	}, next, true, nil
}
