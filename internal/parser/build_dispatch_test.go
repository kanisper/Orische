package parser

import (
	"errors"
	"testing"

	"orische/internal/ast"
	"orische/internal/diagnostic"
)

func TestBuildBlockUnknownDirectiveReturnsDiagnostic(t *testing.T) {
	p := mustCoreParser(t)
	rng := ast.Range{Start: ast.Position{Line: 4, Column: 2}, End: ast.Position{Line: 6, Column: 3}}

	got, err := p.buildBlock(&blockDirectiveNode{dirtype: "MiSsInG", rng: rng})
	if got != nil {
		t.Errorf("buildBlock returned a block: %#v", got)
	}
	var diag *diagnostic.Error
	if !errors.As(err, &diag) {
		t.Fatalf("buildBlock error type = %T, want *diagnostic.Error", err)
	}
	if diag.Message != `unsupported block directive type "missing"` || diag.Range != rng {
		t.Errorf("diagnostic = %#v, want normalized type and original range", diag)
	}
}

func TestConfiguredInlineDefinitionsReachInlineCapableBlocks(t *testing.T) {
	calls := 0
	mark := inlineDefinition{
		policy: inlineContentNested,
		build: func(candidate inlineCandidate) ast.Inline {
			calls++
			return &ast.Emphasis{Content: candidate.nestedContent, Range: candidate.rng}
		},
	}
	p := parserWithInlineDefinitions(t, map[string]inlineDefinition{"mark": mark})

	doc, err := p.Parse("= :[mark]{見出し}\n\n:[mark]{段落}\n\n* :[mark]{外}\n** :[mark]{内}")
	if err != nil {
		t.Fatalf("Parse returned an error: %v", err)
	}
	if calls != 4 {
		t.Errorf("custom inline builder calls = %d, want 4", calls)
	}
	heading := doc.Blocks[0].(*ast.Heading)
	headingMark := heading.Content[0].(*ast.Emphasis)
	if headingMark.Range.Start != (ast.Position{Line: 1, Column: 3}) {
		t.Errorf("heading mark starts at %#v, want line 1 column 3", headingMark.Range.Start)
	}
	paragraph := doc.Blocks[1].(*ast.Paragraph)
	paragraphMark := paragraph.Content[0].(*ast.Emphasis)
	if paragraphMark.Range.Start != (ast.Position{Line: 3, Column: 1}) {
		t.Errorf("paragraph mark starts at %#v, want line 3 column 1", paragraphMark.Range.Start)
	}
	outer := doc.Blocks[2].(*ast.List)
	outerParagraph := outer.Items[0].Blocks[0].(*ast.Paragraph)
	outerMark, ok := outerParagraph.Content[0].(*ast.Emphasis)
	if !ok {
		t.Errorf("outer inline type = %T, want *ast.Emphasis", outerParagraph.Content[0])
	} else if outerMark.Range.Start != (ast.Position{Line: 5, Column: 3}) {
		t.Errorf("outer mark starts at %#v, want line 5 column 3", outerMark.Range.Start)
	}
	nested := outer.Items[0].Blocks[1].(*ast.List)
	nestedParagraph := nested.Items[0].Blocks[0].(*ast.Paragraph)
	nestedMark, ok := nestedParagraph.Content[0].(*ast.Emphasis)
	if !ok {
		t.Errorf("nested inline type = %T, want *ast.Emphasis", nestedParagraph.Content[0])
	} else if nestedMark.Range.Start != (ast.Position{Line: 6, Column: 4}) {
		t.Errorf("nested mark starts at %#v, want line 6 column 4", nestedMark.Range.Start)
	}
}

func TestCodeBlockPreservesInlineMarkersLiterally(t *testing.T) {
	calls := 0
	emphasis := inlineDefinition{
		policy: inlineContentNested,
		build: func(candidate inlineCandidate) ast.Inline {
			calls++
			return &ast.Emphasis{Content: candidate.nestedContent, Range: candidate.rng}
		},
	}
	p := parserWithInlineDefinitions(t, map[string]inlineDefinition{"em": emphasis})

	doc, err := p.Parse(":::[code:txt]\n:[em]{日}\n:::")
	if err != nil {
		t.Fatalf("Parse returned an error: %v", err)
	}
	if calls != 0 {
		t.Errorf("inline builder calls = %d, want 0", calls)
	}
	code := doc.Blocks[0].(*ast.CodeBlock)
	if code.Language != "txt" || code.Text != ":[em]{日}" {
		t.Errorf("CodeBlock = %#v, want literal inline-like text", code)
	}
}
