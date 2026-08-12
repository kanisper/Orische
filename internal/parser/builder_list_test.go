package parser

import (
	"testing"

	"orische/internal/ast"

	"github.com/google/go-cmp/cmp"
)

func TestBuildList(t *testing.T) {
	input := &parsedList{
		Ordered: true,
		Items: []parsedListItem{
			{
				Blocks: []parsedBlockNode{
					&parsedBlock{
						Type: "Paragraph",
						Attr: "",
						Text: "ol level 1 line 1",
						Range: ast.Range{
							Start: ast.Position{Line: 1, Column: 3},
							End:   ast.Position{Line: 1, Column: 19},
						},
					},
				},
				Range: ast.Range{
					Start: ast.Position{Line: 1, Column: 1},
					End:   ast.Position{Line: 1, Column: 19},
				},
			},
			{
				Blocks: []parsedBlockNode{
					&parsedBlock{
						Type: "Paragraph",
						Attr: "",
						Text: "ol level 1 line 2",
						Range: ast.Range{
							Start: ast.Position{Line: 2, Column: 3},
							End:   ast.Position{Line: 2, Column: 19},
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
											Start: ast.Position{Line: 3, Column: 4},
											End:   ast.Position{Line: 3, Column: 20},
										},
									},
								},
								Range: ast.Range{
									Start: ast.Position{Line: 3, Column: 1},
									End:   ast.Position{Line: 3, Column: 20},
								},
							},
						},
						Range: ast.Range{
							Start: ast.Position{Line: 3, Column: 1},
							End:   ast.Position{Line: 3, Column: 20},
						},
					},
				},
				Range: ast.Range{
					Start: ast.Position{Line: 2, Column: 1},
					End:   ast.Position{Line: 3, Column: 20},
				},
			},
			{
				Blocks: []parsedBlockNode{
					&parsedBlock{
						Type: "Paragraph",
						Attr: "",
						Text: "ol level 1 line 3",
						Range: ast.Range{
							Start: ast.Position{Line: 4, Column: 3},
							End:   ast.Position{Line: 4, Column: 19},
						},
					},
				},
				Range: ast.Range{
					Start: ast.Position{Line: 4, Column: 1},
					End:   ast.Position{Line: 4, Column: 19},
				},
			},
		},
		Range: ast.Range{
			Start: ast.Position{Line: 1, Column: 1},
			End:   ast.Position{Line: 4, Column: 19},
		},
	}

	want := &ast.List{
		Ordered: true,
		Items: []*ast.ListItem{
			{
				Blocks: []ast.Block{
					&ast.Paragraph{
						Content: []ast.Inline{
							&ast.Text{
								Value: "ol level 1 line 1",
								Range: ast.Range{
									Start: ast.Position{Line: 1, Column: 3},
									End:   ast.Position{Line: 1, Column: 19},
								},
							},
						},
						Range: ast.Range{
							Start: ast.Position{Line: 1, Column: 3},
							End:   ast.Position{Line: 1, Column: 19},
						},
					},
				},
				Range: ast.Range{
					Start: ast.Position{Line: 1, Column: 1},
					End:   ast.Position{Line: 1, Column: 19},
				},
			},
			{
				Blocks: []ast.Block{
					&ast.Paragraph{
						Content: []ast.Inline{
							&ast.Text{
								Value: "ol level 1 line 2",
								Range: ast.Range{
									Start: ast.Position{Line: 2, Column: 3},
									End:   ast.Position{Line: 2, Column: 19},
								},
							},
						},
						Range: ast.Range{
							Start: ast.Position{Line: 2, Column: 3},
							End:   ast.Position{Line: 2, Column: 19},
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
													Start: ast.Position{Line: 3, Column: 4},
													End:   ast.Position{Line: 3, Column: 20},
												},
											},
										},
										Range: ast.Range{
											Start: ast.Position{Line: 3, Column: 4},
											End:   ast.Position{Line: 3, Column: 20},
										},
									},
								},
								Range: ast.Range{
									Start: ast.Position{Line: 3, Column: 1},
									End:   ast.Position{Line: 3, Column: 20},
								},
							},
						},
						Range: ast.Range{
							Start: ast.Position{Line: 3, Column: 1},
							End:   ast.Position{Line: 3, Column: 20},
						},
					},
				},
				Range: ast.Range{
					Start: ast.Position{Line: 2, Column: 1},
					End:   ast.Position{Line: 3, Column: 20},
				},
			},
			{
				Blocks: []ast.Block{
					&ast.Paragraph{
						Content: []ast.Inline{
							&ast.Text{
								Value: "ol level 1 line 3",
								Range: ast.Range{
									Start: ast.Position{Line: 4, Column: 3},
									End:   ast.Position{Line: 4, Column: 19},
								},
							},
						},
						Range: ast.Range{
							Start: ast.Position{Line: 4, Column: 3},
							End:   ast.Position{Line: 4, Column: 19},
						},
					},
				},
				Range: ast.Range{
					Start: ast.Position{Line: 4, Column: 1},
					End:   ast.Position{Line: 4, Column: 19},
				},
			},
		},
		Range: ast.Range{
			Start: ast.Position{Line: 1, Column: 1},
			End:   ast.Position{Line: 4, Column: 19},
		},
	}

	got, err := (&listBuilder{}).build(input)

	if err != nil {
		t.Fatalf("build returned an error: %v", err)
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("build incorrectly.\n(-want +got)\n%s", diff)
	}
}
