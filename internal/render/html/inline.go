package html

import (
	"fmt"
	"html"
	"io"

	"orische/internal/ast"
)

func (r *Renderer) renderInlines(w io.Writer, inlines []ast.Inline) error {
	for _, inline := range inlines {
		err := r.renderOneInline(w, inline)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *Renderer) renderOneInline(w io.Writer, inline ast.Inline) error {
	renderer, ok := r.spec.getInlineRenderer(inline)
	if !ok {
		return fmt.Errorf("inline renderer: not found the renderer for %T type of inline", inline)
	}

	return renderer.render(r, w, inline)
}

// Text
type textRenderer struct{}

func (*textRenderer) render(_ *Renderer, w io.Writer, text *ast.Text) error {
	_, err := fmt.Fprintf(w, "%s", html.EscapeString(text.Value))
	if err != nil {
		return fmt.Errorf("plain text render: %w", err)
	}

	return nil
}

// Emphasis
type emphasisRenderer struct{}

func (*emphasisRenderer) render(r *Renderer, w io.Writer, emphasis *ast.Emphasis) error {
	_, err := fmt.Fprint(w, "<em>")
	if err != nil {
		return fmt.Errorf("emphasis text render: %w", err)
	}

	err = r.renderInlines(w, emphasis.Content)
	if err != nil {
		return fmt.Errorf("emphasis text render: %w", err)
	}

	_, err = fmt.Fprint(w, "</em>")
	if err != nil {
		return fmt.Errorf("emphasis text render: %w", err)
	}

	return nil
}

// CodeSpan
type codespanRenderer struct{}

func (*codespanRenderer) render(_ *Renderer, w io.Writer, codespan *ast.CodeSpan) error {
	_, err := fmt.Fprint(w, "<code>")
	if err != nil {
		return fmt.Errorf("codespan render: %w", err)
	}

	_, err = fmt.Fprintf(w, "%s", html.EscapeString(codespan.Value))
	if err != nil {
		return fmt.Errorf("codespan render: %w", err)
	}

	_, err = fmt.Fprint(w, "</code>")

	return nil
}

// Link
type linkRenderer struct{}

func (*linkRenderer) render(r *Renderer, w io.Writer, link *ast.Link) error {
	_, err := fmt.Fprintf(w, "<a href=\"%s\">", html.EscapeString(link.URI))
	if err != nil {
		return fmt.Errorf("link render: %w", err)
	}

	err = r.renderInlines(w, link.Content)
	if err != nil {
		return fmt.Errorf("link render: %w", err)
	}

	_, err = fmt.Fprint(w, "</a>")
	if err != nil {
		return fmt.Errorf("link render: %w", err)
	}

	return nil
}
