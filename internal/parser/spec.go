package parser

type Spec struct {
	blockParsers []blockParser
	fallback     blockParser
	builders     map[string]blockBuilder
}

type blockParser interface {
	parse(ctx *blockContext) (parsedBlockNode, bool, error)
}

func newSpec() *Spec {
	return &Spec{
		blockParsers: []blockParser{},
		fallback:     &paragraphParser{},
		builders:     map[string]blockBuilder{},
	}
}

func coreSpec() *Spec {
	s := newSpec()
	s.addBlockParser(&blockDirectiveParser{})
	s.addBlockParser(&headingParser{})
	s.addBlockParser(&listParser{})
	return s
}

func (s *Spec) addBlockParser(b blockParser) {
	s.blockParsers = append(s.blockParsers, b)
}
func (s *Spec) addBlockBuilder(dirtype string, b blockBuilder) {
	s.builders[dirtype] = b
}

func (s *Spec) getParsers() []blockParser {
	parsers := make([]blockParser, 0, len(s.blockParsers)+1)
	parsers = append(parsers, s.blockParsers...)
	parsers = append(parsers, s.fallback)
	return parsers
}

func (s *Spec) getBuilder(dirtype string) blockBuilder {
	return s.builders[dirtype]
}
