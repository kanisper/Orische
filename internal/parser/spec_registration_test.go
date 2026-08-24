package parser

import (
	"errors"
	"strings"
	"testing"

	"orische/internal/ast"
	"orische/internal/diagnostic"
)

func TestCoreSpec_BlockReaderOrder(t *testing.T) {
	readers := coreSpec().getReaders()
	if len(readers) != 4 {
		t.Fatalf("core reader count = %d, want 4", len(readers))
	}

	if _, ok := readers[0].(*blockDirectiveReader); !ok {
		t.Errorf("reader 0 type = %T, want *blockDirectiveReader", readers[0])
	}
	if _, ok := readers[1].(*headingDefinition); !ok {
		t.Errorf("reader 1 type = %T, want *headingDefinition", readers[1])
	}
	if _, ok := readers[2].(*listDefinition); !ok {
		t.Errorf("reader 2 type = %T, want *listDefinition", readers[2])
	}
	if _, ok := readers[3].(*paragraphDefinition); !ok {
		t.Errorf("reader 3 type = %T, want *paragraphDefinition", readers[3])
	}
}

func TestSpec_RegisterBlockClassifiesDefinitionByReaderCapability(t *testing.T) {
	definition := &specRegistrationBlockDefinitionProbe{key: "probe"}
	spec := newSpec()

	if err := spec.registerBlock(&codeBlockDefinition{}); err != nil {
		t.Fatalf("register directive: %v", err)
	}
	if err := spec.registerBlock(definition); err != nil {
		t.Fatalf("register sugar: %v", err)
	}

	readers := spec.getReaders()
	if len(readers) != 2 || readers[1] != definition {
		t.Fatalf("registered readers = %#v, want common directive reader followed only by sugar definition", readers)
	}
	if _, ok := spec.getBlockDefinition(blockTypeCode); !ok {
		t.Error("directive definition was not registered")
	}
	if got, ok := spec.getBlockDefinition("PROBE"); !ok || got != definition {
		t.Errorf("registered definition = %T, %t, want input definition", got, ok)
	}
}

type specRegistrationBlockDefinitionProbe struct {
	key string
}

func (d *specRegistrationBlockDefinitionProbe) blockType() string {
	return d.key
}

func (*specRegistrationBlockDefinitionProbe) read(*blockContext) (parsedBlockNode, bool, error) {
	return nil, false, nil
}

func (*specRegistrationBlockDefinitionProbe) build(*Parser, parsedBlockNode) (ast.Block, error) {
	return nil, nil
}

func TestSpec_BlockRegistrationRejectsIncompleteFeatures(t *testing.T) {
	tests := []struct {
		name     string
		register func(*Spec) error
	}{
		{
			name: "block without definition",
			register: func(spec *Spec) error {
				return spec.registerBlock(nil)
			},
		},
		{
			name: "fallback without definition",
			register: func(spec *Spec) error {
				return spec.registerBlockFallback(nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec := newSpec()
			if err := tt.register(spec); err == nil {
				t.Fatal("incomplete registration returned no error")
			}
			if readers := spec.getReaders(); len(readers) != 1 {
				t.Errorf("incomplete registration installed %d feature readers", len(readers)-1)
			}
			if _, ok := spec.getBlockDefinition("heading"); ok {
				t.Error("incomplete registration installed a heading definition")
			}
			if _, ok := spec.getBlockDefinition("code"); ok {
				t.Error("incomplete registration installed a code definition")
			}
			if _, ok := spec.getBlockDefinition("paragraph"); ok {
				t.Error("incomplete registration installed a paragraph definition")
			}
		})
	}
}

func TestSpec_BlockRegistrationUsesDefinitionType(t *testing.T) {
	tests := []struct {
		name       string
		definition blockSugarDefinition
		key        string
	}{
		{
			name:       "heading",
			definition: &headingDefinition{},
			key:        "heading",
		},
		{
			name:       "list",
			definition: &listDefinition{},
			key:        "list",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec := newSpec()
			if err := spec.registerBlock(tt.definition); err != nil {
				t.Fatalf("register sugar: %v", err)
			}

			got, ok := spec.getBlockDefinition(tt.key)
			if !ok {
				t.Fatalf("definition %q was not registered", tt.key)
			}
			if got != tt.definition {
				t.Errorf("definition %q = %T, want %T", tt.key, got, tt.definition)
			}
			if len(spec.getReaders()) != 2 || spec.getReaders()[1] != tt.definition {
				t.Errorf("registered readers = %#v, want directive reader followed by %T", spec.getReaders(), tt.definition)
			}
		})
	}
}

func TestSpec_BlockRegistrationRejectsEmptyDefinitionType(t *testing.T) {
	spec := newSpec()
	definition := newSpecRegistrationSugarDefinition(&specRegistrationSugarReaderProbe{key: ""}, &paragraphDefinition{})
	if err := spec.registerBlock(definition); err == nil {
		t.Fatal("empty block type registration returned no error")
	}
	if len(spec.getReaders()) != 1 {
		t.Errorf("empty-key registration installed %d sugar readers", len(spec.getReaders())-1)
	}
	if _, ok := spec.getBlockDefinition(""); ok {
		t.Error("empty-type registration installed a definition")
	}
}

func TestSpec_BlockRegistrationRejectsReservedParagraphType(t *testing.T) {
	for _, key := range []string{"paragraph", "Paragraph", "PARAGRAPH", "pArAgRaPh"} {
		t.Run(key, func(t *testing.T) {
			spec := newSpec()
			definition := newSpecRegistrationSugarDefinition(&specRegistrationSugarReaderProbe{key: key}, &paragraphDefinition{})
			if err := spec.registerBlock(definition); err == nil {
				t.Fatal("paragraph sugar registration returned no error")
			}
			if len(spec.getReaders()) != 1 {
				t.Errorf("reserved-key registration installed %d sugar readers", len(spec.getReaders())-1)
			}
			if _, ok := spec.getBlockDefinition("paragraph"); ok {
				t.Error("reserved-type registration installed a paragraph definition")
			}
			if err := spec.registerBlockFallback(&paragraphDefinition{}); err != nil {
				t.Fatalf("register paragraph fallback after rejected sugar: %v", err)
			}
		})
	}
}

func TestParserParse_BlockSugarDefinitionTypeMismatchIsInternalError(t *testing.T) {
	spec := newSpec()
	reader := &specRegistrationSugarReaderProbe{
		key: "PrObE",
		node: &parsedBlock{
			Type:  "AcTuAl",
			Range: ast.Range{Start: ast.Position{Line: 1, Column: 1}, End: ast.Position{Line: 1, Column: 4}},
		},
		ok: true,
	}
	if err := spec.registerBlock(newSpecRegistrationSugarDefinition(reader, &paragraphDefinition{})); err != nil {
		t.Fatalf("register sugar: %v", err)
	}
	if err := spec.registerBlockFallback(&paragraphDefinition{}); err != nil {
		t.Fatalf("register paragraph fallback: %v", err)
	}

	got, err := NewParser(spec).Parse("text")
	if err == nil {
		t.Fatal("Parse accepted a mismatched sugar definition block type")
	}
	if got != nil {
		t.Errorf("Parse returned a document: %#v", got)
	}
	if !strings.Contains(err.Error(), `declared block type "probe"`) || !strings.Contains(err.Error(), `produced "actual"`) {
		t.Errorf("error = %q, want normalized declared and actual block types", err)
	}
	var diag *diagnostic.Error
	if errors.As(err, &diag) {
		t.Errorf("mismatch returned an unsupported-block diagnostic: %v", err)
	}
}

func TestSpec_BlockRegistrationRejectsNormalizedDuplicatesWithoutOverwrite(t *testing.T) {
	spec := newSpec()
	first := newSpecRegistrationDirectiveDefinition("ÄBC", &codeBlockDefinition{})
	second := &paragraphDefinition{}

	if err := spec.registerBlock(first); err != nil {
		t.Fatalf("first registration returned an error: %v", err)
	}
	if err := spec.registerBlock(newSpecRegistrationDirectiveDefinition("äbc", second)); err == nil {
		t.Fatal("case-only duplicate registration returned no error")
	}

	got, ok := spec.getBlockDefinition("ÄbC")
	if !ok {
		t.Fatal("normalized definition lookup failed")
	}
	if got != first {
		t.Errorf("duplicate registration replaced definition with %T", got)
	}

	if err := spec.registerBlock(newSpecRegistrationSugarDefinition(&specRegistrationSugarReaderProbe{key: "ÄbC"}, second)); err == nil {
		t.Fatal("cross-category duplicate registration returned no error")
	}
	if got, ok := spec.getBlockDefinition("äbc"); !ok || got != first {
		t.Errorf("cross-category duplicate replaced the first definition with %T", got)
	}
}

func TestSpec_BlockRegistrationRejectsDuplicateParagraphFallback(t *testing.T) {
	spec := newSpec()
	first := &paragraphDefinition{}
	if err := spec.registerBlockFallback(first); err != nil {
		t.Fatalf("first fallback registration returned an error: %v", err)
	}
	if err := spec.registerBlockFallback(&paragraphDefinition{}); err == nil {
		t.Fatal("duplicate fallback registration returned no error")
	}
	if got, ok := spec.getBlockDefinition("PARAGRAPH"); !ok || got != first {
		t.Errorf("duplicate fallback replaced the first definition with %T", got)
	}
}

func TestSpec_BlockFallbackRejectsNonParagraphDefinition(t *testing.T) {
	spec := newSpec()
	definition := &specRegistrationBlockDefinitionProbe{key: blockTypeHeading}

	if err := spec.registerBlockFallback(definition); err == nil {
		t.Fatal("non-paragraph fallback registration returned no error")
	}
	if _, ok := spec.getBlockDefinition(blockTypeHeading); ok {
		t.Error("rejected fallback installed a definition")
	}
	if len(spec.getReaders()) != 1 {
		t.Errorf("rejected fallback installed %d feature readers", len(spec.getReaders())-1)
	}
}

func TestSpec_BlockRegistrationRejectsParagraphDefinitionWithoutMutation(t *testing.T) {
	for _, key := range []string{"paragraph", "Paragraph", "PARAGRAPH", "pArAgRaPh"} {
		t.Run(key, func(t *testing.T) {
			spec := newSpec()
			if err := spec.registerBlock(newSpecRegistrationDirectiveDefinition(key, &paragraphDefinition{})); err == nil {
				t.Fatal("paragraph directive registration returned no error")
			}
			if _, ok := spec.getBlockDefinition("paragraph"); ok {
				t.Error("rejected paragraph registration installed a definition")
			}
			if err := spec.registerBlockFallback(&paragraphDefinition{}); err != nil {
				t.Fatalf("register paragraph fallback after rejected directive: %v", err)
			}
		})
	}
}

func TestSpec_BlockDefinitionRegistrationRejectsReservedParagraphType(t *testing.T) {
	spec := newSpec()
	if err := spec.registerBlockDefinition(&paragraphDefinition{}); err == nil {
		t.Fatal("paragraph block definition registration returned no error")
	}
	if _, ok := spec.getBlockDefinition("paragraph"); ok {
		t.Error("rejected paragraph block definition was installed")
	}
}

func TestSpec_ParagraphFallbackPreventsLaterDirectiveRegistration(t *testing.T) {
	spec := newSpec()
	paragraph := &paragraphDefinition{}
	if err := spec.registerBlockFallback(paragraph); err != nil {
		t.Fatalf("register paragraph fallback: %v", err)
	}
	if err := spec.registerBlock(newSpecRegistrationDirectiveDefinition("PARAGRAPH", &codeBlockDefinition{})); err == nil {
		t.Fatal("paragraph directive registration after fallback returned no error")
	}
	got, ok := spec.getBlockDefinition("paragraph")
	if !ok {
		t.Fatal("paragraph fallback definition was removed")
	}
	if got != paragraph {
		t.Errorf("paragraph definition = %T, want fallback definition %T", got, paragraph)
	}
}

func TestParserParse_RejectsInvalidSpecBeforeReading(t *testing.T) {
	spec := newSpec()
	reader := &specRegistrationReaderProbe{}
	definition := newSpecRegistrationSugarDefinition(reader, &paragraphDefinition{})
	if err := spec.registerBlock(definition); err != nil {
		t.Fatalf("register sugar: %v", err)
	}

	got, err := NewParser(spec).Parse("plain text")
	if err == nil {
		t.Fatal("Parse accepted a Spec without a Paragraph fallback")
	}
	if got != nil {
		t.Errorf("Parse returned a document: %v", got)
	}
	if !strings.Contains(err.Error(), "block fallback") {
		t.Errorf("error = %q, want missing block fallback", err)
	}
	if reader.calls != 0 {
		t.Errorf("reader calls = %d, want 0", reader.calls)
	}
}

func TestSpec_OneDirectiveReaderDispatchesMultipleDefinitions(t *testing.T) {
	spec := newSpec()
	if err := spec.registerBlock(newSpecRegistrationDirectiveDefinition("alpha", &codeBlockDefinition{})); err != nil {
		t.Fatalf("register alpha directive: %v", err)
	}
	if err := spec.registerBlock(newSpecRegistrationDirectiveDefinition("BETA", &codeBlockDefinition{})); err != nil {
		t.Fatalf("register beta directive: %v", err)
	}
	if err := spec.registerBlockFallback(&paragraphDefinition{}); err != nil {
		t.Fatalf("register paragraph fallback: %v", err)
	}

	got, err := NewParser(spec).Parse(":::[ALPHA:a]\nfirst\n:::\n\n:::[beta:b]\nsecond\n:::")
	if err != nil {
		t.Fatalf("Parse returned an error: %v", err)
	}
	if len(got.Blocks) != 2 {
		t.Fatalf("block count = %d, want 2", len(got.Blocks))
	}

	want := []struct {
		language string
		text     string
	}{
		{language: "a", text: "first"},
		{language: "b", text: "second"},
	}
	for i, expectation := range want {
		block, ok := got.Blocks[i].(*ast.CodeBlock)
		if !ok {
			t.Fatalf("block %d type = %T, want *ast.CodeBlock", i, got.Blocks[i])
		}
		if block.Language != expectation.language || block.Text != expectation.text {
			t.Errorf("block %d = language %q text %q, want language %q text %q", i, block.Language, block.Text, expectation.language, expectation.text)
		}
	}
}

func TestCoreSpec_BlockTypeNormalizationPreservesAttributeAndContent(t *testing.T) {
	got, err := NewParser(coreSpec()).Parse(":::[CoDe:GoLang]\nMiXeD Content\n:::")
	if err != nil {
		t.Fatalf("Parse returned an error: %v", err)
	}
	if len(got.Blocks) != 1 {
		t.Fatalf("block count = %d, want 1", len(got.Blocks))
	}

	block, ok := got.Blocks[0].(*ast.CodeBlock)
	if !ok {
		t.Fatalf("block type = %T, want *ast.CodeBlock", got.Blocks[0])
	}
	if block.Language != "GoLang" {
		t.Errorf("language = %q, want %q", block.Language, "GoLang")
	}
	if block.Text != "MiXeD Content" {
		t.Errorf("text = %q, want %q", block.Text, "MiXeD Content")
	}
}

func TestCoreSpec_NonMatchingSugarReadersDoNotConsumeInput(t *testing.T) {
	readers := coreSpec().getReaders()
	ctx := &blockContext{lines: []string{"plain text"}}

	for i, reader := range readers[:len(readers)-1] {
		node, ok, err := reader.read(ctx)
		if err != nil {
			t.Fatalf("reader %d returned an error: %v", i, err)
		}
		if ok || node != nil {
			t.Fatalf("reader %d unexpectedly accepted plain text", i)
		}
		if ctx.pos != 0 {
			t.Fatalf("reader %d changed cursor to %d, want 0", i, ctx.pos)
		}
	}
}

type specRegistrationReaderProbe struct {
	calls int
}

func (r *specRegistrationReaderProbe) read(*blockContext) (parsedBlockNode, bool, error) {
	r.calls++
	return nil, false, nil
}

func (*specRegistrationReaderProbe) blockType() string {
	return "probe"
}

type specRegistrationSugarReaderProbe struct {
	key   string
	calls int
	node  parsedBlockNode
	ok    bool
}

type specRegistrationSugarDefinition struct {
	reader  specRegistrationSugarReader
	builder blockBuilder
}

type specRegistrationSugarReader interface {
	blockReader
	blockType() string
}

func newSpecRegistrationSugarDefinition(reader specRegistrationSugarReader, builder blockBuilder) *specRegistrationSugarDefinition {
	return &specRegistrationSugarDefinition{reader: reader, builder: builder}
}

func (d *specRegistrationSugarDefinition) blockType() string {
	return d.reader.blockType()
}

func (d *specRegistrationSugarDefinition) read(ctx *blockContext) (parsedBlockNode, bool, error) {
	return d.reader.read(ctx)
}

func (d *specRegistrationSugarDefinition) build(parser *Parser, node parsedBlockNode) (ast.Block, error) {
	return d.builder.build(parser, node)
}

func (r *specRegistrationSugarReaderProbe) blockType() string {
	return r.key
}

func (r *specRegistrationSugarReaderProbe) read(*blockContext) (parsedBlockNode, bool, error) {
	r.calls++
	return r.node, r.ok, nil
}

type specRegistrationDirectiveDefinition struct {
	key     string
	builder blockBuilder
}

func newSpecRegistrationDirectiveDefinition(key string, builder blockBuilder) *specRegistrationDirectiveDefinition {
	return &specRegistrationDirectiveDefinition{key: key, builder: builder}
}

func (d *specRegistrationDirectiveDefinition) blockType() string {
	return d.key
}

func (d *specRegistrationDirectiveDefinition) build(parser *Parser, node parsedBlockNode) (ast.Block, error) {
	return d.builder.build(parser, node)
}
