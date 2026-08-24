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
	if _, ok := readers[1].(*headingReader); !ok {
		t.Errorf("reader 1 type = %T, want *headingReader", readers[1])
	}
	if _, ok := readers[2].(*listReader); !ok {
		t.Errorf("reader 2 type = %T, want *listReader", readers[2])
	}
	if _, ok := readers[3].(*paragraphReader); !ok {
		t.Errorf("reader 3 type = %T, want *paragraphReader", readers[3])
	}
}

func TestSpec_BlockRegistrationRejectsIncompleteFeatures(t *testing.T) {
	tests := []struct {
		name     string
		register func(*Spec) error
	}{
		{
			name: "sugar without reader",
			register: func(spec *Spec) error {
				return spec.registerBlockSugar(nil, &headingBuilder{})
			},
		},
		{
			name: "sugar without builder",
			register: func(spec *Spec) error {
				return spec.registerBlockSugar(&headingReader{}, nil)
			},
		},
		{
			name: "directive without builder",
			register: func(spec *Spec) error {
				return spec.registerBlockDirectiveDefinition("code", nil)
			},
		},
		{
			name: "fallback without builder",
			register: func(spec *Spec) error {
				return spec.registerParagraphFallback(nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec := newSpec()
			if err := tt.register(spec); err == nil {
				t.Fatal("incomplete registration returned no error")
			}
			if readers := spec.getReaders(); len(readers) != 0 {
				t.Errorf("incomplete registration installed %d readers", len(readers))
			}
			if _, ok := spec.getBuilder("heading"); ok {
				t.Error("incomplete registration installed a heading builder")
			}
			if _, ok := spec.getBuilder("code"); ok {
				t.Error("incomplete registration installed a code builder")
			}
			if _, ok := spec.getBuilder("paragraph"); ok {
				t.Error("incomplete registration installed a paragraph builder")
			}
		})
	}
}

func TestSpec_BlockSugarRegistrationUsesReaderBuilderKey(t *testing.T) {
	tests := []struct {
		name    string
		reader  blockSugarReader
		builder blockBuilder
		key     string
	}{
		{
			name:    "heading",
			reader:  &headingReader{},
			builder: &headingBuilder{},
			key:     "heading",
		},
		{
			name:    "list",
			reader:  &listReader{},
			builder: &listBuilder{},
			key:     "list",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec := newSpec()
			if err := spec.registerBlockSugar(tt.reader, tt.builder); err != nil {
				t.Fatalf("register sugar: %v", err)
			}

			got, ok := spec.getBuilder(tt.key)
			if !ok {
				t.Fatalf("builder %q was not registered", tt.key)
			}
			if got != tt.builder {
				t.Errorf("builder %q = %T, want %T", tt.key, got, tt.builder)
			}
			if len(spec.getReaders()) != 1 || spec.getReaders()[0] != tt.reader {
				t.Errorf("registered readers = %#v, want %#v", spec.getReaders(), []blockReader{tt.reader})
			}
		})
	}
}

func TestSpec_BlockSugarRegistrationRejectsEmptyReaderBuilderKey(t *testing.T) {
	spec := newSpec()
	reader := &specRegistrationSugarReaderProbe{key: ""}
	if err := spec.registerBlockSugar(reader, &paragraphBuilder{}); err == nil {
		t.Fatal("empty reader builder key registration returned no error")
	}
	if len(spec.getReaders()) != 0 {
		t.Errorf("empty-key registration installed %d readers", len(spec.getReaders()))
	}
	if _, ok := spec.getBuilder(""); ok {
		t.Error("empty-key registration installed a builder")
	}
}

func TestSpec_BlockSugarRegistrationRejectsReservedParagraphKey(t *testing.T) {
	for _, key := range []string{"paragraph", "Paragraph", "PARAGRAPH", "pArAgRaPh"} {
		t.Run(key, func(t *testing.T) {
			spec := newSpec()
			reader := &specRegistrationSugarReaderProbe{key: key}
			if err := spec.registerBlockSugar(reader, &paragraphBuilder{}); err == nil {
				t.Fatal("paragraph sugar registration returned no error")
			}
			if len(spec.getReaders()) != 0 {
				t.Errorf("reserved-key registration installed %d readers", len(spec.getReaders()))
			}
			if _, ok := spec.getBuilder("paragraph"); ok {
				t.Error("reserved-key registration installed a paragraph builder")
			}
			if err := spec.registerParagraphFallback(&paragraphBuilder{}); err != nil {
				t.Fatalf("register paragraph fallback after rejected sugar: %v", err)
			}
		})
	}
}

func TestParserParse_BlockSugarReaderBuilderKeyMismatchIsInternalError(t *testing.T) {
	spec := newSpec()
	if err := spec.registerBlockDirectiveReader(); err != nil {
		t.Fatalf("register block directive reader: %v", err)
	}
	reader := &specRegistrationSugarReaderProbe{
		key: "PrObE",
		node: &parsedBlock{
			Type:  "AcTuAl",
			Range: ast.Range{Start: ast.Position{Line: 1, Column: 1}, End: ast.Position{Line: 1, Column: 4}},
		},
		ok: true,
	}
	if err := spec.registerBlockSugar(reader, &paragraphBuilder{}); err != nil {
		t.Fatalf("register sugar: %v", err)
	}
	if err := spec.registerParagraphFallback(&paragraphBuilder{}); err != nil {
		t.Fatalf("register paragraph fallback: %v", err)
	}

	got, err := NewParser(spec).Parse("text")
	if err == nil {
		t.Fatal("Parse accepted a mismatched sugar reader builder key")
	}
	if got != nil {
		t.Errorf("Parse returned a document: %#v", got)
	}
	if !strings.Contains(err.Error(), `declared builder key "probe"`) || !strings.Contains(err.Error(), `produced "actual"`) {
		t.Errorf("error = %q, want normalized declared and actual builder keys", err)
	}
	var diag *diagnostic.Error
	if errors.As(err, &diag) {
		t.Errorf("mismatch returned an unsupported-block diagnostic: %v", err)
	}
}

func TestSpec_BlockRegistrationRejectsNormalizedDuplicatesWithoutOverwrite(t *testing.T) {
	spec := newSpec()
	first := &codeBlockBuilder{}
	second := &paragraphBuilder{}

	if err := spec.registerBlockDirectiveDefinition("ÄBC", first); err != nil {
		t.Fatalf("first registration returned an error: %v", err)
	}
	if err := spec.registerBlockDirectiveDefinition("äbc", second); err == nil {
		t.Fatal("case-only duplicate registration returned no error")
	}

	got, ok := spec.getBuilder("ÄbC")
	if !ok {
		t.Fatal("normalized builder lookup failed")
	}
	if got != first {
		t.Errorf("duplicate registration replaced builder with %T", got)
	}

	if err := spec.registerBlockSugar(&specRegistrationSugarReaderProbe{key: "ÄbC"}, second); err == nil {
		t.Fatal("cross-category duplicate registration returned no error")
	}
	if got, ok := spec.getBuilder("äbc"); !ok || got != first {
		t.Errorf("cross-category duplicate replaced the first builder with %T", got)
	}
}

func TestSpec_BlockRegistrationRejectsDuplicateDirectiveReader(t *testing.T) {
	spec := newSpec()
	if err := spec.registerBlockDirectiveReader(); err != nil {
		t.Fatalf("first reader registration returned an error: %v", err)
	}
	if err := spec.registerBlockDirectiveReader(); err == nil {
		t.Fatal("duplicate reader registration returned no error")
	}
}

func TestSpec_BlockRegistrationRejectsDuplicateParagraphFallback(t *testing.T) {
	spec := newSpec()
	first := &paragraphBuilder{}
	if err := spec.registerParagraphFallback(first); err != nil {
		t.Fatalf("first fallback registration returned an error: %v", err)
	}
	if err := spec.registerParagraphFallback(&paragraphBuilder{}); err == nil {
		t.Fatal("duplicate fallback registration returned no error")
	}
	if got, ok := spec.getBuilder("PARAGRAPH"); !ok || got != first {
		t.Errorf("duplicate fallback replaced the first builder with %T", got)
	}
}

func TestSpec_BlockRegistrationRejectsParagraphDirectiveDefinitionWithoutMutation(t *testing.T) {
	for _, key := range []string{"paragraph", "Paragraph", "PARAGRAPH", "pArAgRaPh"} {
		t.Run(key, func(t *testing.T) {
			spec := newSpec()
			if err := spec.registerBlockDirectiveDefinition(key, &paragraphBuilder{}); err == nil {
				t.Fatal("paragraph directive registration returned no error")
			}
			if _, ok := spec.getBuilder("paragraph"); ok {
				t.Error("rejected paragraph directive installed a builder")
			}
			if len(spec.getReaders()) != 0 {
				t.Errorf("rejected paragraph directive installed %d readers", len(spec.getReaders()))
			}
			if err := spec.registerParagraphFallback(&paragraphBuilder{}); err != nil {
				t.Fatalf("register paragraph fallback after rejected directive: %v", err)
			}
		})
	}
}

func TestSpec_BlockBuilderRegistrationRejectsReservedParagraphKey(t *testing.T) {
	spec := newSpec()
	if err := spec.registerBlockBuilder("Paragraph", &paragraphBuilder{}); err == nil {
		t.Fatal("paragraph block builder registration returned no error")
	}
	if _, ok := spec.getBuilder("paragraph"); ok {
		t.Error("rejected paragraph block builder installed a builder")
	}
}

func TestSpec_ParagraphFallbackPreventsLaterDirectiveRegistration(t *testing.T) {
	spec := newSpec()
	paragraph := &paragraphBuilder{}
	if err := spec.registerParagraphFallback(paragraph); err != nil {
		t.Fatalf("register paragraph fallback: %v", err)
	}
	if err := spec.registerBlockDirectiveDefinition("PARAGRAPH", &codeBlockBuilder{}); err == nil {
		t.Fatal("paragraph directive registration after fallback returned no error")
	}
	got, ok := spec.getBuilder("paragraph")
	if !ok {
		t.Fatal("paragraph fallback builder was removed")
	}
	if got != paragraph {
		t.Errorf("paragraph builder = %T, want fallback builder %T", got, paragraph)
	}
}

func TestParserParse_RejectsSpecWithoutBlockDirectiveReader(t *testing.T) {
	spec := newSpec()
	if err := spec.registerParagraphFallback(&paragraphBuilder{}); err != nil {
		t.Fatalf("register paragraph fallback: %v", err)
	}

	got, err := NewParser(spec).Parse("plain text")
	if err == nil {
		t.Fatal("Parse accepted a Spec without a Block Directive reader")
	}
	if got != nil {
		t.Errorf("Parse returned a document: %v", got)
	}
	if !strings.Contains(err.Error(), "block directive reader") {
		t.Errorf("error = %q, want missing block directive reader", err)
	}
}

func TestParserParse_RejectsInvalidSpecBeforeReading(t *testing.T) {
	spec := newSpec()
	if err := spec.registerBlockDirectiveReader(); err != nil {
		t.Fatalf("register directive reader: %v", err)
	}
	reader := &specRegistrationReaderProbe{}
	if err := spec.registerBlockSugar(reader, &paragraphBuilder{}); err != nil {
		t.Fatalf("register sugar: %v", err)
	}

	got, err := NewParser(spec).Parse("plain text")
	if err == nil {
		t.Fatal("Parse accepted a Spec without a Paragraph fallback")
	}
	if got != nil {
		t.Errorf("Parse returned a document: %v", got)
	}
	if !strings.Contains(err.Error(), "paragraph fallback") {
		t.Errorf("error = %q, want missing paragraph fallback", err)
	}
	if reader.calls != 0 {
		t.Errorf("reader calls = %d, want 0", reader.calls)
	}
}

func TestSpec_OneDirectiveReaderDispatchesMultipleDefinitions(t *testing.T) {
	spec := newSpec()
	if err := spec.registerBlockDirectiveReader(); err != nil {
		t.Fatalf("register directive reader: %v", err)
	}
	if err := spec.registerBlockDirectiveDefinition("alpha", &codeBlockBuilder{}); err != nil {
		t.Fatalf("register alpha directive: %v", err)
	}
	if err := spec.registerBlockDirectiveDefinition("BETA", &codeBlockBuilder{}); err != nil {
		t.Fatalf("register beta directive: %v", err)
	}
	if err := spec.registerParagraphFallback(&paragraphBuilder{}); err != nil {
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

func (*specRegistrationReaderProbe) builderKey() string {
	return "probe"
}

type specRegistrationSugarReaderProbe struct {
	key   string
	calls int
	node  parsedBlockNode
	ok    bool
}

func (r *specRegistrationSugarReaderProbe) builderKey() string {
	return r.key
}

func (r *specRegistrationSugarReaderProbe) read(*blockContext) (parsedBlockNode, bool, error) {
	r.calls++
	return r.node, r.ok, nil
}
