package parser

import (
	"fmt"
	"strings"

	"orische/internal/ast"
)

func readInlineDirective(p *inlineParseState, start int) (ast.Inline, int, bool, error) {
	if !strings.HasPrefix(p.ctx.text[start:], ":[") {
		return nil, start, false, nil
	}
	return p.parseDirective(start)
}

func (p *inlineParseState) parseDirective(start int) (ast.Inline, int, bool, error) {
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

	// literalNext bounds fallback for invalid or unsupported headers.
	dirtype, attr, ok := parseInlineHeader(p.ctx.text[headerStart:headerEnd])
	if !ok {
		return nil, literalNext, false, nil
	}

	definition, ok := p.parser.spec.getInlineDirectiveDefinition(dirtype)
	if !ok {
		return nil, literalNext, false, nil
	}
	key := normalizeSyntaxType(dirtype)
	policy := definition.policy

	candidate := inlineCandidate{attribute: attr}
	var next int

	switch policy {
	case inlineContentNested:
		content, contentNext, closed, err := p.parseSeq(contentStart, true)
		if err != nil {
			return nil, start, false, err
		}
		if !closed {
			return nil, start, false, nil
		}
		candidate.nestedContent = content
		next = contentNext

	case inlineContentLiteral:
		if literalNext == start {
			return nil, start, false, nil
		}
		candidate.literalContent = p.ctx.text[contentStart : literalNext-1]
		next = literalNext

	default:
		return nil, start, false, fmt.Errorf(
			"inline directive %q has invalid content policy %d",
			key,
			policy,
		)
	}

	// Definitions only validate structurally closed candidates. An absent
	// validator accepts every attribute.
	if definition.validate != nil && !definition.validate(attr) {
		return nil, literalNext, false, nil
	}

	candidate.rng = p.ctx.rangeOf(start, next)
	if definition.build == nil {
		return nil, start, false, fmt.Errorf("build inline directive %q: definition has no builder", key)
	}
	node := definition.build(candidate)
	if node == nil {
		return nil, start, false, fmt.Errorf("build inline directive %q: definition returned a nil node", key)
	}

	return node, next, true, nil
}

func parseInlineHeader(header string) (dirtype string, attr string, ok bool) {
	dirtype, attr, _ = strings.Cut(header, ":")
	return dirtype, attr, dirtype != ""
}
