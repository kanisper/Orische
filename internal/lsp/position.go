package lsp

import (
	"fmt"
	"unicode/utf8"

	"go.lsp.dev/protocol"

	"orische/internal/ast"
)

type positionMapper struct {
	source   string
	encoding protocol.PositionEncodingKind
	lines    []sourceLine
}

type sourceLine struct {
	start         int
	end           int
	terminatorEnd int
}

func newPositionMapper(source string, encoding protocol.PositionEncodingKind) (*positionMapper, error) {
	if !utf8.ValidString(source) {
		return nil, fmt.Errorf("position mapper: source is not valid UTF-8")
	}
	if !supportedPositionEncoding(encoding) {
		return nil, fmt.Errorf("position mapper: unsupported position encoding %q", encoding)
	}

	return &positionMapper{
		source:   source,
		encoding: encoding,
		lines:    indexSourceLines(source),
	}, nil
}

func (m *positionMapper) byteOffset(position protocol.Position) (int, error) {
	if uint64(position.Line) >= uint64(len(m.lines)) {
		return 0, fmt.Errorf("LSP line %d is outside the document", position.Line)
	}

	line := m.lines[position.Line]
	target := uint64(position.Character)
	units := uint64(0)
	for offset := line.start; offset < line.end; {
		if units == target {
			return offset, nil
		}

		r, size := utf8.DecodeRuneInString(m.source[offset:line.end])
		width := uint64(m.encodedRuneWidth(r, size))
		if target < units+width {
			return 0, fmt.Errorf("LSP character %d is inside an encoded character", position.Character)
		}
		units += width
		offset += size
	}

	// LSP positions past a valid line's encoded length default to its end.
	return line.end, nil
}

func (m *positionMapper) position(offset int) (protocol.Position, error) {
	if offset < 0 || offset > len(m.source) {
		return protocol.Position{}, fmt.Errorf("byte offset %d is outside the document", offset)
	}
	if offset < len(m.source) && offset > 0 && !utf8.RuneStart(m.source[offset]) {
		return protocol.Position{}, fmt.Errorf("byte offset %d is inside a UTF-8 character", offset)
	}

	for lineNumber, line := range m.lines {
		switch {
		case offset >= line.start && offset <= line.end:
			units := 0
			for cursor := line.start; cursor < offset; {
				r, size := utf8.DecodeRuneInString(m.source[cursor:offset])
				units += m.encodedRuneWidth(r, size)
				cursor += size
			}
			return protocol.Position{Line: uint32(lineNumber), Character: uint32(units)}, nil
		case offset > line.end && offset < line.terminatorEnd:
			return protocol.Position{}, fmt.Errorf("byte offset %d is inside a line terminator", offset)
		}
	}

	return protocol.Position{}, fmt.Errorf("byte offset %d cannot be represented as an LSP position", offset)
}

func (m *positionMapper) astPosition(position ast.Position) (protocol.Position, error) {
	offset, err := m.astPositionOffset(position, true)
	if err != nil {
		return protocol.Position{}, err
	}
	return m.position(offset)
}

func (m *positionMapper) astRange(sourceRange ast.Range) (protocol.Range, error) {
	startOffset, err := m.astPositionOffset(sourceRange.Start, false)
	if err != nil {
		return protocol.Range{}, fmt.Errorf("map AST range start: %w", err)
	}
	endOffset, err := m.astPositionOffset(sourceRange.End, false)
	if err != nil {
		return protocol.Range{}, fmt.Errorf("map AST range end: %w", err)
	}
	if startOffset > endOffset {
		return protocol.Range{}, fmt.Errorf("map AST range: start follows end")
	}

	_, endSize := utf8.DecodeRuneInString(m.source[endOffset:])
	endExclusive := endOffset + endSize
	start, err := m.position(startOffset)
	if err != nil {
		return protocol.Range{}, err
	}
	end, err := m.position(endExclusive)
	if err != nil {
		return protocol.Range{}, err
	}
	return protocol.Range{Start: start, End: end}, nil
}

func (m *positionMapper) astPositionOffset(position ast.Position, allowLineEnd bool) (int, error) {
	if position.Line < 1 || position.Line > len(m.lines) {
		return 0, fmt.Errorf("AST line %d is outside the document", position.Line)
	}
	if position.Column < 1 {
		return 0, fmt.Errorf("AST column %d is invalid", position.Column)
	}

	line := m.lines[position.Line-1]
	target := position.Column - 1
	column := 0
	for offset := line.start; offset < line.end; {
		if column == target {
			return offset, nil
		}
		_, size := utf8.DecodeRuneInString(m.source[offset:line.end])
		column++
		offset += size
	}
	if allowLineEnd && column == target {
		return line.end, nil
	}
	return 0, fmt.Errorf("AST column %d is outside line %d", position.Column, position.Line)
}

func (m *positionMapper) encodedRuneWidth(r rune, utf8Width int) int {
	switch m.encoding {
	case protocol.PositionEncodingKindUTF8:
		return utf8Width
	case protocol.PositionEncodingKindUTF16:
		if r > 0xffff {
			return 2
		}
		return 1
	case protocol.PositionEncodingKindUTF32:
		return 1
	default:
		panic("position mapper created with unsupported encoding")
	}
}

func indexSourceLines(source string) []sourceLine {
	lines := make([]sourceLine, 0, 1)
	lineStart := 0
	for offset := 0; offset < len(source); {
		switch source[offset] {
		case '\n':
			lines = append(lines, sourceLine{start: lineStart, end: offset, terminatorEnd: offset + 1})
			offset++
			lineStart = offset
		case '\r':
			terminatorEnd := offset + 1
			if terminatorEnd < len(source) && source[terminatorEnd] == '\n' {
				terminatorEnd++
			}
			lines = append(lines, sourceLine{start: lineStart, end: offset, terminatorEnd: terminatorEnd})
			offset = terminatorEnd
			lineStart = offset
		default:
			_, size := utf8.DecodeRuneInString(source[offset:])
			offset += size
		}
	}
	lines = append(lines, sourceLine{start: lineStart, end: len(source), terminatorEnd: len(source)})
	return lines
}

func supportedPositionEncoding(encoding protocol.PositionEncodingKind) bool {
	switch encoding {
	case protocol.PositionEncodingKindUTF8,
		protocol.PositionEncodingKindUTF16,
		protocol.PositionEncodingKindUTF32:
		return true
	default:
		return false
	}
}

func negotiatePositionEncoding(encodings []protocol.PositionEncodingKind) protocol.PositionEncodingKind {
	for _, encoding := range encodings {
		if supportedPositionEncoding(encoding) {
			return encoding
		}
	}
	return protocol.PositionEncodingKindUTF16
}
