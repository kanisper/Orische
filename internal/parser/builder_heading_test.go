package parser

import (
	"testing"

	"medoc/internal/ast"

	"github.com/google/go-cmp/cmp"
)

func TestBuildHeading(t *testing.T) {
	input := &parsedBlock{
		Type:  "Heading",
		Attr:  "level1",
		Text:  "heading level 1",
		Range: ast.Range{StartLine: 0, EndLine: 0},
	}

	want := &ast.Heading{
		Level: 1,
		Content: []ast.Inline{
			&ast.Text{Value: "heading level 1"},
		},
		Range: ast.Range{StartLine: 0, EndLine: 0},
	}

	got, err := (&headingBuilder{}).build(input)
	if err != nil {
		t.Errorf("build failed: %v", err)
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("build incorrectly.\n(-want +got)\n%s", diff)
	}
}
