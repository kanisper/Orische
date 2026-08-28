package parser

import (
	"fmt"
	"reflect"
	"strings"

	"orische/internal/parser/feature"
)

// compiledSpec is the parser-private compiled form of a feature.Language.
type compiledSpec struct {
	sugarDefinitions  []feature.BlockSugarDefinition
	blockDefinitions  map[string]feature.BlockDefinition
	inlineDefinitions map[string]feature.InlineDirectiveDefinition
}

func compileSpec(language feature.Language) (*compiledSpec, error) {
	s := &compiledSpec{
		blockDefinitions:  make(map[string]feature.BlockDefinition, len(language.Blocks)+1),
		inlineDefinitions: make(map[string]feature.InlineDirectiveDefinition, len(language.Inlines)),
	}

	if err := s.registerBlock(&paragraphDefinition{}); err != nil {
		return nil, fmt.Errorf("register fixed paragraph definition: %w", err)
	}
	for _, definition := range language.Blocks {
		if err := s.registerBlock(definition); err != nil {
			return nil, err
		}
	}
	for _, definition := range language.Inlines {
		if err := s.registerInline(definition); err != nil {
			return nil, err
		}
	}

	return s, nil
}

func (s *compiledSpec) registerBlock(definition feature.BlockDefinition) error {
	if isNilRegistration(definition) {
		return fmt.Errorf("block definition is nil")
	}

	key := normalizeSyntaxType(definition.BlockType())
	if key == "" {
		return fmt.Errorf("block type must not be empty")
	}
	if _, exists := s.blockDefinitions[key]; exists {
		return fmt.Errorf("block definition %q is already registered", key)
	}

	s.blockDefinitions[key] = definition
	if sugar, ok := definition.(feature.BlockSugarDefinition); ok {
		s.sugarDefinitions = append(s.sugarDefinitions, sugar)
	}
	return nil
}

func (s *compiledSpec) registerInline(definition feature.InlineDirectiveDefinition) error {
	if isNilRegistration(definition) {
		return fmt.Errorf("inline definition is nil")
	}

	key := normalizeSyntaxType(definition.InlineType())
	if key == "" {
		return fmt.Errorf("inline definition type must not be empty")
	}
	policy := definition.ContentPolicy()
	if policy != feature.InlineContentNested && policy != feature.InlineContentLiteral {
		return fmt.Errorf("inline directive definition %q has invalid content policy %d", key, policy)
	}
	if _, exists := s.inlineDefinitions[key]; exists {
		return fmt.Errorf("inline directive definition %q is already registered", key)
	}

	s.inlineDefinitions[key] = definition
	return nil
}

func (s *compiledSpec) getReaders() []feature.BlockReader {
	readers := make([]feature.BlockReader, 0, len(s.sugarDefinitions)+2)

	// Block directives use a fixed envelope reader and take precedence over
	// every registered Sugar Reader.
	readers = append(readers, &blockDirectiveReader{})
	for _, definition := range s.sugarDefinitions {
		readers = append(readers, definition)
	}
	// Paragraph is the mandatory fallback and must remain last.
	readers = append(readers, &paragraphReader{})
	return readers
}

func (s *compiledSpec) getBlockDefinition(blockType string) (feature.BlockDefinition, bool) {
	definition, ok := s.blockDefinitions[normalizeSyntaxType(blockType)]
	return definition, ok
}

func (s *compiledSpec) getInlineDirectiveDefinition(dirtype string) (feature.InlineDirectiveDefinition, bool) {
	definition, ok := s.inlineDefinitions[normalizeSyntaxType(dirtype)]
	return definition, ok
}

func normalizeSyntaxType(syntaxType string) string {
	return strings.ToLower(syntaxType)
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
