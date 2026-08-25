package parser

import (
	"fmt"
	"reflect"
	"strings"

	"orische/internal/ast"
)

// Spec owns the ordered block readers and normalized block/inline-definition
// registries used by a Parser. Registration is completed before parsing starts.
type Spec struct {
	sugarDefinitions  []blockSugarDefinition
	blockDefinitions  map[string]blockDefinition
	inlineDefinitions map[string]inlineDefinition
}

type blockReader interface {
	read(ctx *blockContext) (parsedBlockNode, bool, error)
}

type blockBuilder interface {
	build(parser *Parser, block parsedBlockNode) (ast.Block, error)
}

type blockDefinition interface {
	blockBuilder
	blockType() string
}

type blockSugarDefinition interface {
	blockDefinition
	blockReader
}

type inlineDefinition interface {
	inlineType() string
}

func newSpec() *Spec {
	paragraph := &paragraphDefinition{}
	return &Spec{
		blockDefinitions: map[string]blockDefinition{
			paragraph.blockType(): paragraph,
		},
		inlineDefinitions: map[string]inlineDefinition{},
	}
}

// coreSpec assembles the built-in language. Registration failures are
// programmer errors because every entry below is statically defined.
func coreSpec() *Spec {
	s := newSpec()

	mustRegister(s.registerBlock(&codeBlockDefinition{}))
	mustRegister(s.registerBlock(&headingDefinition{}))
	mustRegister(s.registerBlock(&listDefinition{}))
	mustRegister(s.registerInline(&emphasisInlineDefinition{}))
	mustRegister(s.registerInline(&linkInlineDefinition{}))
	mustRegister(s.registerInline(&codeInlineDefinition{}))

	return s
}

// registerBlock adds reader-capable definitions to the ordered sugar chain.
// Definitions without a reader use the common Block Directive envelope.
func (s *Spec) registerBlock(definition blockDefinition) error {
	if err := s.registerBlockDefinition(definition); err != nil {
		return err
	}

	if sugar, ok := definition.(blockSugarDefinition); ok {
		s.sugarDefinitions = append(s.sugarDefinitions, sugar)
	}
	return nil
}

func (s *Spec) registerBlockDefinition(definition blockDefinition) error {
	if isNilRegistration(definition) {
		return fmt.Errorf("block definition is nil")
	}
	key := normalizeSyntaxType(definition.blockType())
	if key == blockTypeParagraph {
		return fmt.Errorf("paragraph is fixed parser infrastructure")
	}
	if err := s.validateBlockDefinition(key, definition); err != nil {
		return err
	}

	s.blockDefinitions[key] = definition
	return nil
}

func (s *Spec) validateBlockDefinition(key string, definition blockDefinition) error {
	if key == "" {
		return fmt.Errorf("block type must not be empty")
	}
	if isNilRegistration(definition) {
		return fmt.Errorf("block definition %q is nil", key)
	}
	if _, exists := s.blockDefinitions[key]; exists {
		return fmt.Errorf("block definition %q is already registered", key)
	}

	return nil
}

// getReaders materializes the fixed precedence: shared directive reader,
// registered sugar definitions, then the fixed Paragraph fallback.
func (s *Spec) getReaders() []blockReader {
	readers := make([]blockReader, 0, len(s.sugarDefinitions)+2)
	// blockDirectiveReader must be first to intercept all block reads.
	readers = append(readers, &blockDirectiveReader{})
	for _, definition := range s.sugarDefinitions {
		readers = append(readers, definition)
	}
	// paragraphDefinition must be last to match all remaining blocks.
	readers = append(readers, &paragraphDefinition{})
	return readers
}

func (s *Spec) getBlockDefinition(blockType string) (blockDefinition, bool) {
	definition, ok := s.blockDefinitions[normalizeSyntaxType(blockType)]
	return definition, ok
}

func (s *Spec) registerInline(definition inlineDefinition) error {
	if isNilRegistration(definition) {
		return fmt.Errorf("inline definition is nil")
	}
	key := normalizeSyntaxType(definition.inlineType())
	if key == "" {
		return fmt.Errorf("inline definition type must not be empty")
	}
	directive, ok := definition.(inlineDirectiveDefinition)
	if !ok {
		return fmt.Errorf("inline definition %q has no parser contract", key)
	}
	policy := directive.contentPolicy()
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
	definition, ok := s.inlineDefinitions[normalizeSyntaxType(dirtype)]
	if !ok {
		return nil, false
	}
	directive, ok := definition.(inlineDirectiveDefinition)
	return directive, ok
}

func (s *Spec) validate() error {
	if _, ok := s.getBlockDefinition(blockTypeParagraph); !ok {
		return fmt.Errorf("paragraph definition is not installed")
	}
	return nil
}

func normalizeSyntaxType(syntaxType string) string {
	return strings.ToLower(syntaxType)
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
