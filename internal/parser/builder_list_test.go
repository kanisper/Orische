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
							StartLine: 1,
							EndLine:   1,
						},
					},
				},
				Range: ast.Range{
					StartLine: 1,
					EndLine:   1,
				},
			},
			{
				Blocks: []parsedBlockNode{
					&parsedBlock{
						Type: "Paragraph",
						Attr: "",
						Text: "ol level 1 line 2",
						Range: ast.Range{
							StartLine: 2,
							EndLine:   2,
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
											StartLine: 3,
											EndLine:   3,
										},
									},
								},
								Range: ast.Range{
									StartLine: 3,
									EndLine:   3,
								},
							},
						},
						Range: ast.Range{
							StartLine: 3,
							EndLine:   3,
						},
					},
				},
				Range: ast.Range{
					StartLine: 2,
					EndLine:   3,
				},
			},
			{
				Blocks: []parsedBlockNode{
					&parsedBlock{
						Type: "Paragraph",
						Attr: "",
						Text: "ol level 1 line 3",
						Range: ast.Range{
							StartLine: 4,
							EndLine:   4,
						},
					},
				},
				Range: ast.Range{
					StartLine: 4,
					EndLine:   4,
				},
			},
		},
		Range: ast.Range{
			StartLine: 1,
			EndLine:   4,
		},
	}

	want := &ast.List{
		Ordered: true,
		Items: []*ast.ListItem{
			{
				Blocks: []ast.Block{
					&ast.Paragraph{
						Content: []ast.Inline{
							&ast.Text{Value: "ol level 1 line 1"},
						},
						Range: ast.Range{StartLine: 1, EndLine: 1},
					},
				},
				Range: ast.Range{StartLine: 1, EndLine: 1},
			},
			{
				Blocks: []ast.Block{
					&ast.Paragraph{
						Content: []ast.Inline{
							&ast.Text{Value: "ol level 1 line 2"},
						},
						Range: ast.Range{StartLine: 2, EndLine: 2},
					},
					&ast.List{
						Ordered: false,
						Items: []*ast.ListItem{
							{
								Blocks: []ast.Block{
									&ast.Paragraph{
										Content: []ast.Inline{
											&ast.Text{Value: "ul level 2 line 1"},
										},
										Range: ast.Range{StartLine: 3, EndLine: 3},
									},
								},
								Range: ast.Range{StartLine: 3, EndLine: 3},
							},
						},
						Range: ast.Range{StartLine: 3, EndLine: 3},
					},
				},
				Range: ast.Range{StartLine: 2, EndLine: 3},
			},
			{
				Blocks: []ast.Block{
					&ast.Paragraph{
						Content: []ast.Inline{
							&ast.Text{Value: "ol level 1 line 3"},
						},
						Range: ast.Range{StartLine: 4, EndLine: 4},
					},
				},
				Range: ast.Range{StartLine: 4, EndLine: 4},
			},
		},
		Range: ast.Range{StartLine: 1, EndLine: 4},
	}

	got, err := (&listBuilder{}).build(input)

	if err != nil {
		t.Errorf("build failed: %v", err)
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("build incorrectly.\n(-want +got)\n%s", diff)
	}
}
