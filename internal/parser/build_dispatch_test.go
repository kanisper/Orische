package parser

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"orische/internal/ast"
	"orische/internal/diagnostic"
	"orische/internal/parser/feature"
	"orische/internal/parser/syntax"
)

func TestBuildBlockUnknownTypeReturnsDiagnostic(t *testing.T) {
	p := mustCoreParser(t)
	rng := ast.Range{Start: ast.Position{Line: 4, Column: 2}, End: ast.Position{Line: 6, Column: 3}}

	got, err := p.buildBlock(&feature.TextBlock{Type: "MiSsInG", Range: rng})
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

func TestBuildBlockRejectsNilNode(t *testing.T) {
	p := mustCoreParser(t)
	var typedNil *feature.TextBlock
	for _, node := range []feature.BlockNode{nil, typedNil} {
		got, err := p.buildBlock(node)
		if got != nil {
			t.Errorf("buildBlock(%#v) returned a block: %#v", node, got)
		}
		if err == nil || !strings.Contains(err.Error(), "build block: node is nil") {
			t.Errorf("buildBlock(%#v) error = %v, want nil-node error", node, err)
		}
	}
}

func TestBuildBlockRejectsTypedNilAST(t *testing.T) {
	language := syntax.Core()
	language.Blocks = append(language.Blocks, &testTypedNilBlockDefinition{})
	p, err := NewParser(language)
	if err != nil {
		t.Fatalf("NewParser returned an error: %v", err)
	}

	got, err := p.buildBlock(&feature.TextBlock{Type: "typed-nil"})
	if got != nil {
		t.Errorf("buildBlock returned a block: %#v", got)
	}
	if err == nil || !strings.Contains(err.Error(), `build "typed-nil" block: definition returned a nil block`) {
		t.Errorf("buildBlock error = %v, want typed-nil AST error", err)
	}
}

func TestBuildBlockWrongIRTypeIsInternalError(t *testing.T) {
	p := mustCoreParser(t)
	got, err := p.buildBlock(&testBlockNode{typ: "code"})
	if got != nil {
		t.Errorf("buildBlock returned a block: %#v", got)
	}
	if err == nil || !strings.Contains(err.Error(), `build "code" block: expected *feature.TextBlock`) {
		t.Errorf("buildBlock error = %v, want code IR mismatch context", err)
	}
	var diag *diagnostic.Error
	if errors.As(err, &diag) {
		t.Errorf("IR mismatch returned a diagnostic: %v", err)
	}
}

func TestBuildBlockPreservesDiagnosticIdentity(t *testing.T) {
	wantErr := &diagnostic.Error{
		Message: "builder diagnostic",
		Range:   ast.Range{Start: ast.Position{Line: 7, Column: 2}, End: ast.Position{Line: 8, Column: 4}},
	}
	language := syntax.Core()
	language.Blocks = append(language.Blocks, &testErrorDefinition{typ: "diagnostic", err: wantErr})
	p, err := NewParser(language)
	if err != nil {
		t.Fatalf("NewParser returned an error: %v", err)
	}

	got, err := p.Parse(":::[diagnostic]\ntext\n:::")
	if got != nil {
		t.Errorf("Parse returned a document: %#v", got)
	}
	if err != wantErr {
		t.Errorf("Parse error = %v, want original diagnostic %v", err, wantErr)
	}
}

func TestListBuildPreservesNestedDiagnosticIdentity(t *testing.T) {
	wantErr := &diagnostic.Error{
		Message: "paragraph diagnostic",
		Range:   ast.Range{Start: ast.Position{Line: 1, Column: 3}, End: ast.Position{Line: 1, Column: 6}},
	}
	language := syntax.Core()
	language.Paragraph = &testErrorDefinition{typ: feature.ParagraphBlockType, err: wantErr}
	p, err := NewParser(language)
	if err != nil {
		t.Fatalf("NewParser returned an error: %v", err)
	}

	_, err = p.Parse("* item")
	if err != wantErr {
		t.Errorf("Parse error = %v, want original diagnostic %v", err, wantErr)
	}
}

func TestListBuildWrapsNestedOrdinaryError(t *testing.T) {
	wantErr := errors.New("paragraph failed")
	language := syntax.Core()
	language.Paragraph = &testErrorDefinition{typ: feature.ParagraphBlockType, err: wantErr}
	p, err := NewParser(language)
	if err != nil {
		t.Fatalf("NewParser returned an error: %v", err)
	}

	_, err = p.Parse("* item")
	if !errors.Is(err, wantErr) {
		t.Errorf("Parse error = %v, want errors.Is(..., paragraph failure)", err)
	}
	if !strings.Contains(err.Error(), `build "paragraph" block`) || !strings.Contains(err.Error(), `build "list" block`) {
		t.Errorf("Parse error = %q, want paragraph and list build context", err)
	}
}

func TestActiveLanguageReachesEveryInlineCapableBuiltInBlock(t *testing.T) {
	calls := 0
	mark := &testInlineDirectiveDefinition{
		typ:    "mark",
		policy: feature.InlineContentNested,
		build: func(candidate feature.InlineDirectiveCandidate) (ast.Inline, error) {
			calls++
			return &ast.Emphasis{Content: candidate.NestedContent, Range: candidate.Range}, nil
		},
	}
	p := mustParserWithAdditionalInlines(t, mark)

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

func TestCodeBlockDoesNotInvokeActiveInlineDefinitions(t *testing.T) {
	calls := 0
	emphasis := &testInlineDirectiveDefinition{
		typ:    "em",
		policy: feature.InlineContentNested,
		build: func(candidate feature.InlineDirectiveCandidate) (ast.Inline, error) {
			calls++
			return &ast.Emphasis{Content: candidate.NestedContent, Range: candidate.Range}, nil
		},
	}
	language := syntax.Core()
	language.Inlines = []feature.InlineDirectiveDefinition{emphasis}
	p, err := NewParser(language)
	if err != nil {
		t.Fatalf("NewParser returned an error: %v", err)
	}

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

type testBlockNode struct {
	typ string
	rng ast.Range
}

func (n *testBlockNode) BlockType() string {
	return n.typ
}

func (n *testBlockNode) BlockRange() ast.Range {
	return n.rng
}

type testErrorDefinition struct {
	typ string
	err error
}

type testTypedNilBlockDefinition struct{}

func (*testTypedNilBlockDefinition) BlockType() string {
	return "typed-nil"
}

func (*testTypedNilBlockDefinition) BuildBlock(feature.BuildContext, feature.BlockNode) (ast.Block, error) {
	var block *ast.Paragraph
	return block, nil
}

func (d *testErrorDefinition) BlockType() string {
	return d.typ
}

func (d *testErrorDefinition) BuildBlock(feature.BuildContext, feature.BlockNode) (ast.Block, error) {
	if d.err == nil {
		return nil, fmt.Errorf("test error definition has no error")
	}
	return nil, d.err
}

func (d *testErrorDefinition) BuildParagraph(feature.BuildContext, feature.BlockNode) (*ast.Paragraph, error) {
	if d.err == nil {
		return nil, fmt.Errorf("test error definition has no error")
	}
	return nil, d.err
}
