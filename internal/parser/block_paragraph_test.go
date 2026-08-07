package parser

import (
	"reflect"
	"testing"

	"orische/internal/ast"
)

func TestParseParagraph(t *testing.T) {
	input := &blockContext{
		lines: []string{
			"paragraph line 1",
			"paragraph line 2",
			":::[code:go]",
		},
		pos: 0,
	}

	want := &parsedBlock{
		Type: "Paragraph",
		Attr: "",
		Text: "paragraph line 1\nparagraph line 2\n:::[code:go]",
		Range: ast.Range{
			Start: ast.Position{Line: 1, Column: 1},
			End:   ast.Position{Line: 3, Column: 12},
		},
	}

	ctx_pos_want := 2

	output, ok, err := (&paragraphParser{}).parse(input)

	if !ok || err != nil {
		t.Errorf("parse failed.")
	}
	if !reflect.DeepEqual(output, want) {
		t.Errorf("parse incorrectly.\nwant:\n%v\ngot:\n%v", want, output)
	}
	if input.pos != ctx_pos_want {
		t.Errorf("position in context is not updated correctly. want: %d, got: %d", ctx_pos_want, input.pos)
	}
}

func TestParseParagraph_EndWithBlankLine(t *testing.T) {
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
		Type: "Paragraph",
		Attr: "",
		Text: "paragraph line 1\nparagraph line 2",
		Range: ast.Range{
			Start: ast.Position{Line: 1, Column: 1},
			End:   ast.Position{Line: 2, Column: 16},
		},
	}

	output, ok, err := (&paragraphParser{}).parse(input)

	if !ok || err != nil {
		t.Errorf("parse failed.")
	}

	if !reflect.DeepEqual(output, want) {
		t.Errorf("parse incorrectly.\nwant:\n%v\ngot:\n%v", want, output)
	}
}
