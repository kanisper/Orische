package completion

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"orische/internal/ast"
)

func TestCompleteDirectiveTypes(t *testing.T) {
	tests := []struct {
		name   string
		source string
		ast    *ast.Document
		want   []Candidate
	}{
		{
			name:   "block prefix without AST",
			source: ":::[hea",
			want: []Candidate{{
				Label: "heading", Replace: Span{Start: 4, End: 7},
				InsertText: "heading", InsertFormat: InsertFormatPlainText,
			}},
		},
		{
			name:   "block prefix with AST",
			source: ":::[HEA",
			ast:    &ast.Document{},
			want: []Candidate{{
				Label: "heading", Replace: Span{Start: 4, End: 7},
				InsertText: "heading", InsertFormat: InsertFormatPlainText,
			}},
		},
		{
			name:   "inline prefix",
			source: ":[str",
			want: []Candidate{{
				Label: "strong", Replace: Span{Start: 2, End: 5},
				InsertText: "strong", InsertFormat: InsertFormatPlainText,
			}},
		},
		{
			name:   "unicode before inline prefix",
			source: "日本 :[str",
			want: []Candidate{{
				Label: "strong", Replace: Span{Start: len("日本 :["), End: len("日本 :[str")},
				InsertText: "strong", InsertFormat: InsertFormatPlainText,
			}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Complete(Request{
				Source: tt.source, CursorOffset: len(tt.source), AST: tt.ast,
			})
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("candidates mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestCompleteEmptyDirectivePrefixes(t *testing.T) {
	tests := []struct {
		name       string
		source     string
		wantLabels []string
	}{
		{name: "block", source: ":::[", wantLabels: []string{"heading", "paragraph", "code"}},
		{name: "inline", source: ":[", wantLabels: []string{"em", "strong", "italic", "bold", "del", "outdated", "link", "code"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Complete(Request{Source: tt.source, CursorOffset: len(tt.source)})
			labels := make([]string, len(got))
			for index := range got {
				labels[index] = got[index].Label
				if got[index].Replace != (Span{Start: len(tt.source), End: len(tt.source)}) {
					t.Errorf("candidate %q span = %v", got[index].Label, got[index].Replace)
				}
			}
			if diff := cmp.Diff(tt.wantLabels, labels); diff != "" {
				t.Errorf("labels mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestCompleteRejectsNonDirectiveContexts(t *testing.T) {
	tests := []string{
		"ordinary text",
		"x:::[hea",
		":[str]",
		":[link:",
		`\:[str`,
		":[zzz",
		":::[zzz",
	}
	for _, source := range tests {
		t.Run(source, func(t *testing.T) {
			if got := Complete(Request{Source: source, CursorOffset: len(source)}); len(got) != 0 {
				t.Errorf("Complete(%q) = %#v, want no candidates", source, got)
			}
		})
	}
}

func TestCompleteSuppressesDirectiveTypesInCodeBlock(t *testing.T) {
	source := ":::[code]\n:[str\n:::"
	document := &ast.Document{Blocks: []ast.Block{
		&ast.CodeBlock{Range: ast.Range{
			Start: ast.Position{Line: 1, Column: 1},
			End:   ast.Position{Line: 3, Column: 3},
		}},
	}}
	if got := Complete(Request{
		Source: source, CursorOffset: len(":::[code]\n:[str"), AST: document,
	}); len(got) != 0 {
		t.Errorf("completion inside Code Block = %#v, want none", got)
	}
}

func TestCompleteAllowsBlockTypeOnCodeBlockOpener(t *testing.T) {
	source := ":::[code]\nbody\n:::"
	document := &ast.Document{Blocks: []ast.Block{
		&ast.CodeBlock{Range: ast.Range{
			Start: ast.Position{Line: 1, Column: 1},
			End:   ast.Position{Line: 3, Column: 3},
		}},
	}}
	want := []Candidate{{
		Label: "code", Replace: Span{Start: 4, End: 6},
		InsertText: "code", InsertFormat: InsertFormatPlainText,
	}}
	if got := Complete(Request{Source: source, CursorOffset: len(":::[co"), AST: document}); !cmp.Equal(got, want) {
		t.Errorf("completion on Code Block opener = %#v, want %#v", got, want)
	}
}

func TestCompleteSuppressesDirectiveTypesInCodeSpan(t *testing.T) {
	source := "text :[code]{:[str}"
	document := &ast.Document{Blocks: []ast.Block{
		&ast.Paragraph{
			Range: ast.Range{Start: ast.Position{Line: 1, Column: 1}, End: ast.Position{Line: 1, Column: 19}},
			Content: []ast.Inline{
				&ast.Text{Value: "text ", Range: ast.Range{
					Start: ast.Position{Line: 1, Column: 1}, End: ast.Position{Line: 1, Column: 5},
				}},
				&ast.CodeSpan{Value: ":[str", Range: ast.Range{
					Start: ast.Position{Line: 1, Column: 6}, End: ast.Position{Line: 1, Column: 19},
				}},
			},
		},
	}}
	if got := Complete(Request{Source: source, CursorOffset: len(source) - 1, AST: document}); len(got) != 0 {
		t.Errorf("completion inside Code Span = %#v, want none", got)
	}
}

func TestCompleteRejectsInvalidCursor(t *testing.T) {
	source := "日:[str"
	for _, offset := range []int{-1, 1, len(source) + 1} {
		if got := Complete(Request{Source: source, CursorOffset: offset}); len(got) != 0 {
			t.Errorf("cursor %d returned %#v", offset, got)
		}
	}
}
