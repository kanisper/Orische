package parser

import (
	"testing"

	"orische/internal/ast"

	"github.com/google/go-cmp/cmp"
)

func TestReadHeading(t *testing.T) {
	input := &blockContext{
		lines: []string{"= Heading1"},
		pos:   0,
	}
	output, ok, err := (&headingReader{}).read(input)
	want := &parsedBlock{
		Type: "Heading",
		Attr: "level1",
		Text: "Heading1",
		Range: ast.Range{
			Start: ast.Position{Line: 1, Column: 1},
			End:   ast.Position{Line: 1, Column: 10},
		},
	}
	if err != nil {
		t.Fatalf("read returned an error: %v", err)
	}
	if !ok {
		t.Fatal("heading reader did not recognize valid input")
	}
	if diff := cmp.Diff(want, output); diff != "" {
		t.Errorf("parsed incorrectly\n(-want +got)\n%s", diff)
	}
}

func TestReadHeading_Level2(t *testing.T) {
	input := &blockContext{
		lines: []string{"== Heading2"},
		pos:   0,
	}
	output, ok, err := (&headingReader{}).read(input)
	want := &parsedBlock{
		Type: "Heading",
		Attr: "level2",
		Text: "Heading2",
		Range: ast.Range{
			Start: ast.Position{Line: 1, Column: 1},
			End:   ast.Position{Line: 1, Column: 11},
		},
	}
	if err != nil {
		t.Fatalf("read returned an error: %v", err)
	}
	if !ok {
		t.Fatal("heading reader did not recognize valid input")
	}
	if diff := cmp.Diff(want, output); diff != "" {
		t.Errorf("parsed incorrectly\n(-want +got)\n%s", diff)
	}
}

func TestReadHeading_UnicodeRange(t *testing.T) {
	input := &blockContext{lines: []string{"= あ😀"}}

	got, ok, err := (&headingReader{}).read(input)
	if err != nil {
		t.Fatalf("read returned an error: %v", err)
	}
	if !ok {
		t.Fatal("heading reader did not recognize valid Unicode input")
	}

	want := &parsedBlock{
		Type: "Heading",
		Attr: "level1",
		Text: "あ😀",
		Range: ast.Range{
			Start: ast.Position{Line: 1, Column: 1},
			End:   ast.Position{Line: 1, Column: 4},
		},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("parsed incorrectly\n(-want +got)\n%s", diff)
	}
}

func TestReadHeading_LevelOutOfRange(t *testing.T) {
	input := &blockContext{
		lines: []string{"======= Heading7"},
		pos:   0,
	}

	output, ok, err := (&headingReader{}).read(input)
	if err != nil {
		t.Fatalf("read returned an error: %v", err)
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

func TestReadHeading_NoSpace(t *testing.T) {
	input := &blockContext{
		lines: []string{"before", "=Heading", "after"},
		pos:   1,
	}
	output, ok, err := (&headingReader{}).read(input)
	if err != nil {
		t.Fatalf("read returned an error: %v", err)
	}
	if ok {
		t.Error("heading without a separator space was parsed successfully")
	}
	if output != nil {
		t.Errorf("heading without a separator space returned a node: %v", output)
	}
	if input.pos != 1 {
		t.Errorf("position in context changed. want 1, got %d", input.pos)
	}
}

func TestReadHeading_NoText(t *testing.T) {
	input := &blockContext{
		lines: []string{"before", "=", "after"},
		pos:   1,
	}
	output, ok, err := (&headingReader{}).read(input)
	if err != nil {
		t.Fatalf("read returned an error: %v", err)
	}
	if ok {
		t.Error("heading without a separator or content was parsed successfully")
	}
	if output != nil {
		t.Errorf("heading without a separator or content returned a node: %v", output)
	}
	if input.pos != 1 {
		t.Errorf("position in context changed. want 1, got %d", input.pos)
	}
}
