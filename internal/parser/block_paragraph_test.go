package parser

import (
	"testing"

	"orische/internal/ast"

	"github.com/google/go-cmp/cmp"
)

func TestReadParagraph(t *testing.T) {
	input := &blockContext{
		lines: []string{
			"paragraph line 1",
			"paragraph line 2",
			":::[code:go]",
		},
		pos: 0,
	}

	want := &parsedBlock{
		Type: blockBuilderKeyParagraph,
		Attr: "",
		Text: "paragraph line 1\nparagraph line 2\n:::[code:go]",
		Range: ast.Range{
			Start: ast.Position{Line: 1, Column: 1},
			End:   ast.Position{Line: 3, Column: 12},
		},
	}

	ctx_pos_want := 2

	output, ok, err := (&paragraphReader{}).read(input)

	if err != nil {
		t.Fatalf("read returned an error: %v", err)
	}
	if !ok {
		t.Fatal("paragraph reader did not recognize valid input")
	}
	if diff := cmp.Diff(want, output); diff != "" {
		t.Errorf("parsed incorrectly\n(-want +got)\n%s", diff)
	}
	if input.pos != ctx_pos_want {
		t.Errorf("position in context is not updated correctly. want: %d, got: %d", ctx_pos_want, input.pos)
	}
}

func TestReadParagraph_EndWithBlankLine(t *testing.T) {
	input := &blockContext{
		lines: []string{
			"paragraph line 1",
			"paragraph line 2",
			"",
			":::[code:go]",
		},
		pos: 0,
	}

	want := &parsedBlock{
		Type: blockBuilderKeyParagraph,
		Attr: "",
		Text: "paragraph line 1\nparagraph line 2",
		Range: ast.Range{
			Start: ast.Position{Line: 1, Column: 1},
			End:   ast.Position{Line: 2, Column: 16},
		},
	}

	output, ok, err := (&paragraphReader{}).read(input)

	if err != nil {
		t.Fatalf("read returned an error: %v", err)
	}
	if !ok {
		t.Fatal("paragraph reader did not recognize valid input")
	}

	if diff := cmp.Diff(want, output); diff != "" {
		t.Errorf("parsed incorrectly\n(-want +got)\n%s", diff)
	}
}

func TestReadParagraph_UnicodeRange(t *testing.T) {
	input := &blockContext{
		lines: []string{"first line", "é😀"},
	}

	got, ok, err := (&paragraphReader{}).read(input)
	if err != nil {
		t.Fatalf("read returned an error: %v", err)
	}
	if !ok {
		t.Fatal("paragraph reader did not recognize valid Unicode input")
	}

	want := &parsedBlock{
		Type: blockBuilderKeyParagraph,
		Text: "first line\né😀",
		Range: ast.Range{
			Start: ast.Position{Line: 1, Column: 1},
			End:   ast.Position{Line: 2, Column: 2},
		},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("parsed incorrectly\n(-want +got)\n%s", diff)
	}
	if input.pos != 1 {
		t.Errorf("position in context is not updated correctly. want 1, got %d", input.pos)
	}
}
