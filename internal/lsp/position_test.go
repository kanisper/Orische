package lsp

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"go.lsp.dev/protocol"

	"orische/internal/ast"
)

func TestPositionMapperEncodings(t *testing.T) {
	source := "A日🍣Z"
	tests := []struct {
		name      string
		encoding  protocol.PositionEncodingKind
		positions []protocol.Position
		offsets   []int
	}{
		{
			name: "utf8", encoding: protocol.PositionEncodingKindUTF8,
			positions: []protocol.Position{{Character: 0}, {Character: 1}, {Character: 4}, {Character: 8}, {Character: 9}},
			offsets:   []int{0, 1, 4, 8, 9},
		},
		{
			name: "utf16", encoding: protocol.PositionEncodingKindUTF16,
			positions: []protocol.Position{{Character: 0}, {Character: 1}, {Character: 2}, {Character: 4}, {Character: 5}},
			offsets:   []int{0, 1, 4, 8, 9},
		},
		{
			name: "utf32", encoding: protocol.PositionEncodingKindUTF32,
			positions: []protocol.Position{{Character: 0}, {Character: 1}, {Character: 2}, {Character: 3}, {Character: 4}},
			offsets:   []int{0, 1, 4, 8, 9},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mapper, err := newPositionMapper(source, tt.encoding)
			if err != nil {
				t.Fatal(err)
			}
			for index, position := range tt.positions {
				offset, err := mapper.byteOffset(position)
				if err != nil {
					t.Fatalf("byteOffset(%v): %v", position, err)
				}
				if offset != tt.offsets[index] {
					t.Errorf("byteOffset(%v) = %d, want %d", position, offset, tt.offsets[index])
				}
				gotPosition, err := mapper.position(offset)
				if err != nil {
					t.Fatalf("position(%d): %v", offset, err)
				}
				if diff := cmp.Diff(position, gotPosition); diff != "" {
					t.Errorf("position(%d) mismatch (-want +got):\n%s", offset, diff)
				}
			}
		})
	}
}

func TestPositionMapperLineEndings(t *testing.T) {
	mapper, err := newPositionMapper("a\nb\r\nc\rd", protocol.PositionEncodingKindUTF16)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		offset   int
		position protocol.Position
	}{
		{0, protocol.Position{Line: 0, Character: 0}},
		{1, protocol.Position{Line: 0, Character: 1}},
		{2, protocol.Position{Line: 1, Character: 0}},
		{3, protocol.Position{Line: 1, Character: 1}},
		{5, protocol.Position{Line: 2, Character: 0}},
		{6, protocol.Position{Line: 2, Character: 1}},
		{7, protocol.Position{Line: 3, Character: 0}},
		{8, protocol.Position{Line: 3, Character: 1}},
	}
	for _, tt := range tests {
		gotPosition, err := mapper.position(tt.offset)
		if err != nil {
			t.Fatalf("position(%d): %v", tt.offset, err)
		}
		if diff := cmp.Diff(tt.position, gotPosition); diff != "" {
			t.Errorf("position(%d) mismatch (-want +got):\n%s", tt.offset, diff)
		}
		gotOffset, err := mapper.byteOffset(tt.position)
		if err != nil {
			t.Fatalf("byteOffset(%v): %v", tt.position, err)
		}
		if gotOffset != tt.offset {
			t.Errorf("byteOffset(%v) = %d, want %d", tt.position, gotOffset, tt.offset)
		}
	}

	if _, err := mapper.position(4); err == nil {
		t.Fatal("position inside CRLF succeeded")
	}
	if _, err := mapper.byteOffset(protocol.Position{Line: 4}); err == nil {
		t.Fatal("position past final line succeeded")
	}
}

func TestPositionMapperTrailingLineEnding(t *testing.T) {
	for _, source := range []string{"a\n", "a\r", "a\r\n"} {
		mapper, err := newPositionMapper(source, protocol.PositionEncodingKindUTF16)
		if err != nil {
			t.Fatal(err)
		}
		position, err := mapper.position(len(source))
		if err != nil {
			t.Fatalf("source %q: %v", source, err)
		}
		want := protocol.Position{Line: 1, Character: 0}
		if diff := cmp.Diff(want, position); diff != "" {
			t.Errorf("source %q EOF mismatch (-want +got):\n%s", source, diff)
		}
	}
}

func TestPositionMapperRejectsInvalidBoundaries(t *testing.T) {
	mapper, err := newPositionMapper("A日🍣Z", protocol.PositionEncodingKindUTF16)
	if err != nil {
		t.Fatal(err)
	}

	invalidPositions := []protocol.Position{
		{Line: 1},
		{Character: 3},
		{Character: 6},
	}
	for _, position := range invalidPositions {
		if _, err := mapper.byteOffset(position); err == nil {
			t.Errorf("byteOffset(%v) succeeded", position)
		}
	}
	for _, offset := range []int{-1, 2, 5, 10} {
		if _, err := mapper.position(offset); err == nil {
			t.Errorf("position(%d) succeeded", offset)
		}
	}

	if _, err := newPositionMapper("\xff", protocol.PositionEncodingKindUTF16); err == nil {
		t.Fatal("invalid UTF-8 source succeeded")
	}
	if _, err := newPositionMapper("text", protocol.PositionEncodingKind("unknown")); err == nil {
		t.Fatal("unknown encoding succeeded")
	}
}

func TestPositionMapperASTConversions(t *testing.T) {
	mapper, err := newPositionMapper("A🍣\r\n日本", protocol.PositionEncodingKindUTF16)
	if err != nil {
		t.Fatal(err)
	}

	position, err := mapper.astPosition(ast.Position{Line: 1, Column: 2})
	if err != nil {
		t.Fatal(err)
	}
	if want := (protocol.Position{Line: 0, Character: 1}); position != want {
		t.Errorf("AST position = %v, want %v", position, want)
	}

	rangeTests := []struct {
		name string
		in   ast.Range
		want protocol.Range
	}{
		{
			name: "emoji",
			in: ast.Range{
				Start: ast.Position{Line: 1, Column: 2},
				End:   ast.Position{Line: 1, Column: 2},
			},
			want: protocol.Range{
				Start: protocol.Position{Line: 0, Character: 1},
				End:   protocol.Position{Line: 0, Character: 3},
			},
		},
		{
			name: "multiline",
			in: ast.Range{
				Start: ast.Position{Line: 1, Column: 2},
				End:   ast.Position{Line: 2, Column: 1},
			},
			want: protocol.Range{
				Start: protocol.Position{Line: 0, Character: 1},
				End:   protocol.Position{Line: 1, Character: 1},
			},
		},
	}
	for _, tt := range rangeTests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := mapper.astRange(tt.in)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("range mismatch (-want +got):\n%s", diff)
			}
		})
	}

	invalidRanges := []ast.Range{
		{},
		{Start: ast.Position{Line: 1, Column: 3}, End: ast.Position{Line: 1, Column: 3}},
		{Start: ast.Position{Line: 2, Column: 1}, End: ast.Position{Line: 1, Column: 1}},
	}
	for _, sourceRange := range invalidRanges {
		if _, err := mapper.astRange(sourceRange); err == nil {
			t.Errorf("astRange(%v) succeeded", sourceRange)
		}
	}
}

func TestNegotiatePositionEncoding(t *testing.T) {
	tests := []struct {
		name string
		in   []protocol.PositionEncodingKind
		want protocol.PositionEncodingKind
	}{
		{name: "default", want: protocol.PositionEncodingKindUTF16},
		{name: "client preference", in: []protocol.PositionEncodingKind{
			protocol.PositionEncodingKindUTF32,
			protocol.PositionEncodingKindUTF8,
		}, want: protocol.PositionEncodingKindUTF32},
		{name: "unsupported", in: []protocol.PositionEncodingKind{"other"}, want: protocol.PositionEncodingKindUTF16},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := negotiatePositionEncoding(tt.in); got != tt.want {
				t.Errorf("encoding = %q, want %q", got, tt.want)
			}
		})
	}
}
