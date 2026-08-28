package block

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"

	"orische/internal/ast"
	"orische/internal/parser/feature"
)

func TestHeadingDefinitionReadBlock(t *testing.T) {
	tests := []struct {
		name    string
		line    string
		matched bool
		want    *feature.TextBlock
	}{
		{
			name:    "level one unicode",
			line:    "= 見出し😀",
			matched: true,
			want: &feature.TextBlock{
				Type:          typeHeading,
				Attr:          "level1",
				Text:          "見出し😀",
				ContentOrigin: ast.Position{Line: 4, Column: 3},
				Range: ast.Range{
					Start: ast.Position{Line: 4, Column: 1},
					End:   ast.Position{Line: 4, Column: 6},
				},
			},
		},
		{
			name:    "level six",
			line:    "====== title",
			matched: true,
			want: &feature.TextBlock{
				Type:          typeHeading,
				Attr:          "level6",
				Text:          "title",
				ContentOrigin: ast.Position{Line: 4, Column: 8},
				Range: ast.Range{
					Start: ast.Position{Line: 4, Column: 1},
					End:   ast.Position{Line: 4, Column: 12},
				},
			},
		},
		{name: "too deep", line: "======= title"},
		{name: "no space", line: "=title"},
		{name: "no text", line: "="},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := (&headingDefinition{}).ReadBlock(testBlockInput{
				{Number: 4, Text: tt.line},
			})
			if err != nil {
				t.Fatalf("ReadBlock returned an error: %v", err)
			}
			if got.Matched != tt.matched {
				t.Fatalf("Matched = %t, want %t", got.Matched, tt.matched)
			}
			if !tt.matched {
				if got.Consumed != 0 || got.Node != nil {
					t.Errorf("rejected result = %#v, want zero value", got)
				}
				return
			}
			if got.Consumed != 1 {
				t.Errorf("Consumed = %d, want 1", got.Consumed)
			}
			if diff := cmp.Diff(tt.want, got.Node); diff != "" {
				t.Errorf("node mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestHeadingDefinitionBuildBlockUsesInlineCapability(t *testing.T) {
	wantErr := errors.New("inline failure")
	ctx := &testBuildContext{
		parseInlines: func(text string, origin ast.Position) ([]ast.Inline, error) {
			if text != "text" || origin != (ast.Position{Line: 3, Column: 4}) {
				t.Errorf("ParseInlines(%q, %#v), want text at line 3 column 4", text, origin)
			}
			return nil, wantErr
		},
	}

	got, err := (&headingDefinition{}).BuildBlock(ctx, &feature.TextBlock{
		Type:          typeHeading,
		Attr:          "level2",
		Text:          "text",
		ContentOrigin: ast.Position{Line: 3, Column: 4},
		Range:         ast.Range{Start: ast.Position{Line: 3, Column: 1}},
	})
	if got != nil {
		t.Errorf("BuildBlock returned a block: %#v", got)
	}
	if !errors.Is(err, wantErr) {
		t.Errorf("BuildBlock error = %v, want %v", err, wantErr)
	}
}

func TestHeadingDefinitionBuildBlockRejectsMalformedAttribute(t *testing.T) {
	tests := []string{"", "level", "other2", "level0", "level7", "levelx"}
	for _, attribute := range tests {
		t.Run(attribute, func(t *testing.T) {
			got, err := (&headingDefinition{}).BuildBlock(&testBuildContext{}, &feature.TextBlock{
				Type: typeHeading,
				Attr: attribute,
			})
			if got != nil {
				t.Errorf("BuildBlock returned a block: %#v", got)
			}
			if err == nil {
				t.Errorf("BuildBlock accepted malformed attribute %q", attribute)
			}
		})
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

func TestListDefinitionReadBlockBuildsPrivateRecursiveIR(t *testing.T) {
	input := testBlockInput{
		{Number: 2, Text: "* 親😀"},
		{Number: 3, Text: "**** child"},
		{Number: 4, Text: "# sibling"},
		{Number: 5, Text: "not a list"},
	}

	got, err := (&listDefinition{}).ReadBlock(input)
	if err != nil {
		t.Fatalf("ReadBlock returned an error: %v", err)
	}
	if !got.Matched || got.Consumed != 3 {
		t.Fatalf("result = %#v, want matched with 3 consumed lines", got)
	}

	list, ok := got.Node.(*listNode)
	if !ok {
		t.Fatalf("node type = %T, want *listNode", got.Node)
	}
	want := &listNode{
		ordered: false,
		items: []listItemNode{
			{
				blocks: []feature.BlockNode{
					&feature.TextBlock{
						Type:          feature.ParagraphBlockType,
						Text:          "親😀",
						ContentOrigin: ast.Position{Line: 2, Column: 3},
						Range: ast.Range{
							Start: ast.Position{Line: 2, Column: 3},
							End:   ast.Position{Line: 2, Column: 4},
						},
					},
					&listNode{
						ordered: false,
						items: []listItemNode{
							{
								blocks: []feature.BlockNode{
									&feature.TextBlock{
										Type:          feature.ParagraphBlockType,
										Text:          "child",
										ContentOrigin: ast.Position{Line: 3, Column: 6},
										Range: ast.Range{
											Start: ast.Position{Line: 3, Column: 6},
											End:   ast.Position{Line: 3, Column: 10},
										},
									},
								},
								rng: ast.Range{
									Start: ast.Position{Line: 3, Column: 1},
									End:   ast.Position{Line: 3, Column: 10},
								},
							},
						},
						rng: ast.Range{
							Start: ast.Position{Line: 3, Column: 1},
							End:   ast.Position{Line: 3, Column: 10},
						},
					},
				},
				rng: ast.Range{
					Start: ast.Position{Line: 2, Column: 1},
					End:   ast.Position{Line: 3, Column: 10},
				},
			},
			{
				blocks: []feature.BlockNode{
					&feature.TextBlock{
						Type:          feature.ParagraphBlockType,
						Text:          "sibling",
						ContentOrigin: ast.Position{Line: 4, Column: 3},
						Range: ast.Range{
							Start: ast.Position{Line: 4, Column: 3},
							End:   ast.Position{Line: 4, Column: 9},
						},
					},
				},
				rng: ast.Range{
					Start: ast.Position{Line: 4, Column: 1},
					End:   ast.Position{Line: 4, Column: 9},
				},
			},
		},
		rng: ast.Range{
			Start: ast.Position{Line: 2, Column: 1},
			End:   ast.Position{Line: 4, Column: 9},
		},
	}
	if diff := cmp.Diff(want, list, cmp.AllowUnexported(listNode{}, listItemNode{})); diff != "" {
		t.Errorf("list IR mismatch (-want +got):\n%s", diff)
	}
}

func TestListDefinitionRejectsWithoutConsumption(t *testing.T) {
	got, err := (&listDefinition{}).ReadBlock(testBlockInput{{Number: 1, Text: "plain"}})
	if err != nil {
		t.Fatalf("ReadBlock returned an error: %v", err)
	}
	if got != (feature.BlockReadResult{}) {
		t.Errorf("result = %#v, want zero value", got)
	}
}

func TestCodeBuilder(t *testing.T) {
	rng := ast.Range{Start: ast.Position{Line: 2, Column: 3}, End: ast.Position{Line: 2, Column: 6}}
	codeNode := &feature.TextBlock{Type: typeCode, Attr: "go", Text: "x := 1", Range: rng}
	gotCode, err := (&codeDefinition{}).BuildBlock(&testBuildContext{}, codeNode)
	if err != nil {
		t.Fatalf("code BuildBlock returned an error: %v", err)
	}
	wantCode := &ast.CodeBlock{Language: "go", Text: "x := 1", Range: rng}
	if diff := cmp.Diff(wantCode, gotCode); diff != "" {
		t.Errorf("code mismatch (-want +got):\n%s", diff)
	}
}

func TestCoreBlockDefinitions(t *testing.T) {
	definitions := Definitions()
	wantTypes := []string{"code", "heading", "list"}
	if len(definitions) != len(wantTypes) {
		t.Fatalf("definition count = %d, want %d", len(definitions), len(wantTypes))
	}
	for i, want := range wantTypes {
		if got := definitions[i].BlockType(); got != want {
			t.Errorf("definition %d type = %q, want %q", i, got, want)
		}
	}
	if _, ok := definitions[0].(feature.BlockReader); ok {
		t.Error("code definition unexpectedly implements BlockReader")
	}
	for _, index := range []int{1, 2} {
		if _, ok := definitions[index].(feature.BlockReader); !ok {
			t.Errorf("definition %d does not implement BlockReader", index)
		}
	}
}

func TestListDefinitionBuildBlockUsesRecursiveCapability(t *testing.T) {
	paragraphRange := ast.Range{Start: ast.Position{Line: 1, Column: 3}, End: ast.Position{Line: 1, Column: 6}}
	nestedRange := ast.Range{Start: ast.Position{Line: 2, Column: 1}, End: ast.Position{Line: 2, Column: 7}}
	nested := &listNode{
		items: []listItemNode{
			{
				blocks: []feature.BlockNode{
					&feature.TextBlock{Type: feature.ParagraphBlockType, Text: "child", Range: nestedRange},
				},
				rng: nestedRange,
			},
		},
		rng: nestedRange,
	}
	rootRange := ast.Range{Start: ast.Position{Line: 1, Column: 1}, End: nestedRange.End}
	root := &listNode{
		ordered: true,
		items: []listItemNode{
			{
				blocks: []feature.BlockNode{
					&feature.TextBlock{Type: feature.ParagraphBlockType, Text: "root", Range: paragraphRange},
					nested,
				},
				rng: rootRange,
			},
		},
		rng: rootRange,
	}

	var ctx *testBuildContext
	ctx = &testBuildContext{
		buildBlock: func(node feature.BlockNode) (ast.Block, error) {
			switch node := node.(type) {
			case *feature.TextBlock:
				return &ast.Paragraph{Range: node.Range}, nil
			case *listNode:
				return (&listDefinition{}).BuildBlock(ctx, node)
			default:
				return nil, errors.New("unexpected node")
			}
		},
	}

	got, err := (&listDefinition{}).BuildBlock(ctx, root)
	if err != nil {
		t.Fatalf("BuildBlock returned an error: %v", err)
	}
	want := &ast.List{
		Ordered: true,
		Items: []*ast.ListItem{
			{
				Blocks: []ast.Block{
					&ast.Paragraph{Range: paragraphRange},
					&ast.List{
						Items: []*ast.ListItem{
							{
								Blocks: []ast.Block{&ast.Paragraph{Range: nestedRange}},
								Range:  nestedRange,
							},
						},
						Range: nestedRange,
					},
				},
				Range: rootRange,
			},
		},
		Range: rootRange,
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("list mismatch (-want +got):\n%s", diff)
	}
	if root.BlockType() != typeList || root.BlockRange() != rootRange {
		t.Error("private list IR does not expose its declared type and range")
	}
}

func TestBlockBuildersRejectForeignIR(t *testing.T) {
	node := &foreignBlockNode{}
	ctx := &testBuildContext{}
	definitions := []feature.BlockDefinition{
		&codeDefinition{},
		&headingDefinition{},
		&listDefinition{},
	}
	for _, definition := range definitions {
		got, err := definition.BuildBlock(ctx, node)
		if got != nil {
			t.Errorf("%s builder returned a block: %#v", definition.BlockType(), got)
		}
		if err == nil {
			t.Errorf("%s builder returned no error", definition.BlockType())
		}
	}
}

type testBlockInput []feature.BlockLine

func (i testBlockInput) Len() int {
	return len(i)
}

func (i testBlockInput) Line(offset int) (feature.BlockLine, bool) {
	if offset < 0 || offset >= len(i) {
		return feature.BlockLine{}, false
	}
	return i[offset], true
}

type testBuildContext struct {
	parseInlines func(string, ast.Position) ([]ast.Inline, error)
	buildBlock   func(feature.BlockNode) (ast.Block, error)
}

type foreignBlockNode struct{}

func (*foreignBlockNode) BlockType() string {
	return "foreign"
}

func (*foreignBlockNode) BlockRange() ast.Range {
	return ast.Range{}
}

func (c *testBuildContext) ParseInlines(text string, origin ast.Position) ([]ast.Inline, error) {
	return c.parseInlines(text, origin)
}

func (c *testBuildContext) BuildBlock(node feature.BlockNode) (ast.Block, error) {
	return c.buildBlock(node)
}
