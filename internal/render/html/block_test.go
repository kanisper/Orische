package html

import (
	"bytes"
	"strings"
	"testing"

	"orische/internal/ast"
)

func TestHeadingRenderer(t *testing.T) {
	input := &ast.Heading{
		Level: 1,
		Content: []ast.Inline{
			&ast.Text{Value: "Heading level 1 text"},
		},
	}

	want := "<h1>Heading level 1 text</h1>\n"

	var buf bytes.Buffer
	renderer := NewRenderer()
	err := renderer.renderBlock(&buf, input)

	if err != nil {
		t.Fatalf("rendering failed: %s", err)
	}

	if buf.String() != want {
		t.Errorf("rendered incorrectly\nWant:\n%s\nGot:\n%s", want, buf.String())
	}
}

func TestCodeBlockRenderer(t *testing.T) {
	input := &ast.CodeBlock{
		Language: "go",
		Text:     "fmt.Println(\"Hello, world!\")",
	}

	want_array := []string{
		"<pre><code data-language=\"go\">",
		"fmt.Println(&#34;Hello, world!&#34;)",
		"</code></pre>",
		"",
	}
	want := strings.Join(want_array, "\n")

	var buf bytes.Buffer
	renderer := NewRenderer()
	err := renderer.renderBlock(&buf, input)

	if err != nil {
		t.Fatalf("rendering failed: %s", err)
	}
	if buf.String() != want {
		t.Errorf("rendered incorrectly\nWant:\n%s\nGot:\n%s", want, buf.String())
	}
}

func TestListRenderer(t *testing.T) {
	input := &ast.List{
		Ordered: false,
		Items: []*ast.ListItem{
			{
				Blocks: []ast.Block{
					&ast.Paragraph{
						Content: []ast.Inline{
							&ast.Text{Value: "ul level 1 line 1"},
						},
					},
					&ast.List{
						Ordered: true,
						Items: []*ast.ListItem{
							{
								Blocks: []ast.Block{
									&ast.Paragraph{
										Content: []ast.Inline{
											&ast.Text{Value: "ol level 2 line 1"},
										},
									},
								},
							},
						},
					},
				},
			},
			{
				Blocks: []ast.Block{
					&ast.Paragraph{
						Content: []ast.Inline{
							&ast.Text{Value: "ul level 1 line 2"},
						},
					},
				},
			},
		},
	}

	want_array := []string{
		"<ul>",
		"<li>ul level 1 line 1</li>",
		"<li>",
		"<ol>",
		"<li>ol level 2 line 1</li>",
		"</ol>",
		"</li>",
		"<li>ul level 1 line 2</li>",
		"</ul>",
		"",
	}
	want := strings.Join(want_array, "\n")

	var buf bytes.Buffer
	renderer := NewRenderer()
	err := renderer.renderBlock(&buf, input)

	if err != nil {
		t.Fatalf("rendering failed: %s", err)
	}
	if buf.String() != want {
		t.Errorf("rendered incorrectly\nWant:\n%s\nGot:\n%s", want, buf.String())
	}
}

func TestParagraphRenderer(t *testing.T) {
	input := &ast.Paragraph{
		Content: []ast.Inline{
			&ast.Text{Value: "Paragraph text"},
		},
	}

	want := "<p>\nParagraph text\n</p>\n"

	var buf bytes.Buffer
	renderer := NewRenderer()
	err := renderer.renderBlock(&buf, input)

	if err != nil {
		t.Fatalf("rendering failed: %s", err)
	}
	if buf.String() != want {
		t.Errorf("rendered incorrectly\nWant:\n%s\nGot:\n%s", want, buf.String())
	}
}
