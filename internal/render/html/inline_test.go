package html

import (
	"bytes"
	"testing"

	"orische/internal/ast"
)

func TestInlineRender(t *testing.T) {
	input := []ast.Inline{
		&ast.Text{
			Value: "Plain Text",
		},
		&ast.Emphasis{
			Content: []ast.Inline{
				&ast.Text{
					Value: "Emphasized Text",
				},
			},
		},
		&ast.CodeSpan{
			Value: "Code Span",
		},
		&ast.Link{
			URI: "https://example.com",
			Content: []ast.Inline{
				&ast.Text{
					Value: "Link Text",
				},
			},
		},
	}

	want := "Plain Text<em>Emphasized Text</em><code>Code Span</code><a href=\"https://example.com\">Link Text</a>"

	var buf bytes.Buffer
	renderer := NewRenderer(coreSpec())
	err := renderer.renderInlines(&buf, input)

	if err != nil {
		t.Fatalf("Render Failed\n%s", err)
	}

	if buf.String() != want {
		t.Errorf("Render Failed\nGot: %s\nWant: %s", buf.String(), want)
	}
}
