package html

type Renderer struct {
	spec *Spec
}

func NewRenderer(spec *Spec) *Renderer {
	if spec == nil {
		spec = coreSpec()
	}
	return &Renderer{spec: spec}
}

// func Render(doc *ast.Document) (string, error) {
// 	return NewRenderer(coreSpec()).Render(doc)
// }

// func (r *Renderer) Render(doc *ast.Document) (string, error)
