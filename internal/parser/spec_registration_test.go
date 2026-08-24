package parser

import (
	"strings"
	"testing"

	"orische/internal/ast"
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
				return spec.registerBlockSugar("heading", nil, &headingBuilder{})
			},
		},
		{
			name: "sugar without builder",
			register: func(spec *Spec) error {
				return spec.registerBlockSugar("heading", &headingReader{}, nil)
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
		{
			name: "paragraph as sugar",
			register: func(spec *Spec) error {
				return spec.registerBlockSugar("paragraph", &paragraphReader{}, &paragraphBuilder{})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec := newSpec()
			if err := tt.register(spec); err == nil {
				t.Fatal("incomplete registration returned no error")
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

	if err := spec.registerBlockSugar("ÄbC", &headingReader{}, second); err == nil {
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
	if err := spec.registerBlockSugar("probe", reader, &paragraphBuilder{}); err != nil {
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
