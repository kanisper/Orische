package parser

import (
	"fmt"
	"strings"

	"orische/internal/ast"
)

type inlineContentPolicy uint8

const (
	inlineContentNested inlineContentPolicy = iota
	inlineContentLiteral
)

type inlineDirectiveCandidate struct {
	Attribute      string
	NestedContent  []ast.Inline
	LiteralContent string
	Range          ast.Range
}

// inlineDirectiveDefinition owns syntax-specific validation, content policy,
// and AST construction. The common parser owns delimiters, fallback, and range
// calculation.
type inlineDirectiveDefinition interface {
	contentPolicy() inlineContentPolicy
	validateAttribute(attribute string) (bool, error)
	buildInline(candidate inlineDirectiveCandidate) (ast.Inline, error)
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
	key := normalizeDirectiveType(dirtype)
	policy := definition.contentPolicy()

	candidate := inlineDirectiveCandidate{Attribute: attr}
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
		candidate.NestedContent = content
		next = contentNext

	case inlineContentLiteral:
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
	accepted, err := definition.validateAttribute(attr)
	if err != nil {
		return nil, start, false, fmt.Errorf("validate inline directive %q: %w", key, err)
	}
	if !accepted {
		return nil, literalNext, false, nil
	}

	candidate.Range = p.ctx.rangeOf(start, next)
	node, err := definition.buildInline(candidate)
	if err != nil {
		return nil, start, false, fmt.Errorf("build inline directive %q: %w", key, err)
	}
	if node == nil {
		return nil, start, false, fmt.Errorf("build inline directive %q: definition returned a nil node", key)
	}

	return node, next, true, nil
}

func parseInlineHeader(header string) (dirtype string, attr string, ok bool) {
	dirtype, attr, _ = strings.Cut(header, ":")
	return dirtype, attr, dirtype != ""
}
