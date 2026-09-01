package parser

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"orische/internal/ast"
)

func TestReadHeading(t *testing.T) {
	tests := []struct {
		name string
		line string
		want *headingNode
	}{
		{
			name: "level one unicode",
			line: "= 見出し😀",
			want: &headingNode{
				level:         1,
				text:          "見出し😀",
				contentOrigin: ast.Position{Line: 1, Column: 3},
				rng: ast.Range{
					Start: ast.Position{Line: 1, Column: 1},
					End:   ast.Position{Line: 1, Column: 6},
				},
			},
		},
		{
			name: "level six",
			line: "====== title",
			want: &headingNode{
				level:         6,
				text:          "title",
				contentOrigin: ast.Position{Line: 1, Column: 8},
				rng: ast.Range{
					Start: ast.Position{Line: 1, Column: 1},
					End:   ast.Position{Line: 1, Column: 12},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, consumed := readHeading(&blockContext{lines: []string{tt.line}})
			if consumed != 1 {
				t.Fatalf("readHeading consumed %d lines, want 1", consumed)
			}
			if diff := cmp.Diff(tt.want, got, cmp.AllowUnexported(headingNode{})); diff != "" {
				t.Errorf("heading mismatch (-want +got):\n%s", diff)
			}
		})
	}

	for _, line := range []string{"======= title", "=title", "="} {
		t.Run("reject "+line, func(t *testing.T) {
			got, consumed := readHeading(&blockContext{lines: []string{line}})
			if got != nil || consumed != 0 {
				t.Errorf("readHeading(%q) = (%#v, %d), want (nil, 0)", line, got, consumed)
			}
		})
	}
}

func TestHeadingKeepsConcreteLevel(t *testing.T) {
	node, consumed := readHeading(&blockContext{
		lines: []string{"== 見出し😀"},
	})
	if consumed != 1 {
		t.Fatalf("readHeading consumed %d lines, want 1", consumed)
	}
	heading, ok := node.(*headingNode)
	if !ok {
		t.Fatalf("readHeading returned %T, want *headingNode", node)
	}
	if heading.level != 2 {
		t.Errorf("heading level = %d, want 2", heading.level)
	}
	if heading.text != "見出し😀" {
		t.Errorf("heading text = %q, want 見出し😀", heading.text)
	}
	if heading.contentOrigin != (ast.Position{Line: 1, Column: 4}) {
		t.Errorf("heading content origin = %#v, want line 1 column 4", heading.contentOrigin)
	}
}

func TestNormalizeListLevel(t *testing.T) {
	tests := []struct {
		previousRaw     int
		previousLogical int
		currentRaw      int
		want            int
	}{
		{0, 0, 3, 1},
		{1, 1, 4, 2},
		{4, 2, 4, 2},
		{4, 3, 2, 1},
		{5, 2, 1, 1},
	}

	for _, tt := range tests {
		if got := normalizeListLevel(tt.previousRaw, tt.previousLogical, tt.currentRaw); got != tt.want {
			t.Errorf("normalizeListLevel(%d, %d, %d) = %d, want %d", tt.previousRaw, tt.previousLogical, tt.currentRaw, got, tt.want)
		}
	}
}

func TestReadListBuildsPrivateRecursiveNodes(t *testing.T) {
	input := &blockContext{
		lines: []string{"before", "* 親😀", "**** child", "# sibling", "not a list"},
		start: 1,
	}

	gotNode, consumed := readList(input)
	if consumed != 3 {
		t.Fatalf("readList consumed %d lines, want 3", consumed)
	}
	got, ok := gotNode.(*listNode)
	if !ok {
		t.Fatalf("node type = %T, want *listNode", gotNode)
	}
	want := &listNode{
		ordered: false,
		items: []listItemNode{
			{
				blocks: []parsedBlock{
					&paragraphNode{
						text:          "親😀\n",
						contentOrigin: ast.Position{Line: 2, Column: 3},
						rng: ast.Range{
							Start: ast.Position{Line: 2, Column: 3},
							End:   ast.Position{Line: 2, Column: 4},
						},
					},
					&listNode{
						ordered: false,
						items: []listItemNode{
							{
								blocks: []parsedBlock{
									&paragraphNode{
										text:          "child\n",
										contentOrigin: ast.Position{Line: 3, Column: 6},
										rng: ast.Range{
											Start: ast.Position{Line: 3, Column: 6},
											End:   ast.Position{Line: 3, Column: 10},
										},
									},
								},
								rng: ast.Range{Start: ast.Position{Line: 3, Column: 1}, End: ast.Position{Line: 3, Column: 10}},
							},
						},
						rng: ast.Range{Start: ast.Position{Line: 3, Column: 1}, End: ast.Position{Line: 3, Column: 10}},
					},
				},
				rng: ast.Range{Start: ast.Position{Line: 2, Column: 1}, End: ast.Position{Line: 3, Column: 10}},
			},
			{
				blocks: []parsedBlock{
					&paragraphNode{
						text:          "sibling\n",
						contentOrigin: ast.Position{Line: 4, Column: 3},
						rng:           ast.Range{Start: ast.Position{Line: 4, Column: 3}, End: ast.Position{Line: 4, Column: 9}},
					},
				},
				rng: ast.Range{Start: ast.Position{Line: 4, Column: 1}, End: ast.Position{Line: 4, Column: 9}},
			},
		},
		rng: ast.Range{Start: ast.Position{Line: 2, Column: 1}, End: ast.Position{Line: 4, Column: 9}},
	}
	if diff := cmp.Diff(want, got, cmp.AllowUnexported(listNode{}, listItemNode{}, paragraphNode{})); diff != "" {
		t.Errorf("list mismatch (-want +got):\n%s", diff)
	}
}

func TestBuildCodeBlock(t *testing.T) {
	rng := ast.Range{Start: ast.Position{Line: 2, Column: 3}, End: ast.Position{Line: 2, Column: 6}}
	got, err := buildCodeBlock(nil, &blockDirectiveNode{attribute: "go", text: "x := 1", rng: rng})
	if err != nil {
		t.Fatalf("buildCodeBlock returned an error: %v", err)
	}
	want := &ast.CodeBlock{Language: "go", Text: "x := 1", Range: rng}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("code mismatch (-want +got):\n%s", diff)
	}
}
