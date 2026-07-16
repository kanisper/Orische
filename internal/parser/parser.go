package parser

import (
	"fmt"
	"strings"

	"orische/internal/ast"
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

	parsed, err := p.parseDocument(lines)
	if err != nil {
		return nil, err
	}

	return p.buildDocument(parsed)
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
			StartLine: 1,
			EndLine:   ctx.getPos(),
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

func (p *Parser) buildDocument(parsedDoc *parsedDocument) (*ast.Document, error) {
	doc := &ast.Document{
		Blocks: make([]ast.Block, 0),
		Range:  parsedDoc.Range,
	}

	for _, node := range parsedDoc.Blocks {
		builder, ok := p.spec.getBuilder(node.getBuilderKey())
		if !ok {
			return nil, fmt.Errorf(
				"build document: builder not found for parsed block %T with key %q",
				node,
				node.getBuilderKey(),
			)
		}

		buildedBlock, err := builder.build(node)
		if err != nil {
			return nil, fmt.Errorf(
				"build document: build failed %q block at lines %d-%d: %w",
				node.getBuilderKey(),
				node.getBlockRange().StartLine,
				node.getBlockRange().EndLine,
				err,
			)
		}

		doc.Blocks = append(doc.Blocks, buildedBlock)
	}

	return doc, nil
}

func splitLines(input string) []string {
	lines := strings.Split(input, "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		if lines[i] != "" {
			return lines[:i+1]
		}
	}
	return nil
}
