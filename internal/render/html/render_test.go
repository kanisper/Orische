package html

import (
	"bytes"
	"testing"

	"orische/internal/ast"

	"github.com/google/go-cmp/cmp"
)

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
				Range: ast.Range{StartLine: 1, EndLine: 1},
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
				Range: ast.Range{StartLine: 3, EndLine: 4},
			},
			&ast.Heading{
				Level: 2,
				Content: []ast.Inline{
					&ast.Text{Value: "heading level 2"},
				},
				Range: ast.Range{StartLine: 6, EndLine: 6},
			},
			&ast.CodeBlock{
				Language: "cpp",
				Text: `#include <iostream>

int main()
{
  std::cout << "Hello, world!" << endl;
  return 0;
}`,
				Range: ast.Range{StartLine: 8, EndLine: 16},
			},
			&ast.Heading{
				Level: 3,
				Content: []ast.Inline{
					&ast.Text{Value: "heading level 3"},
				},
				Range: ast.Range{StartLine: 18, EndLine: 18},
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
								Range: ast.Range{StartLine: 20, EndLine: 20},
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
												Range: ast.Range{StartLine: 21, EndLine: 21},
											},
										},
										Range: ast.Range{StartLine: 21, EndLine: 21},
									},
									{
										Blocks: []ast.Block{
											&ast.Paragraph{
												Content: []ast.Inline{
													&ast.Text{Value: "ul level 2 line 2"},
												},
												Range: ast.Range{StartLine: 22, EndLine: 22},
											},
										},
										Range: ast.Range{StartLine: 22, EndLine: 22},
									},
								},
								Range: ast.Range{StartLine: 21, EndLine: 22},
							},
						},
						Range: ast.Range{StartLine: 20, EndLine: 22},
					},
					{
						Blocks: []ast.Block{
							&ast.Paragraph{
								Content: []ast.Inline{
									&ast.Text{Value: "ol level 1 line 2"},
								},
								Range: ast.Range{StartLine: 23, EndLine: 23},
							},
						},
						Range: ast.Range{StartLine: 23, EndLine: 23},
					},
				},
				Range: ast.Range{StartLine: 20, EndLine: 23},
			},
		},
		Range: ast.Range{StartLine: 1, EndLine: 23},
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
	renderer := NewRenderer(coreSpec())
	err := renderer.Render(&buf, input)

	if err != nil {
		t.Fatalf("rendering failed: %s", err)
	}

	if diff := cmp.Diff(want, buf.String()); diff != "" {
		t.Errorf("rendered incorrectly\n(-want, +got)\n%s", diff)
	}
}
