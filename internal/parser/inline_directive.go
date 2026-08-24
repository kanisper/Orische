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

	dirtype, attr, ok := parseInlineHeader(p.ctx.text[headerStart:headerEnd])
	if !ok {
		return nil, literalNext, false, nil
	}

	definition, ok := p.parser.spec.getInlineDirectiveDefinition(dirtype)
	if !ok {
		return nil, literalNext, false, nil
	}

	candidate := inlineDirectiveCandidate{Attribute: attr}
	var next int

	switch definition.contentPolicy() {
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
			normalizeDirectiveType(dirtype),
			definition.contentPolicy(),
		)
	}

	accepted, err := definition.validateAttribute(attr)
	if err != nil {
		return nil, start, false, fmt.Errorf("validate inline directive %q: %w", normalizeDirectiveType(dirtype), err)
	}
	if !accepted {
		return nil, literalNext, false, nil
	}

	candidate.Range = p.ctx.rangeOf(start, next)
	node, err := definition.buildInline(candidate)
	if err != nil {
		return nil, start, false, fmt.Errorf("build inline directive %q: %w", normalizeDirectiveType(dirtype), err)
	}
	if node == nil {
		return nil, start, false, fmt.Errorf("build inline directive %q: definition returned a nil node", normalizeDirectiveType(dirtype))
	}

	return node, next, true, nil
}

type emphasisInlineDefinition struct{}

func (*emphasisInlineDefinition) contentPolicy() inlineContentPolicy {
	return inlineContentNested
}

func (*emphasisInlineDefinition) validateAttribute(string) (bool, error) {
	return true, nil
}

func (*emphasisInlineDefinition) buildInline(candidate inlineDirectiveCandidate) (ast.Inline, error) {
	return &ast.Emphasis{
		Content: candidate.NestedContent,
		Range:   candidate.Range,
	}, nil
}

type linkInlineDefinition struct{}

func (*linkInlineDefinition) contentPolicy() inlineContentPolicy {
	return inlineContentNested
}

func (*linkInlineDefinition) validateAttribute(attribute string) (bool, error) {
	return attribute != "", nil
}

func (*linkInlineDefinition) buildInline(candidate inlineDirectiveCandidate) (ast.Inline, error) {
	return &ast.Link{
		URI:     candidate.Attribute,
		Content: candidate.NestedContent,
		Range:   candidate.Range,
	}, nil
}

type codeInlineDefinition struct{}

func (*codeInlineDefinition) contentPolicy() inlineContentPolicy {
	return inlineContentLiteral
}

func (*codeInlineDefinition) validateAttribute(string) (bool, error) {
	return true, nil
}

func (*codeInlineDefinition) buildInline(candidate inlineDirectiveCandidate) (ast.Inline, error) {
	return &ast.CodeSpan{
		Value: candidate.LiteralContent,
		Range: candidate.Range,
	}, nil
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
