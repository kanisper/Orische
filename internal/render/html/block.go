package html

import (
	"fmt"
	"io"

	"orische/internal/ast"
)

func (r *Renderer) renderBlock(w io.Writer, block ast.Block) error {
	renderer, ok := r.spec.getBlockRenderer(block)
	if !ok {
		return fmt.Errorf("block renderer: not found the renderer for %T type of block", block)
	}

	return renderer.render(r, w, block)
}

// heading
type headingRenderer struct{}

func (*headingRenderer) render(r *Renderer, w io.Writer, heading *ast.Heading) error {
	_, err := fmt.Fprintf(w, "<h%d>", heading.Level)
	if err != nil {
		return fmt.Errorf("heading renderer: %w", err)
	}

	err = r.renderInlines(w, heading.Content)
	if err != nil {
		return fmt.Errorf("heading renderer: %w", err)
	}

	_, err = fmt.Fprintf(w, "</h%d>\n", heading.Level)
	if err != nil {
		return fmt.Errorf("heading renderer: %w", err)
	}

	return nil
}

// code
type codeblockRenderer struct{}

func (*codeblockRenderer) render(_ *Renderer, w io.Writer, codeblock *ast.CodeBlock) error {
	_, err := fmt.Fprintf(w,
		"<pre><code data-language=\"%s\">\n%s\n</code></pre>\n",
		codeblock.Language,
		codeblock.Text,
	)
	if err != nil {
		return fmt.Errorf("codeblock renderer: %w", err)
	}

	return nil
}

// list
type listRenderer struct{}

func (*listRenderer) render(r *Renderer, w io.Writer, list *ast.List) error {
	var err error
	if list.Ordered {
		_, err = fmt.Fprintln(w, "<ol>")
	} else {
		_, err = fmt.Fprintln(w, "<ul>")
	}
	if err != nil {
		return fmt.Errorf("list renderer: %w", err)
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
		return fmt.Errorf("list renderer: %w", err)
	}

	return nil
}

func renderListItem(r *Renderer, w io.Writer, item *ast.ListItem) error {
	for _, block := range item.Blocks {
		switch b := block.(type) {
		case *ast.Paragraph:
			_, err := fmt.Fprint(w, "<li>")
			if err != nil {
				return fmt.Errorf("list item renderer: %w", err)
			}

			err = r.renderInlines(w, b.Content)
			if err != nil {
				return fmt.Errorf("list item renderer: %w", err)
			}

			_, err = fmt.Fprint(w, "</li>\n")
			if err != nil {
				return fmt.Errorf("list item renderer: %w", err)
			}

		case *ast.List:
			_, err := fmt.Fprintln(w, "<li>")
			if err != nil {
				return fmt.Errorf("list item renderer: %w", err)
			}

			err = (&listRenderer{}).render(r, w, b)
			if err != nil {
				return fmt.Errorf("list item renderer: %w", err)
			}

			_, err = fmt.Fprintln(w, "</li>")
			if err != nil {
				return fmt.Errorf("list item renderer: %w", err)
			}

		default:
			panic("unreachable: list item block type must be paragraph or list.")
		}
	}

	return nil
}

// paragraph
type paragraphRenderer struct{}

func (*paragraphRenderer) render(r *Renderer, w io.Writer, paragraph *ast.Paragraph) error {
	_, err := fmt.Fprintln(w, "<p>")
	if err != nil {
		return fmt.Errorf("paragraph renderer: %w", err)
	}

	err = r.renderInlines(w, paragraph.Content)
	if err != nil {
		return fmt.Errorf("paragraph renderer: %w", err)
	}

	_, err = fmt.Fprintln(w, "\n</p>")
	if err != nil {
		return fmt.Errorf("paragraph renderer: %w", err)
	}

	return nil
}
