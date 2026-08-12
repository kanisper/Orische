package parser

import (
	"testing"

	"orische/internal/ast"

	"github.com/google/go-cmp/cmp"
)

func TestBuildParagraph(t *testing.T) {
	input := &parsedBlock{
		Type: "paragraph",
		Attr: "",
		Text: "First Line\nThis is :[em]{second line}\nThird line",
		Range: ast.Range{
			Start: ast.Position{Line: 1, Column: 1},
			End:   ast.Position{Line: 3, Column: 10},
		},
	}

	want := &ast.Paragraph{
		Content: []ast.Inline{
			&ast.Text{
				Value: "First Line\nThis is ",
				Range: ast.Range{
					Start: ast.Position{Line: 1, Column: 1},
					End:   ast.Position{Line: 2, Column: 8},
				},
			},
			&ast.Emphasis{
				Content: []ast.Inline{
					&ast.Text{
						Value: "second line",
						Range: ast.Range{
							Start: ast.Position{Line: 2, Column: 15},
							End:   ast.Position{Line: 2, Column: 25},
						},
					},
				},
				Range: ast.Range{
					Start: ast.Position{Line: 2, Column: 9},
					End:   ast.Position{Line: 2, Column: 26},
				},
			},
			&ast.Text{
				Value: "\nThird line",
				Range: ast.Range{
					Start: ast.Position{Line: 2, Column: 27},
					End:   ast.Position{Line: 3, Column: 10},
				},
			},
		},
		Range: ast.Range{
			Start: ast.Position{Line: 1, Column: 1},
			End:   ast.Position{Line: 3, Column: 10},
		},
	}

	got, err := (&paragraphBuilder{}).build(input)
	if err != nil {
		t.Fatalf("build returned an error: %v", err)
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("build incorrectly.\n(-want +got)\n%s", diff)
	}
}
