package parser

import (
	"testing"

	"orische/internal/ast"

	"github.com/google/go-cmp/cmp"
)

func TestNewInlineContext_LineStartOffsets(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []int
	}{
		{
			name:  "empty",
			input: "",
			want:  []int{0},
		},
		{
			name:  "LF",
			input: "a\nb",
			want:  []int{0, 2},
		},
		{
			name:  "CRLF",
			input: "a\r\nb",
			want:  []int{0, 3},
		},
		{
			name:  "CR",
			input: "a\rb",
			want:  []int{0, 2},
		},
		{
			name:  "multibyte",
			input: "あ\nい",
			want:  []int{0, 4},
		},
		{
			name:  "mixed line endings",
			input: "line 1 text\nline 2 text\r\n３行目\rline 4 text",
			want:  []int{0, 12, 25, 35},
		},
		{
			name:  "trailing LF",
			input: "a\n",
			want:  []int{0, 2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := newInlineContext(tt.input, ast.Position{Line: 1, Column: 1})

			if diff := cmp.Diff(tt.want, ctx.lineStartOffsets); diff != "" {
				t.Errorf("lineStartOffsets mismatch\n(-want +got)\n%s", diff)
			}
		})
	}
}

func TestInlineContextPositionAt(t *testing.T) {
	input := "あb\ncd"
	ctx := newInlineContext(input, ast.Position{Line: 3, Column: 5})

	tests := []struct {
		name   string
		offset int
		want   ast.Position
	}{
		{
			name:   "origin",
			offset: 0,
			want:   ast.Position{Line: 3, Column: 5},
		},
		{
			name:   "after multibyte code point",
			offset: len("あ"),
			want:   ast.Position{Line: 3, Column: 6},
		},
		{
			name:   "start of second line",
			offset: len("あb\n"),
			want:   ast.Position{Line: 4, Column: 1},
		},
		{
			name:   "second column of second line",
			offset: len("あb\nc"),
			want:   ast.Position{Line: 4, Column: 2},
		},
		{
			name:   "end of input",
			offset: len(input),
			want:   ast.Position{Line: 4, Column: 3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ctx.positionAt(tt.offset)

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf(
					"positionAt(%d) returned wrong position\n(-want +got)\n%s",
					tt.offset,
					diff,
				)
			}
		})
	}
}

func TestInlineContextPositionAt_CRLF(t *testing.T) {
	ctx := newInlineContext("a\r\nb", ast.Position{Line: 1, Column: 1})

	tests := []struct {
		offset int
		want   ast.Position
	}{
		{offset: 1, want: ast.Position{Line: 1, Column: 2}},
		{offset: 2, want: ast.Position{Line: 1, Column: 3}},
		{offset: 3, want: ast.Position{Line: 2, Column: 1}},
	}

	for _, tt := range tests {
		got := ctx.positionAt(tt.offset)
		if diff := cmp.Diff(tt.want, got); diff != "" {
			t.Errorf(
				"positionAt(%d) returned wrong position\n(-want +got)\n%s",
				tt.offset,
				diff,
			)
		}
	}
}

func TestInlineContextRangeOf(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		origin ast.Position
		start  int
		end    int
		want   ast.Range
	}{
		{
			name:   "single ASCII code point",
			input:  "abc",
			origin: ast.Position{Line: 1, Column: 1},
			start:  1,
			end:    2,
			want: ast.Range{
				Start: ast.Position{Line: 1, Column: 2},
				End:   ast.Position{Line: 1, Column: 2},
			},
		},
		{
			name:   "single multibyte code point",
			input:  "aあb",
			origin: ast.Position{Line: 1, Column: 1},
			start:  len("a"),
			end:    len("aあ"),
			want: ast.Range{
				Start: ast.Position{Line: 1, Column: 2},
				End:   ast.Position{Line: 1, Column: 2},
			},
		},
		{
			name:   "multiple lines with non-default origin",
			input:  "ab\ncd",
			origin: ast.Position{Line: 5, Column: 4},
			start:  1,
			end:    len("ab\nc"),
			want: ast.Range{
				Start: ast.Position{Line: 5, Column: 5},
				End:   ast.Position{Line: 6, Column: 1},
			},
		},
		{
			name:   "ASCII through multibyte code point",
			input:  "line 1 text\nline 2 text\r\n３行目\rline 4 text",
			origin: ast.Position{Line: 2, Column: 3},
			start:  17,
			end:    31,
			want: ast.Range{
				Start: ast.Position{Line: 3, Column: 6},
				End:   ast.Position{Line: 4, Column: 2},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := newInlineContext(tt.input, tt.origin)
			got := ctx.rangeOf(tt.start, tt.end)

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("rangeOf() returned wrong range\n(-want +got)\n%s", diff)
			}
		})
	}
}
