package parser

import (
	"reflect"
	"testing"

	"medoc/internal/ast"
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
			StartLine: 0,
			EndLine:   2,
		},
	}

	if !ok || err != nil {
		t.Errorf("parse failed.")
	}
	if !reflect.DeepEqual(output, want) {
		t.Errorf("parsed incorrectly. want: %v, got: %v", want, output)
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
			StartLine: 0,
			EndLine:   2,
		},
	}

	if !ok || err != nil {
		t.Errorf("parse failed.")
	}
	if !reflect.DeepEqual(output, want) {
		t.Errorf("parsed incorrectly. want: %v, got: %v", want, output)
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
		t.Errorf("parsed incorrectly. want: %v, got: %v", want, output)
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
		t.Errorf("parsed incorrectly. want: %v, got: %v", want, output)
	}
}
