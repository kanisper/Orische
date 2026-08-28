package parser

import (
	"fmt"
	"strings"

	"orische/internal/ast"
	"orische/internal/parser/feature"
)

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
	policy := definition.ContentPolicy()

	candidate := feature.InlineDirectiveCandidate{Attribute: attr}
	var next int

	switch policy {
	case feature.InlineContentNested:
		content, contentNext, closed, err := p.parseSeq(contentStart, true)
		if err != nil {
			return nil, start, false, err
		}
		if !closed {
			return nil, start, false, nil
		}
		candidate.NestedContent = content
		next = contentNext

	case feature.InlineContentLiteral:
		if literalNext == start {
			return nil, start, false, nil
		}
		candidate.LiteralContent = p.ctx.text[contentStart : literalNext-1]
		next = literalNext

	default:
		return nil, start, false, fmt.Errorf(
			"inline directive %q has invalid content policy %d",
			key,
			policy,
		)
	}

	// Definitions only validate structurally closed candidates.
	accepted, err := definition.ValidateAttribute(attr)
	if err != nil {
		return nil, start, false, fmt.Errorf("validate inline directive %q: %w", key, err)
	}
	if !accepted {
		return nil, literalNext, false, nil
	}

	candidate.Range = p.ctx.rangeOf(start, next)
	node, err := definition.BuildInline(candidate)
	if err != nil {
		return nil, start, false, fmt.Errorf("build inline directive %q: %w", key, err)
	}
	if isNilRegistration(node) {
		return nil, start, false, fmt.Errorf("build inline directive %q: definition returned a nil node", key)
	}

	return node, next, true, nil
}

func parseInlineHeader(header string) (dirtype string, attr string, ok bool) {
	dirtype, attr, _ = strings.Cut(header, ":")
	return dirtype, attr, dirtype != ""
}
