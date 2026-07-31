package html

import (
	"fmt"
	"io"
	"reflect"

	"orische/internal/ast"
)

type Spec struct {
	blockRenderers  map[reflect.Type]blockRenderer[ast.Block]
	inlineRenderers map[reflect.Type]inlineRenderer[ast.Inline]
}

type blockRenderer[T ast.Block] interface {
	render(r *Renderer, w io.Writer, block T) error
}

type inlineRenderer[T ast.Inline] interface {
	render(r *Renderer, w io.Writer, inline T) error
}

type blockRendererAdapter[T ast.Block] struct {
	renderer blockRenderer[T]
}

type inlineRendererAdapter[T ast.Inline] struct {
	renderer inlineRenderer[T]
}

func (a blockRendererAdapter[T]) render(r *Renderer, w io.Writer, block ast.Block) error {
	typed, ok := block.(T)
	if !ok {
		return fmt.Errorf("html renderer: block renderer expected %T, but got %T", *new(T), block)
	}
	return a.renderer.render(r, w, typed)
}

func (a inlineRendererAdapter[T]) render(r *Renderer, w io.Writer, inline ast.Inline) error {
	typed, ok := inline.(T)
	if !ok {
		return fmt.Errorf("html renderer: inline renderer expected %T, but got %T", *new(T), inline)
	}
	return a.renderer.render(r, w, typed)
}

func newSpec() *Spec {
	return &Spec{
		blockRenderers:  make(map[reflect.Type]blockRenderer[ast.Block]),
		inlineRenderers: make(map[reflect.Type]inlineRenderer[ast.Inline]),
	}
}

func coreSpec() *Spec {
	s := newSpec()

	addBlockRenderer(s, &headingRenderer{})
	addBlockRenderer(s, &codeblockRenderer{})
	addBlockRenderer(s, &listRenderer{})
	addBlockRenderer(s, &paragraphRenderer{})

	addInlineRenderer(s, &textRenderer{})
	addInlineRenderer(s, &emphasisRenderer{})
	addInlineRenderer(s, &codespanRenderer{})
	addInlineRenderer(s, &linkRenderer{})

	return s
}

func addBlockRenderer[T ast.Block](s *Spec, renderer blockRenderer[T]) {
	s.blockRenderers[reflect.TypeFor[T]()] = blockRendererAdapter[T]{renderer: renderer}
}

func addInlineRenderer[T ast.Inline](s *Spec, renderer inlineRenderer[T]) {
	s.inlineRenderers[reflect.TypeFor[T]()] = inlineRendererAdapter[T]{renderer: renderer}
}

func (s *Spec) getBlockRenderer(block ast.Block) (blockRenderer[ast.Block], bool) {
	renderer, ok := s.blockRenderers[reflect.TypeOf(block)]
	return renderer, ok
}

func (s *Spec) getInlineRenderer(inline ast.Inline) (inlineRenderer[ast.Inline], bool) {
	renderer, ok := s.inlineRenderers[reflect.TypeOf(inline)]
	return renderer, ok
}
