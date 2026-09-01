package parser

import (
	"sort"
	"unicode/utf8"

	"orische/internal/ast"
)

type inlineContext struct {
	text             string
	origin           ast.Position
	lineStartOffsets []int
}

func newInlineContext(text string, origin ast.Position) *inlineContext {
	offsets := []int{0}

	for i := 0; i < len(text); i++ {
		switch text[i] {
		case '\r':
			if i+1 >= len(text) || text[i+1] != '\n' {
				offsets = append(offsets, i+1)
			}

		case '\n':
			offsets = append(offsets, i+1)

		default:
			continue
		}
	}

	return &inlineContext{
		text:             text,
		origin:           origin,
		lineStartOffsets: offsets,
	}
}

func (ctx *inlineContext) logicalNewlineEnd(start int) (int, bool) {
	if start < 0 || start >= len(ctx.text) {
		return start, false
	}

	switch ctx.text[start] {
	case '\n':
		return start + 1, true
	case '\r':
		if start+1 < len(ctx.text) && ctx.text[start+1] == '\n' {
			return start + 2, true
		}
		return start + 1, true
	default:
		return start, false
	}
}

// positionAt converts a zero-based UTF-8 byte offset to a one-based
// Unicode code point position.
//
// bytesOffset must be in [0, len(ctx.text)] and must point to a UTF-8
// code point boundary.
func (ctx *inlineContext) positionAt(bytesOffset int) ast.Position {
	idx := sort.Search(len(ctx.lineStartOffsets), func(i int) bool {
		return ctx.lineStartOffsets[i] > bytesOffset
	})

	lineIndex := idx - 1
	lineStart := ctx.lineStartOffsets[lineIndex]

	pos := ctx.origin
	pos.Line += lineIndex

	if lineIndex == 0 {
		pos.Column += utf8.RuneCountInString(ctx.text[:bytesOffset])
	} else {
		pos.Column = 1 + utf8.RuneCountInString(ctx.text[lineStart:bytesOffset])
	}

	return pos
}

// rangeOf converts the non-empty byte span [start, end) to a one-based,
// inclusive Unicode code point range.
//
// start and end must be valid UTF-8 code point boundaries.
// The span must not be empty.
func (ctx *inlineContext) rangeOf(start, end int) ast.Range {
	_, size := utf8.DecodeLastRuneInString(ctx.text[start:end])

	return ast.Range{
		Start: ctx.positionAt(start),
		End:   ctx.positionAt(end - size),
	}
}
