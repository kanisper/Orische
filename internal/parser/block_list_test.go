package parser

import (
	"testing"

	"orische/internal/ast"

	"github.com/google/go-cmp/cmp"
)

func TestParseList(t *testing.T) {
	input := &blockContext{
		lines: []string{
			"# ol level 1 line 1",
			"* ol level 1 line 2",
			"** ul level 2 line 1",
			"# ol level 1 line 3",
		},
		pos: 0,
	}

	want := &parsedList{
		Ordered: true,
		Items: []parsedListItem{
			{
				Blocks: []parsedBlockNode{
					&parsedBlock{
						Type: "Paragraph",
						Attr: "",
						Text: "ol level 1 line 1",
						Range: ast.Range{
							Start: ast.Position{Line: 1, Column: 1},
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
							Start: ast.Position{Line: 2, Column: 1},
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
							Start: ast.Position{Line: 4, Column: 1},
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

	ctx_pos_want := 3

	output, ok, err := (&listParser{}).parse(input)

	if !ok || err != nil {
		t.Errorf("parse failed.")
	}
	if diff := cmp.Diff(want, output); diff != "" {
		t.Errorf("parse incorrectly.\n(-want +got)\n%s", diff)
	}
	if input.pos != ctx_pos_want {
		t.Errorf("position in context is not updated correctly. want %d, got %d", ctx_pos_want, input.pos)
	}
}

func TestParseList_NormalizesRawNestingLevelJump(t *testing.T) {
	input := &blockContext{
		lines: []string{
			"* parent",
			"*** child",
		},
		pos: 0,
	}

	want := &parsedList{
		Ordered: false,
		Items: []parsedListItem{
			{
				Blocks: []parsedBlockNode{
					&parsedBlock{
						Type: "Paragraph",
						Attr: "",
						Text: "parent",
						Range: ast.Range{
							Start: ast.Position{Line: 1, Column: 1},
							End:   ast.Position{Line: 1, Column: 8},
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
										Text: "child",
										Range: ast.Range{
											Start: ast.Position{Line: 2, Column: 1},
											End:   ast.Position{Line: 2, Column: 9},
										},
									},
								},
								Range: ast.Range{
									Start: ast.Position{Line: 2, Column: 1},
									End:   ast.Position{Line: 2, Column: 9},
								},
							},
						},
						Range: ast.Range{
							Start: ast.Position{Line: 2, Column: 1},
							End:   ast.Position{Line: 2, Column: 9},
						},
					},
				},
				Range: ast.Range{
					Start: ast.Position{Line: 1, Column: 1},
					End:   ast.Position{Line: 2, Column: 9},
				},
			},
		},
		Range: ast.Range{
			Start: ast.Position{Line: 1, Column: 1},
			End:   ast.Position{Line: 2, Column: 9},
		},
	}

	output, ok, err := (&listParser{}).parse(input)
	if err != nil {
		t.Fatalf("parse returned an error: %v", err)
	}
	if !ok {
		t.Fatal("list with a raw nesting level jump was not parsed")
	}
	if diff := cmp.Diff(want, output); diff != "" {
		t.Errorf("parse incorrectly.\n(-want +got)\n%s", diff)
	}
	if input.pos != 1 {
		t.Errorf("position in context is not updated correctly. want 1, got %d", input.pos)
	}
}
