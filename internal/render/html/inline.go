package html

import (
	"fmt"
	"html"
	"io"
	"net/url"
	"strings"

	"orische/internal/ast"
)

func (r *Renderer) renderInlines(w io.Writer, inlines []ast.Inline) error {
	for _, inline := range inlines {
		err := r.renderOneInline(w, inline)
		if err != nil {
			return fmt.Errorf(
				"render \"%T\" inline node: %w",
				inline,
				err,
			)
		}
	}

	return nil
}

func (r *Renderer) renderOneInline(w io.Writer, inline ast.Inline) error {
	renderer, ok := r.spec.getInlineRenderer(inline)
	if !ok {
		return fmt.Errorf("not found the renderer for \"%T\" inline node", inline)
	}

	return renderer(r, w, inline)
}

func renderText(_ *Renderer, w io.Writer, text *ast.Text) error {
	_, err := fmt.Fprintf(w, "%s", html.EscapeString(text.Value))
	if err != nil {
		return err
	}

	return nil
}

func renderLineBreak(_ *Renderer, w io.Writer, _ *ast.LineBreak) error {
	_, err := fmt.Fprint(w, "<br>")
	return err
}

func renderEmphasis(r *Renderer, w io.Writer, emphasis *ast.Emphasis) error {
	return renderInlineContainer(r, w, "em", emphasis.Content)
}

func renderStrong(r *Renderer, w io.Writer, strong *ast.Strong) error {
	return renderInlineContainer(r, w, "strong", strong.Content)
}

func renderItalic(r *Renderer, w io.Writer, italic *ast.Italic) error {
	return renderInlineContainer(r, w, "i", italic.Content)
}

func renderBold(r *Renderer, w io.Writer, bold *ast.Bold) error {
	return renderInlineContainer(r, w, "b", bold.Content)
}

func renderUnderline(r *Renderer, w io.Writer, underline *ast.Underline) error {
	return renderInlineContainer(r, w, "u", underline.Content)
}

func renderStrikethrough(r *Renderer, w io.Writer, strike *ast.Strikethrough) error {
	return renderInlineContainer(r, w, "s", strike.Content)
}

func renderInlineContainer(r *Renderer, w io.Writer, tag string, content []ast.Inline) error {
	_, err := fmt.Fprintf(w, "<%s>", tag)
	if err != nil {
		return err
	}

	err = r.renderInlines(w, content)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(w, "</%s>", tag)
	if err != nil {
		return err
	}

	return nil
}

func renderCodeSpan(_ *Renderer, w io.Writer, codespan *ast.CodeSpan) error {
	_, err := fmt.Fprint(w, "<code>")
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(w, "%s", html.EscapeString(codespan.Value))
	if err != nil {
		return err
	}

	_, err = fmt.Fprint(w, "</code>")
	if err != nil {
		return err
	}

	return nil
}

func renderLink(r *Renderer, w io.Writer, link *ast.Link) error {
	parsedURI, err := url.Parse(link.URI)
	if err != nil {
		return err
	}

	switch {
	case strings.EqualFold(parsedURI.Scheme, "http"):
	case strings.EqualFold(parsedURI.Scheme, "https"):
	case strings.EqualFold(parsedURI.Scheme, "mailto"):
	default:
		return fmt.Errorf("unsupported URI scheme %q", parsedURI.Scheme)
	}

	_, err = fmt.Fprintf(w, "<a href=\"%s\">", html.EscapeString(link.URI))
	if err != nil {
		return err
	}

	err = r.renderInlines(w, link.Content)
	if err != nil {
		return err
	}

	_, err = fmt.Fprint(w, "</a>")
	if err != nil {
		return err
	}

	return nil
}
