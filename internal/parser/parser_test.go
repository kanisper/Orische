package parser

import (
	"errors"
	"strings"
	"testing"

	"orische/internal/ast"
	"orische/internal/diagnostic"

	"github.com/google/go-cmp/cmp"
)

func TestParseDocument(t *testing.T) {
	input := []string{
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
	}

	want := &parsedDocument{
		Blocks: []parsedBlockNode{
			&parsedBlock{
				Type: "Heading",
				Attr: "level1",
				Text: "Heading1",
				Range: ast.Range{
					Start: ast.Position{Line: 1, Column: 1},
					End:   ast.Position{Line: 1, Column: 10},
				},
			},
			&parsedBlock{
				Type: "Paragraph",
				Attr: "",
				Text: "paragraph1 line1",
				Range: ast.Range{
					Start: ast.Position{Line: 3, Column: 1},
					End:   ast.Position{Line: 3, Column: 16},
				},
			},
			&parsedList{
				Ordered: true,
				Items: []parsedListItem{
					{
						Blocks: []parsedBlockNode{
							&parsedBlock{
								Type: "Paragraph",
								Attr: "",
								Text: "ol level 1 line 1",
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
						Blocks: []parsedBlockNode{
							&parsedBlock{
								Type: "Paragraph",
								Attr: "",
								Text: "ol level 1 line 2",
								Range: ast.Range{
									Start: ast.Position{Line: 6, Column: 3},
									End:   ast.Position{Line: 6, Column: 19},
								},
							},
							&parsedList{
								Ordered: false,
								Items: []parsedListItem{
									{
										Blocks: []parsedBlockNode{
											&parsedBlock{
												Type: "Paragraph",
												Attr: "",
												Text: "ul level 2 line 1",
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
						Blocks: []parsedBlockNode{
							&parsedBlock{
								Type: "Paragraph",
								Attr: "",
								Text: "ol level 1 line 3",
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
			&parsedBlock{
				Type: "code",
				Attr: "go",
				Text: "fmt.Println(\"Hello\")",
				Range: ast.Range{
					Start: ast.Position{Line: 10, Column: 1},
					End:   ast.Position{Line: 12, Column: 3},
				},
			},
			&parsedBlock{
				Type: "Heading",
				Attr: "level2",
				Text: "Heading2",
				Range: ast.Range{
					Start: ast.Position{Line: 14, Column: 1},
					End:   ast.Position{Line: 14, Column: 11},
				},
			},
			&parsedBlock{
				Type: "Paragraph",
				Attr: "",
				Text: "paragraph2 line1\nparagraph2 line2",
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

	output, err := NewParser(coreSpec()).parseDocument(input)

	if err != nil {
		t.Fatalf("parse returned an error: %v", err)
	}
	if diff := cmp.Diff(want, output); diff != "" {
		t.Errorf("parse incorrectly.\n(-want +got)\n%s", diff)
	}
}

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
						Value: "paragraph2 line1\nparagraph2 line2",
						Range: ast.Range{
							Start: ast.Position{Line: 16, Column: 1},
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
