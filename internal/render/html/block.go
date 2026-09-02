package html

import (
	"fmt"
	"html"
	"io"

	"orische/internal/ast"
)

func (r *Renderer) renderBlock(w io.Writer, block ast.Block) error {
	renderer, ok := r.spec.getBlockRenderer(block)
	if !ok {
		return fmt.Errorf("renderBlock: not found the renderer for \"%T\" block", block)
	}

	err := renderer(r, w, block)
	if err != nil {
		return fmt.Errorf(
			"render \"%T\" block: %w",
			block,
			err,
		)
	}

	return nil
}

func renderHeading(r *Renderer, w io.Writer, heading *ast.Heading) error {
	_, err := fmt.Fprintf(w, "<h%d>", heading.Level)
	if err != nil {
		return err
	}

	err = r.renderInlines(w, heading.Content)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(w, "</h%d>\n", heading.Level)
	if err != nil {
		return err
	}

	return nil
}

func renderCodeBlock(_ *Renderer, w io.Writer, codeblock *ast.CodeBlock) error {
	_, err := fmt.Fprintf(w,
		"<pre><code data-language=\"%s\">\n%s\n</code></pre>\n",
		html.EscapeString(codeblock.Language),
		html.EscapeString(codeblock.Text),
	)
	if err != nil {
		return err
	}

	return nil
}

func renderList(r *Renderer, w io.Writer, list *ast.List) error {
	var err error
	if list.Ordered {
		_, err = fmt.Fprintln(w, "<ol>")
	} else {
		_, err = fmt.Fprintln(w, "<ul>")
	}
	if err != nil {
		return err
	}

	for _, item := range list.Items {
		err = renderListItem(r, w, item)
		if err != nil {
			return err
		}
	}

	if list.Ordered {
		_, err = fmt.Fprintln(w, "</ol>")
	} else {
		_, err = fmt.Fprintln(w, "</ul>")
	}
	if err != nil {
		return err
	}

	return nil
}

func renderListItem(r *Renderer, w io.Writer, item *ast.ListItem) error {
	for _, block := range item.Blocks {
		switch b := block.(type) {
		case *ast.Paragraph:
			_, err := fmt.Fprint(w, "<li>")
			if err != nil {
				return err
			}

			err = r.renderInlines(w, b.Content)
			if err != nil {
				return err
			}

			_, err = fmt.Fprint(w, "</li>\n")
			if err != nil {
				return err
			}

		case *ast.List:
			_, err := fmt.Fprintln(w, "<li>")
			if err != nil {
				return err
			}

			err = renderList(r, w, b)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(w, "</li>")
			if err != nil {
				return err
			}

		default:
			panic("unreachable: list item block type must be paragraph or list.")
		}
	}

	return nil
}

func renderParagraph(r *Renderer, w io.Writer, paragraph *ast.Paragraph) error {
	_, err := fmt.Fprintln(w, "<p>")
	if err != nil {
		return err
	}

	err = r.renderInlines(w, paragraph.Content)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintln(w, "\n</p>")
	if err != nil {
		return err
	}

	return nil
}
