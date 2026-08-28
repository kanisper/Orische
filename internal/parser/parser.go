package parser

import (
	"errors"
	"fmt"
	"strings"

	"orische/internal/ast"
	"orische/internal/diagnostic"
	"orische/internal/parser/feature"
	"orische/internal/parser/syntax"
)

// Parser converts Orische source to AST using one compiled language spec. Use
// NewParser to construct one; the zero value returns an initialization error.
type Parser struct {
	spec *compiledSpec
}

// NewParser validates and compiles language before parsing starts.
func NewParser(language feature.Language) (*Parser, error) {
	spec, err := compileSpec(language)
	if err != nil {
		return nil, fmt.Errorf("invalid parser language: %w", err)
	}
	return &Parser{spec: spec}, nil
}

// Parse parses input with the built-in language.
func Parse(input string) (*ast.Document, error) {
	p, err := NewParser(syntax.Core())
	if err != nil {
		panic(fmt.Sprintf("compile built-in parser language: %v", err))
	}
	return p.Parse(input)
}

// Parse parses input with the Parser's compiled language.
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
	blocks []feature.BlockNode
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

func (p *Parser) parseBlocks(lines []string) ([]feature.BlockNode, ast.Position, error) {
	var blocks []feature.BlockNode
	var documentEnd ast.Position

	for pos := 0; pos < len(lines); {
		if strings.TrimSpace(lines[pos]) == "" {
			pos++
			continue
		}

		node, consumed, err := p.parseOneBlock(&blockInput{lines: lines, start: pos})
		if err != nil {
			return nil, ast.Position{}, err
		}
		documentEnd = node.BlockRange().End
		blocks = append(blocks, node)
		pos += consumed
	}

	return blocks, documentEnd, nil
}

func (p *Parser) parseOneBlock(input *blockInput) (feature.BlockNode, int, error) {
	for _, reader := range p.spec.getReaders() {
		result, err := reader.ReadBlock(input)
		if err != nil {
			return nil, 0, err
		}
		if err := validateBlockReadResult(reader, input.Len(), result); err != nil {
			return nil, 0, err
		}
		if !result.Matched {
			continue
		}

		if sugar, ok := reader.(feature.BlockSugarDefinition); ok {
			declaredType := normalizeSyntaxType(sugar.BlockType())
			actualType := normalizeSyntaxType(result.Node.BlockType())
			if declaredType != actualType {
				return nil, 0, fmt.Errorf(
					"block sugar definition %T declared block type %q but produced %q",
					reader,
					declaredType,
					actualType,
				)
			}
		}

		return result.Node, result.Consumed, nil
	}

	panic("unreachable: paragraph fallback reader must always succeed")
}

func validateBlockReadResult(
	reader feature.BlockReader,
	available int,
	result feature.BlockReadResult,
) error {
	if !result.Matched {
		if result.Consumed != 0 {
			return fmt.Errorf("block reader %T: unmatched block reader consumed %d lines", reader, result.Consumed)
		}
		if result.Node != nil {
			return fmt.Errorf("block reader %T: unmatched block reader returned a node", reader)
		}
		return nil
	}

	if result.Consumed <= 0 {
		return fmt.Errorf("block reader %T: matched block reader consumed %d lines", reader, result.Consumed)
	}
	if result.Consumed > available {
		return fmt.Errorf(
			"block reader %T: matched block reader consumed %d lines with only %d available",
			reader,
			result.Consumed,
			available,
		)
	}
	if isNilRegistration(result.Node) {
		return fmt.Errorf("block reader %T: matched block reader returned no node", reader)
	}

	return nil
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

func (p *Parser) buildBlock(node feature.BlockNode) (ast.Block, error) {
	if isNilRegistration(node) {
		return nil, fmt.Errorf("build block: node is nil")
	}
	rawType := node.BlockType()
	blockType := normalizeSyntaxType(rawType)
	definition, ok := p.spec.getBlockDefinition(rawType)
	if !ok {
		return nil, &diagnostic.Error{
			Message: fmt.Sprintf("unsupported block directive type %q", blockType),
			Range:   node.BlockRange(),
		}
	}

	block, err := definition.BuildBlock(buildContext{parser: p}, node)
	if err != nil {
		var diag *diagnostic.Error
		if errors.As(err, &diag) {
			return nil, err
		}
		return nil, fmt.Errorf("build %q block: %w", blockType, err)
	}
	if isNilRegistration(block) {
		return nil, fmt.Errorf("build %q block: definition returned a nil block", blockType)
	}

	return block, nil
}

type buildContext struct {
	parser *Parser
}

func (c buildContext) ParseInlines(text string, origin ast.Position) ([]ast.Inline, error) {
	return c.parser.parseInlines(text, origin)
}

func (c buildContext) BuildBlock(node feature.BlockNode) (ast.Block, error) {
	return c.parser.buildBlock(node)
}

func splitLines(input string) []string {
	input = strings.TrimRight(input, "\n")
	if input == "" {
		return nil
	}
	return strings.Split(input, "\n")
}
