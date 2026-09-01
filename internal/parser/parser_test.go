package parser

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"orische/internal/ast"
	"orische/internal/diagnostic"
)

func TestBuildAST(t *testing.T) {
	buffer := []string{
		"= Heading1",
		"",
		"paragraph1 line1",
		"",
		"# ol level 1 line 1",
		"* ol level 1 line 2",
		"** ul level 2 line 1",
		"# ol level 1 line 3",
		"",
		":::[code:go]",
		"fmt.Println(\"Hello\")",
		":::",
		"",
		"== Heading2",
		"",
		"paragraph2 line1",
		"paragraph2 line2",
		"",
	}

	input := strings.Join(buffer, "\n")

	want := &ast.Document{
		Blocks: []ast.Block{
			&ast.Heading{
				Level: 1,
				Content: []ast.Inline{
					&ast.Text{
						Value: "Heading1",
						Range: ast.Range{
							Start: ast.Position{Line: 1, Column: 3},
							End:   ast.Position{Line: 1, Column: 10},
						},
					},
				},
				Range: ast.Range{
					Start: ast.Position{Line: 1, Column: 1},
					End:   ast.Position{Line: 1, Column: 10},
				},
			},
			&ast.Paragraph{
				Content: []ast.Inline{
					&ast.Text{
						Value: "paragraph1 line1",
						Range: ast.Range{
							Start: ast.Position{Line: 3, Column: 1},
							End:   ast.Position{Line: 3, Column: 16},
						},
					},
				},
				Range: ast.Range{
					Start: ast.Position{Line: 3, Column: 1},
					End:   ast.Position{Line: 3, Column: 16},
				},
			},
			&ast.List{
				Ordered: true,
				Items: []*ast.ListItem{
					{
						Blocks: []ast.Block{
							&ast.Paragraph{
								Content: []ast.Inline{
									&ast.Text{
										Value: "ol level 1 line 1",
										Range: ast.Range{
											Start: ast.Position{Line: 5, Column: 3},
											End:   ast.Position{Line: 5, Column: 19},
										},
									},
								},
								Range: ast.Range{
									Start: ast.Position{Line: 5, Column: 3},
									End:   ast.Position{Line: 5, Column: 19},
								},
							},
						},
						Range: ast.Range{
							Start: ast.Position{Line: 5, Column: 1},
							End:   ast.Position{Line: 5, Column: 19},
						},
					},
					{
						Blocks: []ast.Block{
							&ast.Paragraph{
								Content: []ast.Inline{
									&ast.Text{
										Value: "ol level 1 line 2",
										Range: ast.Range{
											Start: ast.Position{Line: 6, Column: 3},
											End:   ast.Position{Line: 6, Column: 19},
										},
									},
								},
								Range: ast.Range{
									Start: ast.Position{Line: 6, Column: 3},
									End:   ast.Position{Line: 6, Column: 19},
								},
							},
							&ast.List{
								Ordered: false,
								Items: []*ast.ListItem{
									{
										Blocks: []ast.Block{
											&ast.Paragraph{
												Content: []ast.Inline{
													&ast.Text{
														Value: "ul level 2 line 1",
														Range: ast.Range{
															Start: ast.Position{Line: 7, Column: 4},
															End:   ast.Position{Line: 7, Column: 20},
														},
													},
												},
												Range: ast.Range{
													Start: ast.Position{Line: 7, Column: 4},
													End:   ast.Position{Line: 7, Column: 20},
												},
											},
										},
										Range: ast.Range{
											Start: ast.Position{Line: 7, Column: 1},
											End:   ast.Position{Line: 7, Column: 20},
										},
									},
								},
								Range: ast.Range{
									Start: ast.Position{Line: 7, Column: 1},
									End:   ast.Position{Line: 7, Column: 20},
								},
							},
						},
						Range: ast.Range{
							Start: ast.Position{Line: 6, Column: 1},
							End:   ast.Position{Line: 7, Column: 20},
						},
					},
					{
						Blocks: []ast.Block{
							&ast.Paragraph{
								Content: []ast.Inline{
									&ast.Text{
										Value: "ol level 1 line 3",
										Range: ast.Range{
											Start: ast.Position{Line: 8, Column: 3},
											End:   ast.Position{Line: 8, Column: 19},
										},
									},
								},
								Range: ast.Range{
									Start: ast.Position{Line: 8, Column: 3},
									End:   ast.Position{Line: 8, Column: 19},
								},
							},
						},
						Range: ast.Range{
							Start: ast.Position{Line: 8, Column: 1},
							End:   ast.Position{Line: 8, Column: 19},
						},
					},
				},
				Range: ast.Range{
					Start: ast.Position{Line: 5, Column: 1},
					End:   ast.Position{Line: 8, Column: 19},
				},
			},
			&ast.CodeBlock{
				Language: "go",
				Text:     "fmt.Println(\"Hello\")",
				Range: ast.Range{
					Start: ast.Position{Line: 10, Column: 1},
					End:   ast.Position{Line: 12, Column: 3},
				},
			},
			&ast.Heading{
				Level: 2,
				Content: []ast.Inline{
					&ast.Text{
						Value: "Heading2",
						Range: ast.Range{
							Start: ast.Position{Line: 14, Column: 4},
							End:   ast.Position{Line: 14, Column: 11},
						},
					},
				},
				Range: ast.Range{
					Start: ast.Position{Line: 14, Column: 1},
					End:   ast.Position{Line: 14, Column: 11},
				},
			},
			&ast.Paragraph{
				Content: []ast.Inline{
					&ast.Text{
						Value: "paragraph2 line1",
						Range: ast.Range{
							Start: ast.Position{Line: 16, Column: 1},
							End:   ast.Position{Line: 16, Column: 16},
						},
					},
					&ast.Text{
						Value: "paragraph2 line2",
						Range: ast.Range{
							Start: ast.Position{Line: 17, Column: 1},
							End:   ast.Position{Line: 17, Column: 16},
						},
					},
				},
				Range: ast.Range{
					Start: ast.Position{Line: 16, Column: 1},
					End:   ast.Position{Line: 17, Column: 16},
				},
			},
		},
		Range: ast.Range{
			Start: ast.Position{Line: 1, Column: 1},
			End:   ast.Position{Line: 17, Column: 16},
		},
	}

	got, err := Parse(input)
	if err != nil {
		t.Fatalf("parse returned an error: %v", err)
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("parse incorrectly.\n(-want +got)\n%s", diff)
	}
}

func TestUninitializedParserReturnsError(t *testing.T) {
	tests := []*Parser{nil, {}}
	for _, p := range tests {
		got, err := p.Parse("text")
		if got != nil {
			t.Errorf("Parse returned a document: %#v", got)
		}
		if err == nil || err.Error() != "parser is not initialized; use NewParser" {
			t.Errorf("Parse error = %v, want uninitialized parser error", err)
		}
	}
}

func TestUnsupportedDirective(t *testing.T) {
	input := `:::[Unsupported]
content
:::`

	_, err := Parse(input)

	var diag *diagnostic.Error
	if !errors.As(err, &diag) {
		t.Fatalf("expected diagnostic error, but got %v", err)
	}

	if diag.Message != "unsupported block directive type \"unsupported\"" {
		t.Errorf("expected message: unsupported directive type \"unsupported\", but got: %s", diag.Message)
	}

	if diag.Range.Start.Line != 1 {
		t.Errorf("expected start line: 1, but got: %d", diag.Range.Start.Line)
	}
}

func TestDefaultParserUsesCoreInlineDefinitions(t *testing.T) {
	input := "= 日 :[em]{見出し}\n\n* :[link:/x]{項目}\n** :[code]{内}\n\n:::[code:go]\n:[em]{literal}\n:::"

	want, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse returned an error: %v", err)
	}
	p := NewParser()
	got, err := p.Parse(input)
	if err != nil {
		t.Fatalf("explicit core Parse returned an error: %v", err)
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("default and explicit core differ (-default +explicit):\n%s", diff)
	}
}

func TestParseOneBlockUsesDirectiveSugarParagraphOrder(t *testing.T) {
	p := mustCoreParser(t)

	node, consumed, err := p.parseOneBlock(&blockContext{
		lines: []string{":::[code]", "= source", ":::"},
	})
	if err != nil {
		t.Fatalf("parseOneBlock returned an error: %v", err)
	}
	if consumed != 3 {
		t.Fatalf("directive consumed %d lines, want 3", consumed)
	}
	if _, ok := node.(*blockDirectiveNode); !ok {
		t.Fatalf("directive node type = %T, want *blockDirectiveNode", node)
	}

	node, consumed, err = p.parseOneBlock(&blockContext{lines: []string{"== heading"}})
	if err != nil {
		t.Fatalf("parseOneBlock returned an error: %v", err)
	}
	if consumed != 1 {
		t.Fatalf("heading consumed %d lines, want 1", consumed)
	}
	if _, ok := node.(*headingNode); !ok {
		t.Fatalf("heading node type = %T, want *headingNode", node)
	}

	node, consumed, err = p.parseOneBlock(&blockContext{lines: []string{"plain"}})
	if err != nil {
		t.Fatalf("parseOneBlock returned an error: %v", err)
	}
	if consumed != 1 {
		t.Fatalf("paragraph consumed %d lines, want 1", consumed)
	}
	if _, ok := node.(*paragraphNode); !ok {
		t.Fatalf("paragraph node type = %T, want *paragraphNode", node)
	}
}

func TestListRunReturnsFollowingLineToFrontend(t *testing.T) {
	got, err := Parse("* item\nplain")
	if err != nil {
		t.Fatalf("Parse returned an error: %v", err)
	}
	if len(got.Blocks) != 2 {
		t.Fatalf("Block count = %d, want list and paragraph", len(got.Blocks))
	}
	if _, ok := got.Blocks[0].(*ast.List); !ok {
		t.Errorf("Block 0 type = %T, want *ast.List", got.Blocks[0])
	}
	paragraph, ok := got.Blocks[1].(*ast.Paragraph)
	if !ok || len(paragraph.Content) != 1 {
		t.Fatalf("Block 1 = %#v, want one-content Paragraph", got.Blocks[1])
	}
	text, ok := paragraph.Content[0].(*ast.Text)
	if !ok || text.Value != "plain" || text.Range.Start.Line != 2 {
		t.Errorf("following Paragraph content = %#v, want plain on line 2", paragraph.Content[0])
	}
}

func TestMalformedBlockCandidatesFallBackToParagraph(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		visible string
	}{
		{name: "unterminated directive", input: ":::[code]\ntext", visible: ":::[code]text"},
		{name: "invalid heading", input: "======= title", visible: "======= title"},
		{name: "invalid list", input: "- item", visible: "- item"},
		{
			name:    "later block markers inside paragraph",
			input:   "plain\n= heading\n:::[code]\ntext\n:::",
			visible: "plain= heading:::[code]text:::",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := Parse(tt.input)
			if err != nil {
				t.Fatalf("Parse returned an error: %v", err)
			}
			if len(doc.Blocks) != 1 {
				t.Fatalf("Block count = %d, want one Paragraph", len(doc.Blocks))
			}
			paragraph, ok := doc.Blocks[0].(*ast.Paragraph)
			if !ok {
				t.Fatalf("Block type = %T, want *ast.Paragraph", doc.Blocks[0])
			}
			var visible string
			for _, inline := range paragraph.Content {
				text, ok := inline.(*ast.Text)
				if !ok {
					t.Fatalf("Paragraph inline type = %T, want *ast.Text", inline)
				}
				visible += text.Value
			}
			if visible != tt.visible {
				t.Errorf("visible Paragraph content = %q, want %q", visible, tt.visible)
			}
		})
	}
}

func TestFallbackHandsNextBlankSeparatedBlockToFrontend(t *testing.T) {
	doc, err := Parse(":::[code]\nunterminated\n\n= valid")
	if err != nil {
		t.Fatalf("Parse returned an error: %v", err)
	}
	if len(doc.Blocks) != 2 {
		t.Fatalf("Block count = %d, want fallback Paragraph and Heading", len(doc.Blocks))
	}
	if _, ok := doc.Blocks[0].(*ast.Paragraph); !ok {
		t.Errorf("Block 0 type = %T, want *ast.Paragraph", doc.Blocks[0])
	}
	heading, ok := doc.Blocks[1].(*ast.Heading)
	if !ok || heading.Range.Start.Line != 4 {
		t.Errorf("Block 1 = %#v, want Heading starting on line 4", doc.Blocks[1])
	}
}
