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
	return &Spec{}
}

func coreSpec() *Spec {
	s := newSpec()
	s.addBlockParser(&blockDirectiveParser{})
	return s
}

func (s *Spec) addBlockParser(b blockParser) {
	s.blockParsers = append(s.blockParsers, b)
}
func (s *Spec) addBlockBuilder(dirtype string, b blockBuilder) {
	s.builders[dirtype] = b
}

func (s *Spec) getParsers() []blockParser {
	return s.blockParsers
}
func (s *Spec) getBuilder(dirtype string) blockBuilder {
	return s.builders[dirtype]
}
