package parser

import (
	"reflect"
	"testing"

	"medoc/internal/ast"
)

func TestParseParagraph(t *testing.T) {
	input := &blockContext{
		lines: []string{
			"paragraph line 1",
			"paragraph line 2",
			":::[code:go]",
		},
		pos:    0,
		nested: false,
		parser: nil,
	}

	want := &parsedBlock{
		Type: "Paragraph",
		Attr: "",
		Text: "paragraph line 1\nparagraph line 2\n:::[code:go]",
		Range: ast.Range{
			StartLine: 0,
			EndLine:   3,
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

func TestParseParagraph_EndWithBlankLine(t *testing.T) {
	input := &blockContext{
		lines: []string{
			"paragraph line 1",
			"paragraph line 2",
			"",
			":::[code:go]",
		},
		pos:    0,
		nested: false,
		parser: nil,
	}

	want := &parsedBlock{
		Type: "Paragraph",
		Attr: "",
		Text: "paragraph line 1\nparagraph line 2",
		Range: ast.Range{
			StartLine: 0,
			EndLine:   2,
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
