package parser

import (
	"testing"

	"medoc/internal/ast"

	"github.com/google/go-cmp/cmp"
)

func TestBuildParagraph(t *testing.T) {
	input := &parsedBlock{
		Type:  "paragraph",
		Attr:  "",
		Text:  "First Line\nThis is :[em]{second line}\nThird line",
		Range: ast.Range{StartLine: 1, EndLine: 3},
	}

	want := &ast.Paragraph{
		Content: []ast.Inline{
			&ast.Text{Value: "First Line\nThis is "},
			&ast.Emphasis{
				Content: []ast.Inline{
					&ast.Text{Value: "second line"},
				},
			},
			&ast.Text{Value: "\nThird line"},
		},
		Range: ast.Range{StartLine: 1, EndLine: 3},
	}

	got, err := (&paragraphBuilder{}).build(input)
	if err != nil {
		t.Errorf("build failed: %v", err)
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("build incorrectly.\n(-want +got)\n%s", diff)
	}
}
