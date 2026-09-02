package parser

import "strings"

// spec is the small parser configuration that still provides useful value:
// directive semantics, sugar precedence, and inline definitions.
type spec struct {
	directives        map[string]blockDirectiveBuilder
	sugars            []blockSugar
	inlineReaders     []inlineReader
	inlineDefinitions map[string]inlineDefinition
}

func newSpec() *spec {
	s := &spec{
		directives: map[string]blockDirectiveBuilder{
			typeCodeBlock: buildCodeBlock,
			typeHeading:   buildHeadingDirective,
			typeParagraph: buildParagraphDirective,
		},
		sugars: []blockSugar{
			readHeading,
			readList,
		},
		// Reader order is inline grammar precedence.
		inlineReaders: []inlineReader{
			readInlineEscape,
			readInlineDirective,
			readStrongSugar,
			readEmphasisSugar,
			readBoldSugar,
			readItalicSugar,
			readUnderlineSugar,
			readStrikethroughSugar,
			readCodeSpanSugar,
			readLinkSugar,
			readLineBreak,
		},
		inlineDefinitions: make(map[string]inlineDefinition, 8),
	}

	for typ, definition := range coreInlineDefinitions() {
		s.inlineDefinitions[normalizeSyntaxType(typ)] = definition
	}

	return s
}

func (s *spec) getInlineDirectiveDefinition(dirtype string) (inlineDefinition, bool) {
	definition, ok := s.inlineDefinitions[normalizeSyntaxType(dirtype)]
	return definition, ok
}

func normalizeSyntaxType(syntaxType string) string {
	return strings.ToLower(syntaxType)
}
