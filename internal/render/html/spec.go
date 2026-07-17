package html

import (
	"io"
	"reflect"

	"orische/internal/ast"
)

type Spec struct {
	blockRenderers  map[reflect.Type]blockRenderer
	inlineRenderers map[reflect.Type]inlineRenderer
}

type blockRenderer interface {
	render(w io.Writer, block ast.Block) error
}

type inlineRenderer interface {
	render(w io.Writer, inline ast.Inline) error
}

func newSpec() *Spec {
	return &Spec{}
}

func coreSpec() *Spec {
	s := newSpec()
	return s
}

func (s *Spec) addBlockRenderer(block ast.Block, renderer blockRenderer) {
	s.blockRenderers[reflect.TypeOf(block)] = renderer
}

func (s *Spec) addInlineRenderer(inline ast.Inline, renderer inlineRenderer) {
	s.inlineRenderers[reflect.TypeOf(inline)] = renderer
}

func (s *Spec) getBlockRenderer(block ast.Block) (blockRenderer, bool) {
	renderer, ok := s.blockRenderers[reflect.TypeOf(block)]
	return renderer, ok
}

func (s *Spec) getInlineRenderer(inline ast.Inline) (inlineRenderer, bool) {
	renderer, ok := s.inlineRenderers[reflect.TypeOf(inline)]
	return renderer, ok
}
