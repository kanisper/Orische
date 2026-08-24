package parser

import (
	"fmt"
	"reflect"
	"strings"

	"orische/internal/ast"
)

// Spec owns the ordered block readers and normalized builder/inline-definition
// registries used by a Parser. Registration is completed before parsing starts.
type Spec struct {
	blockDirectiveReader blockReader
	sugarReaders         []blockSugarReader
	paragraphFallback    blockReader
	builders             map[string]blockBuilder
	inlineDefinitions    map[string]inlineDirectiveDefinition
}

type blockReader interface {
	read(ctx *blockContext) (parsedBlockNode, bool, error)
}

type blockSugarReader interface {
	blockReader
	builderKey() string
}

type blockBuilder interface {
	build(parser *Parser, block parsedBlockNode) (ast.Block, error)
}

func newSpec() *Spec {
	return &Spec{
		builders:          map[string]blockBuilder{},
		inlineDefinitions: map[string]inlineDirectiveDefinition{},
	}
}

// coreSpec assembles the built-in language. Registration failures are
// programmer errors because every entry below is statically defined.
func coreSpec() *Spec {
	s := newSpec()

	mustRegister(s.registerBlockDirectiveReader())
	mustRegister(s.registerBlockDirectiveDefinition(blockBuilderKeyCode, &codeBlockBuilder{}))
	mustRegister(s.registerBlockSugar(&headingReader{}, &headingBuilder{}))
	mustRegister(s.registerBlockSugar(&listReader{}, &listBuilder{}))
	mustRegister(s.registerParagraphFallback(&paragraphBuilder{}))
	mustRegister(s.registerInlineDirectiveDefinition("em", &emphasisInlineDefinition{}))
	mustRegister(s.registerInlineDirectiveDefinition("link", &linkInlineDefinition{}))
	mustRegister(s.registerInlineDirectiveDefinition("code", &codeInlineDefinition{}))

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

// registerBlockSugar installs the builder before mutating reader order, so a
// rejected registration leaves the Spec unchanged.
func (s *Spec) registerBlockSugar(reader blockSugarReader, builder blockBuilder) error {
	if isNilRegistration(reader) {
		return fmt.Errorf("block sugar has no reader")
	}
	if err := s.registerBlockBuilder(reader.builderKey(), builder); err != nil {
		return err
	}

	s.sugarReaders = append(s.sugarReaders, reader)
	return nil
}

// registerParagraphFallback reserves Paragraph for the final reader and
// installs its reader and builder atomically.
func (s *Spec) registerParagraphFallback(builder blockBuilder) error {
	if s.paragraphFallback != nil {
		return fmt.Errorf("paragraph fallback is already registered")
	}
	if err := s.validateBlockBuilder(blockBuilderKeyParagraph, builder); err != nil {
		return err
	}

	s.builders[blockBuilderKeyParagraph] = builder
	s.paragraphFallback = &paragraphReader{}
	return nil
}

func (s *Spec) registerBlockBuilder(dirtype string, builder blockBuilder) error {
	key := normalizeDirectiveType(dirtype)
	if key == blockBuilderKeyParagraph {
		return fmt.Errorf("paragraph must be registered as the fallback")
	}
	if err := s.validateBlockBuilder(key, builder); err != nil {
		return err
	}

	s.builders[key] = builder
	return nil
}

func (s *Spec) validateBlockBuilder(key string, builder blockBuilder) error {
	if key == "" {
		return fmt.Errorf("block builder type must not be empty")
	}
	if isNilRegistration(builder) {
		return fmt.Errorf("block builder %q is nil", key)
	}
	if _, exists := s.builders[key]; exists {
		return fmt.Errorf("block builder %q is already registered", key)
	}

	return nil
}

// getReaders materializes the fixed precedence: shared directive reader,
// registered sugar readers, then the mandatory Paragraph fallback.
func (s *Spec) getReaders() []blockReader {
	readers := make([]blockReader, 0, len(s.sugarReaders)+2)
	if s.blockDirectiveReader != nil {
		readers = append(readers, s.blockDirectiveReader)
	}
	for _, reader := range s.sugarReaders {
		readers = append(readers, reader)
	}
	if s.paragraphFallback != nil {
		readers = append(readers, s.paragraphFallback)
	}
	return readers
}

func (s *Spec) getBuilder(dirtype string) (blockBuilder, bool) {
	builder, ok := s.builders[normalizeDirectiveType(dirtype)]
	return builder, ok
}

func (s *Spec) registerInlineDirectiveDefinition(dirtype string, definition inlineDirectiveDefinition) error {
	key := normalizeDirectiveType(dirtype)
	if key == "" {
		return fmt.Errorf("inline directive type must not be empty")
	}
	if isNilRegistration(definition) {
		return fmt.Errorf("inline directive definition %q is nil", key)
	}
	policy := definition.contentPolicy()
	if policy != inlineContentNested && policy != inlineContentLiteral {
		return fmt.Errorf("inline directive definition %q has invalid content policy %d", key, policy)
	}
	if _, exists := s.inlineDefinitions[key]; exists {
		return fmt.Errorf("inline directive definition %q is already registered", key)
	}

	s.inlineDefinitions[key] = definition
	return nil
}

func (s *Spec) getInlineDirectiveDefinition(dirtype string) (inlineDirectiveDefinition, bool) {
	definition, ok := s.inlineDefinitions[normalizeDirectiveType(dirtype)]
	return definition, ok
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

// isNilRegistration catches typed nil pointers stored in interface values.
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
