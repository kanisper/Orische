package parser

import (
	"reflect"
	"testing"

	"orische/internal/ast"
)

func TestParseHeading(t *testing.T) {
	input := &blockContext{
		lines: []string{"= Heading1"},
		pos:   0,
	}
	output, ok, err := (&headingParser{}).parse(input)
	want := &parsedBlock{
		Type: "Heading",
		Attr: "level1",
		Text: "Heading1",
		Range: ast.Range{
			StartLine: 1,
			EndLine:   1,
		},
	}
	if !ok || err != nil {
		t.Errorf("parse failed.")
	}
	if !reflect.DeepEqual(output, want) {
		t.Errorf("parsed incorrectly.\nwant:\n%v\ngot:\n%v", want, output)
	}
}

func TestParseHeading_Level2(t *testing.T) {
	input := &blockContext{
		lines: []string{"== Heading2"},
		pos:   0,
	}
	output, ok, err := (&headingParser{}).parse(input)
	want := &parsedBlock{
		Type: "Heading",
		Attr: "level2",
		Text: "Heading2",
		Range: ast.Range{
			StartLine: 1,
			EndLine:   1,
		},
	}
	if !ok || err != nil {
		t.Errorf("parse failed.")
	}
	if !reflect.DeepEqual(output, want) {
		t.Errorf("parsed incorrectly.\nwant:\n%v\ngot:\n%v", want, output)
	}
}

func TestParseHeading_LevelOutOfRange(t *testing.T) {
	input := &blockContext{
		lines: []string{"======= Heading7"},
		pos:   0,
	}

	output, ok, err := (&headingParser{}).parse(input)
	if err != nil {
		t.Fatalf("parse returned an error: %v", err)
	}
	if ok {
		t.Error("heading above level 6 was parsed successfully")
	}
	if output != nil {
		t.Errorf("heading above level 6 returned a node: %v", output)
	}
	if input.pos != 0 {
		t.Errorf("position in context changed. want 0, got %d", input.pos)
	}
}

func TestParseHeading_NoSpace(t *testing.T) {
	input := &blockContext{
		lines: []string{"=Heading"},
		pos:   0,
	}
	output, ok, err := (&headingParser{}).parse(input)
	if err != nil {
		t.Fatalf("parse returned an error: %v", err)
	}
	if ok {
		t.Error("heading without a separator space was parsed successfully")
	}
	if output != nil {
		t.Errorf("heading without a separator space returned a node: %v", output)
	}
}

func TestParseHeading_NoText(t *testing.T) {
	input := &blockContext{
		lines: []string{"="},
		pos:   0,
	}
	output, ok, err := (&headingParser{}).parse(input)
	if err != nil {
		t.Fatalf("parse returned an error: %v", err)
	}
	if ok {
		t.Error("heading without a separator or content was parsed successfully")
	}
	if output != nil {
		t.Errorf("heading without a separator or content returned a node: %v", output)
	}
}
