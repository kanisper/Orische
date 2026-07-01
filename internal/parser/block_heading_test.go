package parser

import (
	"reflect"
	"testing"

	"medoc/internal/ast"
)

func TestParseHeading(t *testing.T) {
	input := &blockContext{
		lines:  []string{"= Heading1"},
		pos:    0,
		nested: false,
		parser: nil,
	}
	output, ok, err := (&headingParser{}).parse(input)
	want := &parsedBlock{
		Type: "Heading",
		Attr: "level1",
		Text: "Heading1",
		Range: ast.Range{
			StartLine: 0,
			EndLine:   0,
		},
	}
	if !ok || err != nil {
		t.Errorf("parse failed.")
	}
	if !reflect.DeepEqual(output, want) {
		t.Errorf("parsed incorrectly. want: %v, got: %v", want, output)
	}
}

func TestParseHeading_Level2(t *testing.T) {
	input := &blockContext{
		lines:  []string{"== Heading2"},
		pos:    0,
		nested: false,
		parser: nil,
	}
	output, ok, err := (&headingParser{}).parse(input)
	want := &parsedBlock{
		Type: "Heading",
		Attr: "level2",
		Text: "Heading2",
		Range: ast.Range{
			StartLine: 0,
			EndLine:   0,
		},
	}
	if !ok || err != nil {
		t.Errorf("parse failed.")
	}
	if !reflect.DeepEqual(output, want) {
		t.Errorf("parsed incorrectly. want: %v, got: %v", want, output)
	}
}

func TestParseHeading_NoSpace(t *testing.T) {
	input := &blockContext{
		lines:  []string{"=Heading"},
		pos:    0,
		nested: false,
		parser: nil,
	}
	output, ok, _ := (&headingParser{}).parse(input)
	want := &parsedBlock{}
	if ok || !reflect.DeepEqual(output, want) {
		t.Errorf("parsed incorrectly. want: %v, got: %v", want, output)
	}
}

func TestParseHeading_NoText(t *testing.T) {
	input := &blockContext{
		lines:  []string{"="},
		pos:    0,
		nested: false,
		parser: nil,
	}
	output, ok, _ := (&headingParser{}).parse(input)
	want := &parsedBlock{}
	if ok || !reflect.DeepEqual(output, want) {
		t.Errorf("parsed incorrectly. want: %v, got: %v", want, output)
	}
}
