package parser

import "orische/internal/ast"

type Spec struct {
	blockReaders []blockReader
	fallback     blockReader
	builders     map[string]blockBuilder
}

type blockReader interface {
	read(ctx *blockContext) (parsedBlockNode, bool, error)
}

type blockBuilder interface {
	build(parser *Parser, block parsedBlockNode) (ast.Block, error)
}

func newSpec() *Spec {
	return &Spec{
		blockReaders: []blockReader{},
		fallback:     &paragraphReader{},
		builders:     map[string]blockBuilder{},
	}
}

func coreSpec() *Spec {
	s := newSpec()

	s.addBlockReader(&blockDirectiveReader{})
	s.addBlockReader(&headingReader{})
	s.addBlockReader(&listReader{})

	s.addBlockBuilder("heading", &headingBuilder{})
	s.addBlockBuilder("code", &codeBlockBuilder{})
	s.addBlockBuilder("paragraph", &paragraphBuilder{})
	s.addBlockBuilder("list", &listBuilder{})

	return s
}

func (s *Spec) addBlockReader(reader blockReader) {
	s.blockReaders = append(s.blockReaders, reader)
}
func (s *Spec) addBlockBuilder(dirtype string, b blockBuilder) {
	s.builders[dirtype] = b
}

func (s *Spec) getReaders() []blockReader {
	readers := make([]blockReader, 0, len(s.blockReaders)+1)
	readers = append(readers, s.blockReaders...)
	readers = append(readers, s.fallback)
	return readers
}

func (s *Spec) getBuilder(dirtype string) (blockBuilder, bool) {
	builder, ok := s.builders[dirtype]
	return builder, ok
}
