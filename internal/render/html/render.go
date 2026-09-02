package html

import (
	"io"

	"orische/internal/ast"
)

type Renderer struct {
	spec *spec
}

func NewRenderer() *Renderer {
	return &Renderer{spec: coreSpec()}
}

func Render(w io.Writer, doc *ast.Document) error {
	return NewRenderer().Render(w, doc)
}

func (r *Renderer) Render(w io.Writer, doc *ast.Document) error {
	for _, block := range doc.Blocks {
		if err := r.renderBlock(w, block); err != nil {
			return err
		}
	}

	return nil
}
