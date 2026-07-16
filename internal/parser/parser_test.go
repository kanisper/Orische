package parser

import (
	"strings"
	"testing"

	"orische/internal/ast"

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
					StartLine: 1,
					EndLine:   1,
				},
			},
			&parsedBlock{
				Type: "Paragraph",
				Attr: "",
				Text: "paragraph1 line1",
				Range: ast.Range{
					StartLine: 3,
					EndLine:   3,
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
									StartLine: 5,
									EndLine:   5,
								},
							},
						},
						Range: ast.Range{
							StartLine: 5,
							EndLine:   5,
						},
					},
					{
						Blocks: []parsedBlockNode{
							&parsedBlock{
								Type: "Paragraph",
								Attr: "",
								Text: "ol level 1 line 2",
								Range: ast.Range{
									StartLine: 6,
									EndLine:   6,
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
													StartLine: 7,
													EndLine:   7,
												},
											},
										},
										Range: ast.Range{
											StartLine: 7,
											EndLine:   7,
										},
									},
								},
								Range: ast.Range{
									StartLine: 7,
									EndLine:   7,
								},
							},
						},
						Range: ast.Range{
							StartLine: 6,
							EndLine:   7,
						},
					},
					{
						Blocks: []parsedBlockNode{
							&parsedBlock{
								Type: "Paragraph",
								Attr: "",
								Text: "ol level 1 line 3",
								Range: ast.Range{
									StartLine: 8,
									EndLine:   8,
								},
							},
						},
						Range: ast.Range{
							StartLine: 8,
							EndLine:   8,
						},
					},
				},
				Range: ast.Range{
					StartLine: 5,
					EndLine:   8,
				},
			},
			&parsedBlock{
				Type: "code",
				Attr: "go",
				Text: "fmt.Println(\"Hello\")",
				Range: ast.Range{
					StartLine: 10,
					EndLine:   12,
				},
			},
			&parsedBlock{
				Type: "Heading",
				Attr: "level2",
				Text: "Heading2",
				Range: ast.Range{
					StartLine: 14,
					EndLine:   14,
				},
			},
			&parsedBlock{
				Type: "Paragraph",
				Attr: "",
				Text: "paragraph2 line1\nparagraph2 line2",
				Range: ast.Range{
					StartLine: 16,
					EndLine:   17,
				},
			},
		},
		Range: ast.Range{
			StartLine: 1,
			EndLine:   17,
		},
	}

	output, err := NewParser(coreSpec()).parseDocument(input)

	if err != nil {
		t.Errorf("parse failed.\n%s", err)
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
					&ast.Text{Value: "Heading1"},
				},
				Range: ast.Range{StartLine: 1, EndLine: 1},
			},
			&ast.Paragraph{
				Content: []ast.Inline{
					&ast.Text{Value: "paragraph1 line1"},
				},
				Range: ast.Range{StartLine: 3, EndLine: 3},
			},
			&ast.List{
				Ordered: true,
				Items: []*ast.ListItem{
					{
						Blocks: []ast.Block{
							&ast.Paragraph{
								Content: []ast.Inline{
									&ast.Text{Value: "ol level 1 line 1"},
								},
								Range: ast.Range{StartLine: 5, EndLine: 5},
							},
						},
						Range: ast.Range{StartLine: 5, EndLine: 5},
					},
					{
						Blocks: []ast.Block{
							&ast.Paragraph{
								Content: []ast.Inline{
									&ast.Text{Value: "ol level 1 line 2"},
								},
								Range: ast.Range{StartLine: 6, EndLine: 6},
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
												Range: ast.Range{StartLine: 7, EndLine: 7},
											},
										},
										Range: ast.Range{StartLine: 7, EndLine: 7},
									},
								},
								Range: ast.Range{StartLine: 7, EndLine: 7},
							},
						},
						Range: ast.Range{StartLine: 6, EndLine: 7},
					},
					{
						Blocks: []ast.Block{
							&ast.Paragraph{
								Content: []ast.Inline{
									&ast.Text{Value: "ol level 1 line 3"},
								},
								Range: ast.Range{StartLine: 8, EndLine: 8},
							},
						},
						Range: ast.Range{StartLine: 8, EndLine: 8},
					},
				},
				Range: ast.Range{StartLine: 5, EndLine: 8},
			},
			&ast.CodeBlock{
				Language: "go",
				Text:     "fmt.Println(\"Hello\")",
				Range:    ast.Range{StartLine: 10, EndLine: 12},
			},
			&ast.Heading{
				Level: 2,
				Content: []ast.Inline{
					&ast.Text{Value: "Heading2"},
				},
				Range: ast.Range{StartLine: 14, EndLine: 14},
			},
			&ast.Paragraph{
				Content: []ast.Inline{
					&ast.Text{Value: "paragraph2 line1\nparagraph2 line2"},
				},
				Range: ast.Range{StartLine: 16, EndLine: 17},
			},
		},
		Range: ast.Range{StartLine: 1, EndLine: 17},
	}

	got, err := Parse(input)
	if err != nil {
		t.Errorf("parse failed: %s", err)
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("parse incorrectly.\n(-want +got)\n%s", diff)
	}
}
