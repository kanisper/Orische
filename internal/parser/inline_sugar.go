package parser

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"orische/internal/ast"
)

func readStrongSugar(p *inlineParseState, start int) (ast.Inline, int, bool, error) {
	return p.readDelimitedSugar(start, "*", "strong", true)
}

func readEmphasisSugar(p *inlineParseState, start int) (ast.Inline, int, bool, error) {
	return p.readDelimitedSugar(start, "_", "em", true)
}

func readBoldSugar(p *inlineParseState, start int) (ast.Inline, int, bool, error) {
	return p.readDelimitedSugar(start, "**", "bold", true)
}

func readItalicSugar(p *inlineParseState, start int) (ast.Inline, int, bool, error) {
	return p.readDelimitedSugar(start, "__", "italic", true)
}

func readDeletedSugar(p *inlineParseState, start int) (ast.Inline, int, bool, error) {
	return p.readDelimitedSugar(start, "--", "del", true)
}

func readOutdatedSugar(p *inlineParseState, start int) (ast.Inline, int, bool, error) {
	return p.readDelimitedSugar(start, "~", "outdated", true)
}

func readCodeSpanSugar(p *inlineParseState, start int) (ast.Inline, int, bool, error) {
	return p.readDelimitedSugar(start, "`", "code", false)
}

func (p *inlineParseState) readDelimitedSugar(
	start int,
	marker string,
	typ string,
	escapeClosingMarker bool,
) (ast.Inline, int, bool, error) {
	if !hasExactDelimiterRun(p.ctx.text, start, marker) ||
		!p.hasOpeningSugarBoundary(start) ||
		isEscapedDelimiter(p.ctx.text, start) {
		return nil, start, false, nil
	}

	lineEnd := p.logicalLineEnd(start)
	contentStart := start + len(marker)
	closingStart := -1
	for search := contentStart; search < lineEnd; {
		relative := strings.Index(p.ctx.text[search:lineEnd], marker)
		if relative < 0 {
			break
		}
		candidate := search + relative
		candidateEnd := candidate + len(marker)
		if hasExactDelimiterRun(p.ctx.text, candidate, marker) &&
			(!escapeClosingMarker || !isEscapedDelimiter(p.ctx.text, candidate)) &&
			p.hasClosingSugarBoundary(candidateEnd) {
			closingStart = candidate
			break
		}
		search = candidate + 1
	}

	if closingStart < 0 {
		return nil, lineEnd, false, nil
	}

	next := closingStart + len(marker)
	if contentStart == closingStart ||
		p.ctx.text[contentStart] == ' ' ||
		p.ctx.text[closingStart-1] == ' ' {
		return nil, next, false, nil
	}

	definition, ok := p.parser.spec.getInlineDirectiveDefinition(typ)
	if !ok {
		return nil, start, false, fmt.Errorf("inline sugar %q has no definition", typ)
	}
	candidate := inlineCandidate{rng: p.ctx.rangeOf(start, next)}
	switch definition.policy {
	case inlineContentNested:
		content, err := p.parser.parseInlines(
			p.ctx.text[contentStart:closingStart],
			p.ctx.positionAt(contentStart),
		)
		if err != nil {
			return nil, start, false, err
		}
		candidate.nestedContent = content
	case inlineContentLiteral:
		candidate.literalContent = p.ctx.text[contentStart:closingStart]
	default:
		return nil, start, false, fmt.Errorf(
			"inline sugar %q has invalid content policy %d",
			typ,
			definition.policy,
		)
	}

	node, accepted, err := buildSugarCandidate(typ, definition, candidate)
	if err != nil {
		return nil, start, false, err
	}
	if !accepted {
		return nil, next, false, nil
	}
	return node, next, true, nil
}

func readLinkSugar(p *inlineParseState, start int) (ast.Inline, int, bool, error) {
	const opening = "["
	if !hasExactDelimiterRun(p.ctx.text, start, opening) ||
		!p.hasOpeningSugarBoundary(start) ||
		isEscapedDelimiter(p.ctx.text, start) {
		return nil, start, false, nil
	}

	lineEnd := p.logicalLineEnd(start)
	labelStart := start + len(opening)
	labelEnd := findUnescapedSequence(p.ctx.text, labelStart, lineEnd, "](")
	if labelEnd < 0 {
		return nil, start, false, nil
	}

	uriStart := labelEnd + len("](")
	uriEnd := findUnescapedSequence(p.ctx.text, uriStart, lineEnd, ")")
	if uriEnd < 0 {
		return nil, lineEnd, false, nil
	}
	next := uriEnd + 1
	if !p.hasClosingSugarBoundary(next) {
		return nil, next, false, nil
	}
	if labelStart == labelEnd || uriStart == uriEnd ||
		p.ctx.text[labelStart] == ' ' || p.ctx.text[labelEnd-1] == ' ' ||
		p.ctx.text[uriStart] == ' ' || p.ctx.text[uriEnd-1] == ' ' {
		return nil, next, false, nil
	}

	content, err := p.parser.parseInlines(
		p.ctx.text[labelStart:labelEnd],
		p.ctx.positionAt(labelStart),
	)
	if err != nil {
		return nil, start, false, err
	}
	definition, ok := p.parser.spec.getInlineDirectiveDefinition("link")
	if !ok {
		return nil, start, false, fmt.Errorf("inline sugar %q has no definition", "link")
	}
	candidate := inlineCandidate{
		attribute:     unescapeASCIIPunctuation(p.ctx.text[uriStart:uriEnd]),
		nestedContent: content,
		rng:           p.ctx.rangeOf(start, next),
	}
	node, accepted, err := buildSugarCandidate("link", definition, candidate)
	if err != nil {
		return nil, start, false, err
	}
	if !accepted {
		return nil, next, false, nil
	}
	return node, next, true, nil
}

func buildSugarCandidate(
	typ string,
	definition inlineDefinition,
	candidate inlineCandidate,
) (ast.Inline, bool, error) {
	if definition.validate != nil && !definition.validate(candidate.attribute) {
		return nil, false, nil
	}
	if definition.build == nil {
		return nil, false, fmt.Errorf("build inline sugar %q: definition has no builder", typ)
	}
	node := definition.build(candidate)
	if node == nil {
		return nil, false, fmt.Errorf("build inline sugar %q: definition returned a nil node", typ)
	}
	return node, true, nil
}

func hasExactDelimiterRun(text string, start int, marker string) bool {
	if !strings.HasPrefix(text[start:], marker) {
		return false
	}
	delimiter := marker[0]
	if start > 0 && text[start-1] == delimiter {
		return false
	}
	end := start + len(marker)
	return end >= len(text) || text[end] != delimiter
}

func isEscapedDelimiter(text string, start int) bool {
	backslashes := 0
	for pos := start - 1; pos >= 0 && text[pos] == '\\'; pos-- {
		backslashes++
	}
	return backslashes%2 == 1
}

func findUnescapedSequence(text string, start, end int, sequence string) int {
	for search := start; search < end; {
		relative := strings.Index(text[search:end], sequence)
		if relative < 0 {
			return -1
		}
		candidate := search + relative
		if !isEscapedDelimiter(text, candidate) {
			return candidate
		}
		search = candidate + 1
	}
	return -1
}

func unescapeASCIIPunctuation(text string) string {
	var value strings.Builder
	value.Grow(len(text))
	for pos := 0; pos < len(text); pos++ {
		if text[pos] == '\\' && pos+1 < len(text) && isASCIIPunctuation(text[pos+1]) {
			pos++
		}
		value.WriteByte(text[pos])
	}
	return value.String()
}

func (p *inlineParseState) logicalLineEnd(start int) int {
	for pos := start; pos < len(p.ctx.text); pos++ {
		if _, ok := p.ctx.logicalNewlineEnd(pos); ok {
			return pos
		}
	}
	return len(p.ctx.text)
}

func (p *inlineParseState) hasOpeningSugarBoundary(start int) bool {
	if start == 0 || p.ctx.text[start-1] == '\n' || p.ctx.text[start-1] == '\r' {
		return true
	}
	previous, _ := utf8.DecodeLastRuneInString(p.ctx.text[:start])
	return previous == ' ' || unicode.IsPunct(previous)
}

func (p *inlineParseState) hasClosingSugarBoundary(end int) bool {
	if end == len(p.ctx.text) {
		return true
	}
	if _, ok := p.ctx.logicalNewlineEnd(end); ok {
		return true
	}
	next, _ := utf8.DecodeRuneInString(p.ctx.text[end:])
	return next == ' ' || unicode.IsPunct(next)
}
