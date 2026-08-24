package parser

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"orische/internal/ast"
)

func TestParserParse_PropagatesActiveSpecToParagraphBuilder(t *testing.T) {
	parser, probes := newActiveSpecTestParser(t)

	got, err := parser.Parse("日 :[em]{é :[link:/x]{界}}")
	if err != nil {
		t.Fatalf("Parse returned an error: %v", err)
	}

	want := &ast.Document{
		Blocks: []ast.Block{
			&ast.Paragraph{
				Content: []ast.Inline{
					&ast.Text{
						Value: "日 ",
						Range: ast.Range{
							Start: ast.Position{Line: 1, Column: 1},
							End:   ast.Position{Line: 1, Column: 2},
						},
					},
					&ast.Emphasis{
						Content: []ast.Inline{
							&ast.Text{
								Value: "é ",
								Range: ast.Range{
									Start: ast.Position{Line: 1, Column: 9},
									End:   ast.Position{Line: 1, Column: 10},
								},
							},
							&ast.Link{
								URI: "/x",
								Content: []ast.Inline{
									&ast.Text{
										Value: "界",
										Range: ast.Range{
											Start: ast.Position{Line: 1, Column: 22},
											End:   ast.Position{Line: 1, Column: 22},
										},
									},
								},
								Range: ast.Range{
									Start: ast.Position{Line: 1, Column: 11},
									End:   ast.Position{Line: 1, Column: 23},
								},
							},
						},
						Range: ast.Range{
							Start: ast.Position{Line: 1, Column: 3},
							End:   ast.Position{Line: 1, Column: 24},
						},
					},
				},
				Range: ast.Range{
					Start: ast.Position{Line: 1, Column: 1},
					End:   ast.Position{Line: 1, Column: 24},
				},
			},
		},
		Range: ast.Range{
			Start: ast.Position{Line: 1, Column: 1},
			End:   ast.Position{Line: 1, Column: 24},
		},
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Parse returned an unexpected document (-want +got):\n%s", diff)
	}
	if got := probes["paragraph"].calls; got != 1 {
		t.Errorf("paragraph builder calls = %d, want 1", got)
	}
}

func TestParserParse_PropagatesActiveSpecToHeadingBuilder(t *testing.T) {
	parser, probes := newActiveSpecTestParser(t)

	got, err := parser.Parse("= 日 :[em]{界}")
	if err != nil {
		t.Fatalf("Parse returned an error: %v", err)
	}

	want := &ast.Document{
		Blocks: []ast.Block{
			&ast.Heading{
				Level: 1,
				Content: []ast.Inline{
					&ast.Text{
						Value: "日 ",
						Range: ast.Range{
							Start: ast.Position{Line: 1, Column: 3},
							End:   ast.Position{Line: 1, Column: 4},
						},
					},
					&ast.Emphasis{
						Content: []ast.Inline{
							&ast.Text{
								Value: "界",
								Range: ast.Range{
									Start: ast.Position{Line: 1, Column: 11},
									End:   ast.Position{Line: 1, Column: 11},
								},
							},
						},
						Range: ast.Range{
							Start: ast.Position{Line: 1, Column: 5},
							End:   ast.Position{Line: 1, Column: 12},
						},
					},
				},
				Range: ast.Range{
					Start: ast.Position{Line: 1, Column: 1},
					End:   ast.Position{Line: 1, Column: 12},
				},
			},
		},
		Range: ast.Range{
			Start: ast.Position{Line: 1, Column: 1},
			End:   ast.Position{Line: 1, Column: 12},
		},
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Parse returned an unexpected document (-want +got):\n%s", diff)
	}
	if got := probes["heading"].calls; got != 1 {
		t.Errorf("heading builder calls = %d, want 1", got)
	}
}

func TestParserParse_PropagatesActiveSpecThroughListItems(t *testing.T) {
	spec := newSpec()
	readerProbe := &activeSpecReaderProbe{}
	spec.addBlockReader(readerProbe)
	spec.addBlockReader(&listReader{})

	parser := NewParser(spec)
	paragraphProbe := &activeParserProbeBuilder{
		t:          t,
		wantParser: parser,
		wantSpec:   spec,
		delegate:   &paragraphBuilder{},
	}
	listProbe := &activeParserProbeBuilder{
		t:          t,
		wantParser: parser,
		wantSpec:   spec,
		delegate:   &listBuilder{},
	}
	spec.addBlockBuilder("paragraph", paragraphProbe)
	spec.addBlockBuilder("list", listProbe)

	got, err := parser.Parse("* 日 :[em]{外}\n** 界 :[link:/x]{内}")
	if err != nil {
		t.Fatalf("Parse returned an error: %v", err)
	}

	if readerProbe.calls != 1 {
		t.Errorf("document reader probe calls = %d, want 1", readerProbe.calls)
	}
	if listProbe.calls != 2 {
		t.Errorf("list builder calls = %d, want 2", listProbe.calls)
	}
	if paragraphProbe.calls != 2 {
		t.Errorf("paragraph builder calls = %d, want 2", paragraphProbe.calls)
	}

	if len(got.Blocks) != 1 {
		t.Fatalf("document blocks = %d, want 1", len(got.Blocks))
	}
	outer, ok := got.Blocks[0].(*ast.List)
	if !ok {
		t.Fatalf("document block type = %T, want *ast.List", got.Blocks[0])
	}
	if len(outer.Items) != 1 || len(outer.Items[0].Blocks) != 2 {
		t.Fatalf("outer list shape = %#v, want one item with paragraph and nested list", outer)
	}

	outerParagraph, ok := outer.Items[0].Blocks[0].(*ast.Paragraph)
	if !ok {
		t.Fatalf("outer item block type = %T, want *ast.Paragraph", outer.Items[0].Blocks[0])
	}
	wantOuterRange := ast.Range{
		Start: ast.Position{Line: 1, Column: 3},
		End:   ast.Position{Line: 1, Column: 12},
	}
	if diff := cmp.Diff(wantOuterRange, outerParagraph.Range); diff != "" {
		t.Errorf("outer paragraph range mismatch (-want +got):\n%s", diff)
	}
	outerEmphasis, ok := outerParagraph.Content[1].(*ast.Emphasis)
	if !ok {
		t.Fatalf("outer inline type = %T, want *ast.Emphasis", outerParagraph.Content[1])
	}
	wantEmphasisRange := ast.Range{
		Start: ast.Position{Line: 1, Column: 5},
		End:   ast.Position{Line: 1, Column: 12},
	}
	if diff := cmp.Diff(wantEmphasisRange, outerEmphasis.Range); diff != "" {
		t.Errorf("outer emphasis range mismatch (-want +got):\n%s", diff)
	}
	wantOuterTextRange := ast.Range{
		Start: ast.Position{Line: 1, Column: 11},
		End:   ast.Position{Line: 1, Column: 11},
	}
	outerText, ok := outerEmphasis.Content[0].(*ast.Text)
	if !ok {
		t.Fatalf("outer nested inline type = %T, want *ast.Text", outerEmphasis.Content[0])
	}
	if diff := cmp.Diff(wantOuterTextRange, outerText.Range); diff != "" {
		t.Errorf("outer nested text range mismatch (-want +got):\n%s", diff)
	}

	nested, ok := outer.Items[0].Blocks[1].(*ast.List)
	if !ok {
		t.Fatalf("nested block type = %T, want *ast.List", outer.Items[0].Blocks[1])
	}
	if len(nested.Items) != 1 || len(nested.Items[0].Blocks) != 1 {
		t.Fatalf("nested list shape = %#v, want one item with one paragraph", nested)
	}
	nestedParagraph, ok := nested.Items[0].Blocks[0].(*ast.Paragraph)
	if !ok {
		t.Fatalf("nested item block type = %T, want *ast.Paragraph", nested.Items[0].Blocks[0])
	}
	wantNestedRange := ast.Range{
		Start: ast.Position{Line: 2, Column: 4},
		End:   ast.Position{Line: 2, Column: 18},
	}
	if diff := cmp.Diff(wantNestedRange, nestedParagraph.Range); diff != "" {
		t.Errorf("nested paragraph range mismatch (-want +got):\n%s", diff)
	}
	nestedLink, ok := nestedParagraph.Content[1].(*ast.Link)
	if !ok {
		t.Fatalf("nested inline type = %T, want *ast.Link", nestedParagraph.Content[1])
	}
	wantLinkRange := ast.Range{
		Start: ast.Position{Line: 2, Column: 6},
		End:   ast.Position{Line: 2, Column: 18},
	}
	if diff := cmp.Diff(wantLinkRange, nestedLink.Range); diff != "" {
		t.Errorf("nested link range mismatch (-want +got):\n%s", diff)
	}
	wantNestedTextRange := ast.Range{
		Start: ast.Position{Line: 2, Column: 17},
		End:   ast.Position{Line: 2, Column: 17},
	}
	nestedText, ok := nestedLink.Content[0].(*ast.Text)
	if !ok {
		t.Fatalf("nested link content type = %T, want *ast.Text", nestedLink.Content[0])
	}
	if diff := cmp.Diff(wantNestedTextRange, nestedText.Range); diff != "" {
		t.Errorf("nested link text range mismatch (-want +got):\n%s", diff)
	}
}

func TestInlineParseState_RetainsActiveParserDuringRecursion(t *testing.T) {
	spec := newSpec()
	parser := NewParser(spec)
	state := &inlineParseState{
		parser: parser,
		ctx: newInlineContext(
			":[em]{外 :[link:/x]{界}}",
			ast.Position{Line: 4, Column: 7},
		),
	}

	if state.parser != parser {
		t.Fatal("inline parse state did not retain the active Parser")
	}
	if state.parser.spec != spec {
		t.Fatal("inline parse state did not retain the active Spec")
	}

	got, next, closed, err := state.parseSeq(0, false)
	if err != nil {
		t.Fatalf("parseSeq returned an error: %v", err)
	}
	if next != len(state.ctx.text) {
		t.Errorf("parseSeq stopped at byte %d, want %d", next, len(state.ctx.text))
	}
	if closed {
		t.Error("top-level parseSeq reported a closing brace")
	}

	want := []ast.Inline{
		&ast.Emphasis{
			Content: []ast.Inline{
				&ast.Text{
					Value: "外 ",
					Range: ast.Range{
						Start: ast.Position{Line: 4, Column: 13},
						End:   ast.Position{Line: 4, Column: 14},
					},
				},
				&ast.Link{
					URI: "/x",
					Content: []ast.Inline{
						&ast.Text{
							Value: "界",
							Range: ast.Range{
								Start: ast.Position{Line: 4, Column: 26},
								End:   ast.Position{Line: 4, Column: 26},
							},
						},
					},
					Range: ast.Range{
						Start: ast.Position{Line: 4, Column: 15},
						End:   ast.Position{Line: 4, Column: 27},
					},
				},
			},
			Range: ast.Range{
				Start: ast.Position{Line: 4, Column: 7},
				End:   ast.Position{Line: 4, Column: 28},
			},
		},
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("parseSeq returned unexpected inlines (-want +got):\n%s", diff)
	}
	if state.parser != parser || state.parser.spec != spec {
		t.Error("recursive inline parsing replaced the active Parser or Spec")
	}
}

func TestParserParse_CustomSpecKeepsCodeBlockContentLiteral(t *testing.T) {
	parser, probes := newActiveSpecTestParser(t)
	input := ":::[code:txt]\n:[em]{日}\n:::"

	got, err := parser.Parse(input)
	if err != nil {
		t.Fatalf("Parse returned an error: %v", err)
	}

	want := &ast.Document{
		Blocks: []ast.Block{
			&ast.CodeBlock{
				Language: "txt",
				Text:     ":[em]{日}",
				Range: ast.Range{
					Start: ast.Position{Line: 1, Column: 1},
					End:   ast.Position{Line: 3, Column: 3},
				},
			},
		},
		Range: ast.Range{
			Start: ast.Position{Line: 1, Column: 1},
			End:   ast.Position{Line: 3, Column: 3},
		},
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Parse returned an unexpected document (-want +got):\n%s", diff)
	}
	if got := probes["paragraph"].calls; got != 0 {
		t.Errorf("paragraph builder calls = %d, want 0", got)
	}
	if got := probes["heading"].calls; got != 0 {
		t.Errorf("heading builder calls = %d, want 0", got)
	}
}

type activeParserProbeBuilder struct {
	t          *testing.T
	wantParser *Parser
	wantSpec   *Spec
	delegate   blockBuilder
	calls      int
}

type activeSpecReaderProbe struct {
	calls int
}

func (r *activeSpecReaderProbe) read(*blockContext) (parsedBlockNode, bool, error) {
	r.calls++
	return nil, false, nil
}

func (b *activeParserProbeBuilder) build(parser *Parser, node parsedBlockNode) (ast.Block, error) {
	b.t.Helper()
	b.calls++

	if parser != b.wantParser {
		b.t.Errorf("builder received Parser %p, want active Parser %p", parser, b.wantParser)
	}
	if parser == nil || parser.spec != b.wantSpec {
		b.t.Errorf("builder did not receive the Parser owning active Spec %p", b.wantSpec)
	}

	return b.delegate.build(parser, node)
}

func newActiveSpecTestParser(t *testing.T) (*Parser, map[string]*activeParserProbeBuilder) {
	t.Helper()

	spec := newSpec()
	spec.addBlockReader(&blockDirectiveReader{})
	spec.addBlockReader(&headingReader{})

	parser := NewParser(spec)
	probes := map[string]*activeParserProbeBuilder{
		"paragraph": {
			t:          t,
			wantParser: parser,
			wantSpec:   spec,
			delegate:   &paragraphBuilder{},
		},
		"heading": {
			t:          t,
			wantParser: parser,
			wantSpec:   spec,
			delegate:   &headingBuilder{},
		},
	}

	spec.addBlockBuilder("paragraph", probes["paragraph"])
	spec.addBlockBuilder("heading", probes["heading"])
	spec.addBlockBuilder("code", &codeBlockBuilder{})

	return parser, probes
}
