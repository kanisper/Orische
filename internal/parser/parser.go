package parser

import (
	"errors"
	"fmt"
	"strings"

	"orische/internal/ast"
	"orische/internal/diagnostic"
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

	blocks, endDocPosition, err := p.parseBlocks(ctx)
	if err != nil {
		return nil, err
	}

	doc := &parsedDocument{
		Blocks: blocks,
		Range: ast.Range{
			Start: ast.Position{Line: 1, Column: 1},
			End:   endDocPosition,
		},
	}

	return doc, nil
}

func (p *Parser) parseBlocks(ctx *blockContext) ([]parsedBlockNode, ast.Position, error) {
	var blocks []parsedBlockNode
	var docEndPosition ast.Position

	for !ctx.isEOF() {
		if strings.TrimSpace(ctx.getLine()) == "" {
			ctx.advance(1)
			continue
		}

		block, ok, err := p.parseOneBlock(ctx)
		if err != nil {
			return nil, ast.Position{}, err
		}
		if ok {
			docEndPosition = block.getBlockRange().End
			blocks = append(blocks, block)
			ctx.advance(1)
			continue
		}

		panic("unreachable: paragraph fallback reader must always succeed")
	}
	return blocks, docEndPosition, nil
}

func (p *Parser) parseOneBlock(ctx *blockContext) (parsedBlockNode, bool, error) {
	for _, reader := range p.spec.getReaders() {
		block, ok, err := reader.read(ctx)
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
			return nil, &diagnostic.Error{
				Message: fmt.Sprintf("unsupported block directive type %q", node.getBuilderKey()),
				Range:   node.getBlockRange(),
			}
		}

		builtBlock, err := builder.build(p, node)
		if err != nil {
			var diag *diagnostic.Error
			if errors.As(err, &diag) {
				return nil, err
			}

			return nil, fmt.Errorf(
				"build %q block: %w",
				node.getBuilderKey(),
				err,
			)
		}

		doc.Blocks = append(doc.Blocks, builtBlock)
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
