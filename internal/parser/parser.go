package parser

import (
	"errors"
	"fmt"
	"strings"

	"orische/internal/ast"
	"orische/internal/diagnostic"
)

// Parser converts Orische source to AST using one active Spec.
type Parser struct {
	spec *Spec
}

// NewParser creates a Parser with spec, or the built-in language when spec is nil.
func NewParser(spec *Spec) *Parser {
	if spec == nil {
		spec = coreSpec()
	}
	return &Parser{spec: spec}
}

// Parse parses input with the built-in language.
func Parse(input string) (*ast.Document, error) {
	return NewParser(nil).Parse(input)
}

// Parse parses input with the Parser's active Spec.
func (p *Parser) Parse(input string) (*ast.Document, error) {
	if err := p.spec.validate(); err != nil {
		return nil, fmt.Errorf("invalid parser spec: %w", err)
	}

	parsed, err := p.parseDocument(splitLines(input))
	if err != nil {
		return nil, err
	}

	return p.buildDocument(parsed)
}

func (p *Parser) parseDocument(lines []string) (*parsedDocument, error) {
	ctx := &blockContext{lines: lines}

	blocks, endDocPosition, err := p.parseBlocks(ctx)
	if err != nil {
		return nil, err
	}

	return &parsedDocument{
		Blocks: blocks,
		Range: ast.Range{
			Start: ast.Position{Line: 1, Column: 1},
			End:   endDocPosition,
		},
	}, nil
}

func (p *Parser) parseBlocks(ctx *blockContext) ([]parsedBlockNode, ast.Position, error) {
	var blocks []parsedBlockNode
	var docEndPosition ast.Position

	for !ctx.isEOF() {
		if strings.TrimSpace(ctx.line()) == "" {
			ctx.pos++
			continue
		}

		block, err := p.parseOneBlock(ctx)
		if err != nil {
			return nil, ast.Position{}, err
		}
		docEndPosition = block.getBlockRange().End
		blocks = append(blocks, block)
		ctx.pos++
	}
	return blocks, docEndPosition, nil
}

func (p *Parser) parseOneBlock(ctx *blockContext) (parsedBlockNode, error) {
	for _, reader := range p.spec.getReaders() {
		block, ok, err := reader.read(ctx)
		if err != nil {
			return nil, err
		}
		if ok {
			if block == nil {
				return nil, fmt.Errorf("block reader %T succeeded with a nil parsed block", reader)
			}
			if sugarDefinition, isSugar := reader.(blockSugarDefinition); isSugar {
				declaredType := normalizeSyntaxType(sugarDefinition.blockType())
				actualType := normalizeSyntaxType(block.blockType())
				if declaredType != actualType {
					return nil, fmt.Errorf(
						"block sugar definition %T declared block type %q but produced %q",
						reader,
						declaredType,
						actualType,
					)
				}
			}
			return block, nil
		}
	}
	// Spec validation guarantees that the final Paragraph reader accepts every
	// nonblank line passed here.
	panic("unreachable: paragraph fallback reader must always succeed")
}

func (p *Parser) buildDocument(parsedDoc *parsedDocument) (*ast.Document, error) {
	doc := &ast.Document{
		Blocks: make([]ast.Block, 0, len(parsedDoc.Blocks)),
		Range:  parsedDoc.Range,
	}

	for _, node := range parsedDoc.Blocks {
		block, err := p.buildBlock(node)
		if err != nil {
			return nil, err
		}

		doc.Blocks = append(doc.Blocks, block)
	}

	return doc, nil
}

func (p *Parser) buildBlock(node parsedBlockNode) (ast.Block, error) {
	rawType := node.blockType()
	blockType := normalizeSyntaxType(rawType)
	definition, ok := p.spec.getBlockDefinition(rawType)
	if !ok {
		return nil, &diagnostic.Error{
			Message: fmt.Sprintf("unsupported block directive type %q", blockType),
			Range:   node.getBlockRange(),
		}
	}

	block, err := definition.build(p, node)
	if err != nil {
		var diag *diagnostic.Error
		if errors.As(err, &diag) {
			return nil, err
		}

		return nil, fmt.Errorf(
			"build %q block: %w",
			blockType,
			err,
		)
	}

	return block, nil
}

func splitLines(input string) []string {
	input = strings.TrimRight(input, "\n")
	if input == "" {
		return nil
	}
	return strings.Split(input, "\n")
}
