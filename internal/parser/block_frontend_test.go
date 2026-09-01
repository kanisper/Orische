package parser

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"orische/internal/ast"
)

func TestReadBlockDirective(t *testing.T) {
	input := &blockContext{
		lines: []string{"before", ":::[code:go]", "fmt.Println(\"日😀\")", ":::", "after"},
		start: 1,
	}

	got, consumed := readBlockDirective(input)
	want := &blockDirectiveNode{
		dirtype:       "code",
		attribute:     "go",
		text:          "fmt.Println(\"日😀\")",
		contentOrigin: ast.Position{Line: 3, Column: 1},
		rng: ast.Range{
			Start: ast.Position{Line: 2, Column: 1},
			End:   ast.Position{Line: 4, Column: 3},
		},
	}
	if consumed != 3 {
		t.Fatalf("readBlockDirective consumed %d lines, want 3", consumed)
	}
	if diff := cmp.Diff(want, got, cmp.AllowUnexported(blockDirectiveNode{})); diff != "" {
		t.Errorf("directive mismatch (-want +got):\n%s", diff)
	}
}

func TestReadBlockDirectiveFallsBackWhenUnterminated(t *testing.T) {
	got, consumed := readBlockDirective(&blockContext{lines: []string{":::[code]", "text"}})
	if got != nil || consumed != 0 {
		t.Errorf("unterminated directive = (%#v, %d), want (nil, 0)", got, consumed)
	}
}

func TestReadParagraph(t *testing.T) {
	input := &blockContext{
		lines: []string{"before", "日😀", "second", "", "after"},
		start: 1,
	}

	got, consumed := readParagraph(input)
	want := &paragraphNode{
		text:          "日😀\nsecond",
		contentOrigin: ast.Position{Line: 2, Column: 1},
		rng: ast.Range{
			Start: ast.Position{Line: 2, Column: 1},
			End:   ast.Position{Line: 3, Column: 6},
		},
	}
	if consumed != 2 {
		t.Fatalf("readParagraph consumed %d lines, want 2", consumed)
	}
	if diff := cmp.Diff(want, got, cmp.AllowUnexported(paragraphNode{})); diff != "" {
		t.Errorf("paragraph mismatch (-want +got):\n%s", diff)
	}
}
