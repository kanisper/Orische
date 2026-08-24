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
