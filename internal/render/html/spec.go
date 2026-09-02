package html

import (
	"fmt"
	"io"
	"reflect"

	"orische/internal/ast"
)

type spec struct {
	blockRenderers  map[reflect.Type]blockRenderFunc
	inlineRenderers map[reflect.Type]inlineRenderFunc
}

type blockRenderFunc func(*Renderer, io.Writer, ast.Block) error
type inlineRenderFunc func(*Renderer, io.Writer, ast.Inline) error

func newSpec() *spec {
	return &spec{
		blockRenderers:  make(map[reflect.Type]blockRenderFunc),
		inlineRenderers: make(map[reflect.Type]inlineRenderFunc),
	}
}

func coreSpec() *spec {
	s := newSpec()

	addBlockRenderer(s, renderHeading)
	addBlockRenderer(s, renderCodeBlock)
	addBlockRenderer(s, renderList)
	addBlockRenderer(s, renderParagraph)

	addInlineRenderer(s, renderText)
	addInlineRenderer(s, renderLineBreak)
	addInlineRenderer(s, renderEmphasis)
	addInlineRenderer(s, renderStrong)
	addInlineRenderer(s, renderItalic)
	addInlineRenderer(s, renderBold)
	addInlineRenderer(s, renderDeleted)
	addInlineRenderer(s, renderOutdated)
	addInlineRenderer(s, renderCodeSpan)
	addInlineRenderer(s, renderLink)

	return s
}

func addBlockRenderer[T ast.Block](s *spec, render func(*Renderer, io.Writer, T) error) {
	s.blockRenderers[reflect.TypeFor[T]()] = func(r *Renderer, w io.Writer, block ast.Block) error {
		typed, ok := block.(T)
		if !ok {
			return fmt.Errorf("html renderer: block renderer expected %T, but got %T", *new(T), block)
		}
		return render(r, w, typed)
	}
}

func addInlineRenderer[T ast.Inline](s *spec, render func(*Renderer, io.Writer, T) error) {
	s.inlineRenderers[reflect.TypeFor[T]()] = func(r *Renderer, w io.Writer, inline ast.Inline) error {
		typed, ok := inline.(T)
		if !ok {
			return fmt.Errorf("html renderer: inline renderer expected %T, but got %T", *new(T), inline)
		}
		return render(r, w, typed)
	}
}

func (s *spec) getBlockRenderer(block ast.Block) (blockRenderFunc, bool) {
	renderer, ok := s.blockRenderers[reflect.TypeOf(block)]
	return renderer, ok
}

func (s *spec) getInlineRenderer(inline ast.Inline) (inlineRenderFunc, bool) {
	renderer, ok := s.inlineRenderers[reflect.TypeOf(inline)]
	return renderer, ok
}
