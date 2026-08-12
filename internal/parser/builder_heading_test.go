package parser

import (
	"testing"

	"orische/internal/ast"

	"github.com/google/go-cmp/cmp"
)

func TestBuildHeading(t *testing.T) {
	input := &parsedBlock{
		Type: "Heading",
		Attr: "level1",
		Text: "heading level 1",
		Range: ast.Range{
			Start: ast.Position{Line: 1, Column: 1},
			End:   ast.Position{Line: 1, Column: 17},
		},
	}

	want := &ast.Heading{
		Level: 1,
		Content: []ast.Inline{
			&ast.Text{
				Value: "heading level 1",
				Range: ast.Range{
					Start: ast.Position{Line: 1, Column: 3},
					End:   ast.Position{Line: 1, Column: 17},
				},
			},
		},
		Range: ast.Range{
			Start: ast.Position{Line: 1, Column: 1},
			End:   ast.Position{Line: 1, Column: 17},
		},
	}

	got, err := (&headingBuilder{}).build(input)
	if err != nil {
		t.Fatalf("build returned an error: %v", err)
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("build incorrectly.\n(-want +got)\n%s", diff)
	}
}
