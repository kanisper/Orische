package parser

import (
	"medoc/internal/ast"
	"strings"
)

type Parser struct {
	spec *Spec
}

func NewParser(spec *Spec) *Parser {
	if spec == nil {
		spec = coreSpec()
	}
	return &Parser{
		spec: spec,
	}
}

func Parse(input string) (*ast.Document, error) {
	return NewParser(coreSpec()).Parse(input)
}

func (p *Parser) Parse(input string) (*ast.Document, error) {
	lines := splitLines(input)

	_, err := p.parseDocument(lines)
	//parsed, err := p.parseDocument(lines)
	if err != nil {
		return nil, err
	}

	return &ast.Document{}, err
	// return p.buildDocument(parsed)
}

func (p *Parser) parseDocument(lines []string) (*parsedDocument, error) {
	ctx := newBlockContext(lines, 0)

	blocks, err := p.parseBlocks(ctx)
	if err != nil {
		return nil, err
	}

	doc := &parsedDocument{
		Blocks: blocks,
		Range: ast.Range{
			StartLine: 0,
			EndLine:   ctx.getPos() - 1,
		},
	}

	return doc, nil
}

func (p *Parser) parseBlocks(ctx *blockContext) ([]parsedBlockNode, error) {
	var blocks []parsedBlockNode

	for !ctx.isEOF() {
		if strings.TrimSpace(ctx.getLine()) == "" {
			ctx.advance(1)
			continue
		}

		block, ok, err := p.parseOneBlock(ctx)
		if err != nil {
			return nil, err
		}
		if ok {
			blocks = append(blocks, block)
			ctx.advance(1)
			continue
		}

		panic("unreachable: paragraph fallback must always parse")
	}
	return blocks, nil
}

func (p *Parser) parseOneBlock(ctx *blockContext) (parsedBlockNode, bool, error) {
	for _, bp := range p.spec.getParsers() {
		block, ok, err := bp.parse(ctx)
		if err != nil {
			return nil, false, err
		}
		if ok {
			return block, true, nil
		}
	}
	return nil, false, nil
}

// TODO: func (*Parser) buildDocument(parsed *parsedDocument) (*ast.Document, error)

func splitLines(input string) []string {
	if input == "" {
		return nil
	}
	lines := strings.Split(input, "\n")
	if lines[len(lines)-1] == "" {
		return lines[:len(lines)-1]
	}
	return lines
}
