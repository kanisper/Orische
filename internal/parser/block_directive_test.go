package parser

import (
	"reflect"
	"testing"

	"orische/internal/ast"
)

func TestParseBlockDirective(t *testing.T) {
	input := &blockContext{
		lines: []string{
			":::[code:go]",
			"fmt.Println(\"Hello, world!\")",
			":::",
		},
		pos: 0,
	}

	output, ok, err := (&blockDirectiveParser{}).parse(input)

	want := &parsedBlock{
		Type: "code",
		Attr: "go",
		Text: "fmt.Println(\"Hello, world!\")",
		Range: ast.Range{
			StartLine: 1,
			EndLine:   3,
		},
	}

	ctx_pos_want := 2

	if !ok || err != nil {
		t.Errorf("parse failed.")
	}
	if !reflect.DeepEqual(output, want) {
		t.Errorf("parsed incorrectly.\nwant:\n%v\ngot:\n%v", want, output)
	}
	if input.pos != ctx_pos_want {
		t.Errorf("position in context is not updated correctly. want: %d, got: %d", ctx_pos_want, input.pos)
	}
}

func TestParseBlockDirective_NoAttr(t *testing.T) {
	input := &blockContext{
		lines: []string{
			":::[code]",
			"fmt.Println(\"Hello, world!\")",
			":::",
		},
		pos: 0,
	}

	output, ok, err := (&blockDirectiveParser{}).parse(input)

	want := &parsedBlock{
		Type: "code",
		Attr: "",
		Text: "fmt.Println(\"Hello, world!\")",
		Range: ast.Range{
			StartLine: 1,
			EndLine:   3,
		},
	}

	if !ok || err != nil {
		t.Errorf("parse failed.")
	}
	if !reflect.DeepEqual(output, want) {
		t.Errorf("parsed incorrectly.\nwant:\n%v\ngot:\n%v", want, output)
	}
}

func TestParseBlockDirective_NoClosing(t *testing.T) {
	input := &blockContext{
		lines: []string{
			":::[code:go]",
			"fmt.Println(\"Hello, world!\")",
		},
		pos: 0,
	}

	output, ok, _ := (&blockDirectiveParser{}).parse(input)

	want := &parsedBlock{}

	if ok && !reflect.DeepEqual(output, want) {
		t.Errorf("parsed incorrectly.\nwant:\n%v\ngot:\n%v", want, output)
	}
}

func TestParseBlockDirective_NoType(t *testing.T) {
	input := &blockContext{
		lines: []string{
			":::[:go]",
			"fmt.Println(\"Hello, world!\")",
			":::",
		},
		pos: 0,
	}

	output, ok, _ := (&blockDirectiveParser{}).parse(input)

	want := &parsedBlock{}

	if ok && !reflect.DeepEqual(output, want) {
		t.Errorf("parsed incorrectly.\nwant:\n%v\ngot:\n%v", want, output)
	}
}
