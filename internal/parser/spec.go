package parser

import (
	"fmt"
	"reflect"
	"strings"

	"orische/internal/ast"
)

type Spec struct {
	blockDirectiveReader blockReader
	sugarReaders         []blockReader
	paragraphFallback    blockReader
	builders             map[string]blockBuilder
}

type blockReader interface {
	read(ctx *blockContext) (parsedBlockNode, bool, error)
}

type blockBuilder interface {
	build(parser *Parser, block parsedBlockNode) (ast.Block, error)
}

func newSpec() *Spec {
	return &Spec{
		sugarReaders: []blockReader{},
		builders:     map[string]blockBuilder{},
	}
}

func coreSpec() *Spec {
	s := newSpec()

	mustRegister(s.registerBlockDirectiveReader())
	mustRegister(s.registerBlockDirectiveDefinition("code", &codeBlockBuilder{}))
	mustRegister(s.registerBlockSugar("heading", &headingReader{}, &headingBuilder{}))
	mustRegister(s.registerBlockSugar("list", &listReader{}, &listBuilder{}))
	mustRegister(s.registerParagraphFallback(&paragraphBuilder{}))

	return s
}

func (s *Spec) registerBlockDirectiveReader() error {
	if s.blockDirectiveReader != nil {
		return fmt.Errorf("block directive reader is already registered")
	}

	s.blockDirectiveReader = &blockDirectiveReader{}
	return nil
}

func (s *Spec) registerBlockDirectiveDefinition(dirtype string, builder blockBuilder) error {
	return s.registerBlockBuilder(dirtype, builder)
}

func (s *Spec) registerBlockSugar(dirtype string, reader blockReader, builder blockBuilder) error {
	key := normalizeDirectiveType(dirtype)
	if key == "paragraph" {
		return fmt.Errorf("paragraph must be registered as the fallback")
	}
	if isNilRegistration(reader) {
		return fmt.Errorf("block sugar %q has no reader", key)
	}
	if _, ok := reader.(*paragraphReader); ok {
		return fmt.Errorf("paragraph reader must be registered as the fallback")
	}
	if err := s.registerBlockBuilder(key, builder); err != nil {
		return err
	}

	s.sugarReaders = append(s.sugarReaders, reader)
	return nil
}

func (s *Spec) registerParagraphFallback(builder blockBuilder) error {
	if s.paragraphFallback != nil {
		return fmt.Errorf("paragraph fallback is already registered")
	}
	if err := s.registerBlockBuilder("paragraph", builder); err != nil {
		return err
	}

	s.paragraphFallback = &paragraphReader{}
	return nil
}

func (s *Spec) registerBlockBuilder(dirtype string, builder blockBuilder) error {
	key := normalizeDirectiveType(dirtype)
	if key == "" {
		return fmt.Errorf("block builder type must not be empty")
	}
	if isNilRegistration(builder) {
		return fmt.Errorf("block builder %q is nil", key)
	}
	if _, exists := s.builders[key]; exists {
		return fmt.Errorf("block builder %q is already registered", key)
	}

	s.builders[key] = builder
	return nil
}

func (s *Spec) getReaders() []blockReader {
	readers := make([]blockReader, 0, len(s.sugarReaders)+2)
	if s.blockDirectiveReader != nil {
		readers = append(readers, s.blockDirectiveReader)
	}
	readers = append(readers, s.sugarReaders...)
	if s.paragraphFallback != nil {
		readers = append(readers, s.paragraphFallback)
	}
	return readers
}

func (s *Spec) getBuilder(dirtype string) (blockBuilder, bool) {
	builder, ok := s.builders[normalizeDirectiveType(dirtype)]
	return builder, ok
}

func (s *Spec) validate() error {
	if s.blockDirectiveReader == nil {
		return fmt.Errorf("block directive reader is not registered")
	}
	if s.paragraphFallback == nil {
		return fmt.Errorf("paragraph fallback is not registered")
	}
	return nil
}

func normalizeDirectiveType(dirtype string) string {
	return strings.ToLower(dirtype)
}

func isNilRegistration(value any) bool {
	if value == nil {
		return true
	}

	reflected := reflect.ValueOf(value)
	switch reflected.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return reflected.IsNil()
	default:
		return false
	}
}

func mustRegister(err error) {
	if err != nil {
		panic(fmt.Sprintf("register core parser feature: %v", err))
	}
}
