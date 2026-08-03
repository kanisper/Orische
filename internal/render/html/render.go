package html

import (
	"io"

	"orische/internal/ast"
)

type Renderer struct {
	spec *Spec
}

//	func NewRenderer(spec *Spec) *Renderer {
//		if spec == nil {
//			spec = coreSpec()
//		}
//		return &Renderer{spec: spec}
//	}
func NewRenderer() *Renderer {
	return &Renderer{spec: coreSpec()}
}

func Render(w io.Writer, doc *ast.Document) error {
	return NewRenderer().Render(w, doc)
}

func (r *Renderer) Render(w io.Writer, doc *ast.Document) error {
	for _, block := range doc.Blocks {
		err := r.renderBlock(w, block)
		if err != nil {
			return err
		}
	}

	return nil
}
