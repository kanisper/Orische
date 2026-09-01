package parser

import (
	"errors"
	"fmt"
	"strings"

	"orische/internal/ast"
	"orische/internal/diagnostic"
)

// Parser converts Orische source to AST using one parser spec. Use NewParser
// to construct one; the zero value returns an initialization error.
type Parser struct {
	spec *spec
}

// NewParser constructs a parser with the built-in syntax definitions.
func NewParser() *Parser {
	return newParser()
}

// Parse parses input with the built-in parser definitions.
func Parse(input string) (*ast.Document, error) {
	return NewParser().Parse(input)
}

// Parse parses input with the Parser's spec.
func (p *Parser) Parse(input string) (*ast.Document, error) {
	if p == nil || p.spec == nil {
		return nil, errors.New("parser is not initialized; use NewParser")
	}
	parsed, err := p.parseDocument(splitLines(input))
	if err != nil {
		return nil, err
	}
	return p.buildDocument(parsed)
}

type parsedDocument struct {
	blocks []parsedBlock
	rng    ast.Range
}

func (p *Parser) parseDocument(lines []string) (*parsedDocument, error) {
	blocks, end, err := p.parseBlocks(lines)
	if err != nil {
		return nil, err
	}

	return &parsedDocument{
		blocks: blocks,
		rng: ast.Range{
			Start: ast.Position{Line: 1, Column: 1},
			End:   end,
		},
	}, nil
}

func (p *Parser) parseBlocks(lines []string) ([]parsedBlock, ast.Position, error) {
	var blocks []parsedBlock
	var documentEnd ast.Position

	for pos := 0; pos < len(lines); {
		if strings.TrimSpace(lines[pos]) == "" {
			pos++
			continue
		}

		node, consumed, err := p.parseOneBlock(&blockContext{lines: lines, start: pos})
		if err != nil {
			return nil, ast.Position{}, err
		}
		documentEnd = node.blockRange().End
		blocks = append(blocks, node)
		pos += consumed
	}

	return blocks, documentEnd, nil
}

func (p *Parser) parseOneBlock(input *blockContext) (parsedBlock, int, error) {
	// Block precedence is part of the grammar: the fixed directive envelope,
	// ordered built-in sugar, and the fixed paragraph fallback.
	if node, consumed := readBlockDirective(input); node != nil {
		return node, consumed, nil
	}
	for _, sugar := range p.spec.sugars {
		if node, consumed := sugar(input); node != nil {
			return node, consumed, nil
		}
	}
	if node, consumed := readParagraph(input); node != nil {
		return node, consumed, nil
	}
	return nil, 0, fmt.Errorf("block parser could not read input")
}

func (p *Parser) buildDocument(parsed *parsedDocument) (*ast.Document, error) {
	doc := &ast.Document{
		Blocks: make([]ast.Block, 0, len(parsed.blocks)),
		Range:  parsed.rng,
	}

	for _, node := range parsed.blocks {
		block, err := p.buildBlock(node)
		if err != nil {
			return nil, err
		}
		doc.Blocks = append(doc.Blocks, block)
	}
	return doc, nil
}

func (p *Parser) buildBlock(node parsedBlock) (ast.Block, error) {
	if node == nil {
		return nil, fmt.Errorf("build block: node is nil")
	}

	switch node := node.(type) {
	case *blockDirectiveNode:
		blockType := normalizeSyntaxType(node.dirtype)
		builder, ok := p.spec.directives[blockType]
		if !ok {
			return nil, &diagnostic.Error{
				Message: fmt.Sprintf("unsupported block directive type %q", blockType),
				Range:   node.rng,
			}
		}
		block, err := builder(p, node)
		return finishBlockBuild(blockType, block, err)

	case *headingNode:
		block, err := p.buildHeading(node)
		return finishBlockBuild(typeHeading, block, err)

	case *listNode:
		block, err := p.buildList(node)
		return finishBlockBuild(typeList, block, err)

	case *paragraphNode:
		block, err := p.buildParagraph(node)
		return finishBlockBuild(typeParagraph, block, err)

	default:
		return nil, fmt.Errorf("build block: unknown node %T", node)
	}
}

func finishBlockBuild(blockType string, block ast.Block, err error) (ast.Block, error) {
	if err == nil {
		return block, nil
	}
	var diag *diagnostic.Error
	if errors.As(err, &diag) {
		return nil, err
	}
	return nil, fmt.Errorf("build %q block: %w", blockType, err)
}

func splitLines(input string) []string {
	input = strings.NewReplacer("\r\n", "\n", "\r", "\n").Replace(input)
	input = strings.TrimRight(input, "\n")
	if input == "" {
		return nil
	}
	return strings.Split(input, "\n")
}
