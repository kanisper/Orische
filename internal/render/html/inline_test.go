package html

import (
	"bytes"
	"strings"
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

func TestInlineRenderEscapesHTML(t *testing.T) {
	input := []ast.Inline{
		&ast.Text{Value: `<strong>"text" & 'text'</strong>`},
		&ast.CodeSpan{Value: `left < right && right > "value"`},
		&ast.Link{
			URI:     `https://example.com/search?q="quoted"&page=1`,
			Content: []ast.Inline{&ast.Text{Value: `<link>`}},
		},
	}
	want := "&lt;strong&gt;&#34;text&#34; &amp; &#39;text&#39;&lt;/strong&gt;" +
		"<code>left &lt; right &amp;&amp; right &gt; &#34;value&#34;</code>" +
		"<a href=\"https://example.com/search?q=&#34;quoted&#34;&amp;page=1\">&lt;link&gt;</a>"

	var buf bytes.Buffer
	if err := NewRenderer(coreSpec()).renderInlines(&buf, input); err != nil {
		t.Fatalf("rendering failed: %s", err)
	}

	if buf.String() != want {
		t.Errorf("rendered incorrectly\nGot:  %s\nWant: %s", buf.String(), want)
	}
}

func TestLinkRendererAllowsSupportedSchemes(t *testing.T) {
	tests := []struct {
		name string
		uri  string
	}{
		{name: "http", uri: "http://example.com"},
		{name: "https", uri: "https://example.com"},
		{name: "mailto", uri: "mailto:user@example.com"},
		{name: "case insensitive", uri: "HTTPS://example.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := &ast.Link{URI: tt.uri, Content: []ast.Inline{&ast.Text{Value: "link"}}}
			want := `<a href="` + tt.uri + `">link</a>`

			var buf bytes.Buffer
			err := (&linkRenderer{}).render(NewRenderer(coreSpec()), &buf, input)
			if err != nil {
				t.Fatalf("rendering failed: %s", err)
			}
			if buf.String() != want {
				t.Errorf("rendered incorrectly\nGot:  %s\nWant: %s", buf.String(), want)
			}
		})
	}
}

func TestLinkRendererRejectsUnsupportedSchemes(t *testing.T) {
	tests := []struct {
		name string
		uri  string
	}{
		{name: "javascript", uri: "javascript:alert(1)"},
		{name: "data", uri: "data:text/html,unsafe"},
		{name: "ftp", uri: "ftp://example.com/file"},
		{name: "relative", uri: "/documentation"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := &ast.Link{URI: tt.uri, Content: []ast.Inline{&ast.Text{Value: "link"}}}

			var buf bytes.Buffer
			err := (&linkRenderer{}).render(NewRenderer(coreSpec()), &buf, input)
			if err == nil {
				t.Fatal("rendering succeeded for an unsupported URI scheme")
			}
			if !strings.Contains(err.Error(), "unsupported URI scheme") {
				t.Fatalf("unexpected error: %s", err)
			}
			if buf.Len() != 0 {
				t.Errorf("renderer wrote partial output: %q", buf.String())
			}
		})
	}
}
