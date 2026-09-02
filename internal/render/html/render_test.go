package html

import (
	"bytes"
	"errors"
	"reflect"
	"strings"
	"testing"

	"orische/internal/ast"
	"orische/internal/parser"

	"github.com/google/go-cmp/cmp"
)

type errorWriter struct {
	err                       error
	successfulWritesRemaining int
}

func (w *errorWriter) Write(p []byte) (int, error) {
	if w.successfulWritesRemaining > 0 {
		w.successfulWritesRemaining--
		return len(p), nil
	}
	return 0, w.err
}

/*
 * = heading level 1
 *
 * Plain text :[em]{Emphasized Text}
 * :[link:https://example.com]{link text}
 *
 * == heading level 2
 *
 * :::[code:cpp]
 * #include <iostream>
 *
 * int main()
 * {
 *   std::cout << "Hello, world!" << endl;
 *   return 0;
 * }
 * :::
 *
 * === heading level 3
 *
 * # ol level 1 line 1
 * ** ul level 2 line 1
 * ** ul level 2 line 2
 * # ol level 1 line 2
 *
 */

func TestRender(t *testing.T) {
	input := &ast.Document{
		Blocks: []ast.Block{
			&ast.Heading{
				Level: 1,
				Content: []ast.Inline{
					&ast.Text{Value: "heading level 1"},
				},
				Range: ast.Range{
					Start: ast.Position{Line: 1, Column: 1},
					End:   ast.Position{Line: 1, Column: 17},
				},
			},
			&ast.Paragraph{
				Content: []ast.Inline{
					&ast.Text{Value: "Plain text "},
					&ast.Emphasis{
						Content: []ast.Inline{
							&ast.Text{Value: "Emphasized Text"},
						},
					},
					&ast.Text{Value: "\n"},
					&ast.Link{
						URI: "https://example.com",
						Content: []ast.Inline{
							&ast.Text{Value: "link text"},
						},
					},
				},
				Range: ast.Range{
					Start: ast.Position{Line: 3, Column: 1},
					End:   ast.Position{Line: 4, Column: 38},
				},
			},
			&ast.Heading{
				Level: 2,
				Content: []ast.Inline{
					&ast.Text{Value: "heading level 2"},
				},
				Range: ast.Range{
					Start: ast.Position{Line: 6, Column: 1},
					End:   ast.Position{Line: 6, Column: 18},
				},
			},
			&ast.CodeBlock{
				Language: "cpp",
				Text: `#include <iostream>

int main()
{
  std::cout << "Hello, world!" << endl;
  return 0;
}`,
				Range: ast.Range{
					Start: ast.Position{Line: 8, Column: 1},
					End:   ast.Position{Line: 16, Column: 3},
				},
			},
			&ast.Heading{
				Level: 3,
				Content: []ast.Inline{
					&ast.Text{Value: "heading level 3"},
				},
				Range: ast.Range{
					Start: ast.Position{Line: 18, Column: 1},
					End:   ast.Position{Line: 18, Column: 19},
				},
			},
			&ast.List{
				Ordered: true,
				Items: []*ast.ListItem{
					{
						Blocks: []ast.Block{
							&ast.Paragraph{
								Content: []ast.Inline{
									&ast.Text{Value: "ol level 1 line 1"},
								},
								Range: ast.Range{
									Start: ast.Position{Line: 20, Column: 1},
									End:   ast.Position{Line: 20, Column: 19},
								},
							},
							&ast.List{
								Ordered: false,
								Items: []*ast.ListItem{
									{
										Blocks: []ast.Block{
											&ast.Paragraph{
												Content: []ast.Inline{
													&ast.Text{Value: "ul level 2 line 1"},
												},
												Range: ast.Range{
													Start: ast.Position{Line: 21, Column: 1},
													End:   ast.Position{Line: 21, Column: 20},
												},
											},
										},
										Range: ast.Range{
											Start: ast.Position{Line: 21, Column: 1},
											End:   ast.Position{Line: 21, Column: 20},
										},
									},
									{
										Blocks: []ast.Block{
											&ast.Paragraph{
												Content: []ast.Inline{
													&ast.Text{Value: "ul level 2 line 2"},
												},
												Range: ast.Range{
													Start: ast.Position{Line: 22, Column: 1},
													End:   ast.Position{Line: 22, Column: 20},
												},
											},
										},
										Range: ast.Range{
											Start: ast.Position{Line: 22, Column: 1},
											End:   ast.Position{Line: 22, Column: 20},
										},
									},
								},
								Range: ast.Range{
									Start: ast.Position{Line: 21, Column: 1},
									End:   ast.Position{Line: 22, Column: 20},
								},
							},
						},
						Range: ast.Range{
							Start: ast.Position{Line: 20, Column: 1},
							End:   ast.Position{Line: 22, Column: 20},
						},
					},
					{
						Blocks: []ast.Block{
							&ast.Paragraph{
								Content: []ast.Inline{
									&ast.Text{Value: "ol level 1 line 2"},
								},
								Range: ast.Range{
									Start: ast.Position{Line: 23, Column: 1},
									End:   ast.Position{Line: 23, Column: 19},
								},
							},
						},
						Range: ast.Range{
							Start: ast.Position{Line: 23, Column: 1},
							End:   ast.Position{Line: 23, Column: 19},
						},
					},
				},
				Range: ast.Range{
					Start: ast.Position{Line: 20, Column: 1},
					End:   ast.Position{Line: 23, Column: 19},
				},
			},
		},
		Range: ast.Range{
			Start: ast.Position{Line: 1, Column: 1},
			End:   ast.Position{Line: 23, Column: 19},
		},
	}

	want := `<h1>heading level 1</h1>
<p>
Plain text <em>Emphasized Text</em>
<a href="https://example.com">link text</a>
</p>
<h2>heading level 2</h2>
<pre><code data-language="cpp">
#include &lt;iostream&gt;

int main()
{
  std::cout &lt;&lt; &#34;Hello, world!&#34; &lt;&lt; endl;
  return 0;
}
</code></pre>
<h3>heading level 3</h3>
<ol>
<li>ol level 1 line 1</li>
<li>
<ul>
<li>ul level 2 line 1</li>
<li>ul level 2 line 2</li>
</ul>
</li>
<li>ol level 1 line 2</li>
</ol>
`

	var buf bytes.Buffer
	renderer := NewRenderer()
	err := renderer.Render(&buf, input)

	if err != nil {
		t.Fatalf("rendering failed: %s", err)
	}

	if diff := cmp.Diff(want, buf.String()); diff != "" {
		t.Errorf("rendered incorrectly\n(-want, +got)\n%s", diff)
	}
}

func TestRenderParsedInlineSugar(t *testing.T) {
	doc, err := parser.Parse("*important* **visually bold** _emphasized_ __visually italic__ --removed-- ~old information~")
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	var buf bytes.Buffer
	if err := NewRenderer().Render(&buf, doc); err != nil {
		t.Fatalf("render failed: %v", err)
	}
	want := "<p>\n<strong>important</strong> <b>visually bold</b> <em>emphasized</em> " +
		"<i>visually italic</i> <del>removed</del> <s>old information</s>\n</p>\n"
	if buf.String() != want {
		t.Errorf("rendered incorrectly\nGot:  %s\nWant: %s", buf.String(), want)
	}
}

func TestRenderPropagatesWriterError(t *testing.T) {
	wantErr := errors.New("write failed")
	doc := &ast.Document{Blocks: []ast.Block{
		&ast.Paragraph{Content: []ast.Inline{&ast.Text{Value: "text"}}},
	}}

	tests := []struct {
		name                      string
		successfulWritesRemaining int
		wantContexts              []string
	}{
		{
			name:         "block write",
			wantContexts: []string{`render "*ast.Paragraph" block`},
		},
		{
			name:                      "nested inline write",
			successfulWritesRemaining: 1,
			wantContexts: []string{
				`render "*ast.Paragraph" block`,
				`render "*ast.Text" inline node`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			writer := &errorWriter{
				err:                       wantErr,
				successfulWritesRemaining: tt.successfulWritesRemaining,
			}
			err := Render(writer, doc)
			if !errors.Is(err, wantErr) {
				t.Fatalf("expected writer error, got %v", err)
			}
			for _, context := range tt.wantContexts {
				if !strings.Contains(err.Error(), context) {
					t.Errorf("error %q does not contain context %q", err, context)
				}
			}
		})
	}
}

func TestRenderReportsMissingBlockRenderer(t *testing.T) {
	renderer := NewRenderer()
	delete(renderer.spec.blockRenderers, reflect.TypeFor[*ast.Heading]())

	err := renderer.Render(&bytes.Buffer{}, &ast.Document{Blocks: []ast.Block{&ast.Heading{}}})
	if err == nil {
		t.Fatal("rendering succeeded without a registered block renderer")
	}
	if !strings.Contains(err.Error(), `not found the renderer for "*ast.Heading" block`) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRenderReportsMissingInlineRenderer(t *testing.T) {
	renderer := NewRenderer()
	delete(renderer.spec.inlineRenderers, reflect.TypeFor[*ast.Text]())
	doc := &ast.Document{Blocks: []ast.Block{
		&ast.Paragraph{Content: []ast.Inline{&ast.Text{Value: "text"}}},
	}}

	err := renderer.Render(&bytes.Buffer{}, doc)
	if err == nil {
		t.Fatal("rendering succeeded without a registered inline renderer")
	}
	if !strings.Contains(err.Error(), `not found the renderer for "*ast.Text" inline node`) {
		t.Fatalf("unexpected error: %v", err)
	}
}
