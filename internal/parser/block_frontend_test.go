package parser

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"orische/internal/ast"
	"orische/internal/parser/feature"
)

func TestBlockDirectiveReader(t *testing.T) {
	tests := []struct {
		name  string
		lines []string
		want  feature.BlockReadResult
	}{
		{
			name:  "with attribute and unicode content",
			lines: []string{":::[code:go]", "fmt.Println(\"日😀\")", ":::"},
			want: feature.BlockReadResult{
				Matched:  true,
				Consumed: 3,
				Node: &feature.TextBlock{
					Type:          "code",
					Attr:          "go",
					Text:          "fmt.Println(\"日😀\")",
					ContentOrigin: ast.Position{Line: 3, Column: 1},
					Range: ast.Range{
						Start: ast.Position{Line: 2, Column: 1},
						End:   ast.Position{Line: 4, Column: 3},
					},
				},
			},
		},
		{
			name:  "without attribute",
			lines: []string{":::[code]", "text", ":::"},
			want: feature.BlockReadResult{
				Matched:  true,
				Consumed: 3,
				Node: &feature.TextBlock{
					Type:          "code",
					Text:          "text",
					ContentOrigin: ast.Position{Line: 3, Column: 1},
					Range: ast.Range{
						Start: ast.Position{Line: 2, Column: 1},
						End:   ast.Position{Line: 4, Column: 3},
					},
				},
			},
		},
		{name: "unterminated", lines: []string{":::[code]", "text"}},
		{name: "empty type", lines: []string{":::[]", "text", ":::"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := (&blockDirectiveReader{}).ReadBlock(&blockInput{lines: append([]string{"before"}, tt.lines...), start: 1})
			if err != nil {
				t.Fatalf("ReadBlock returned an error: %v", err)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("result mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestParagraphReader(t *testing.T) {
	input := &blockInput{
		lines: []string{"before", "日😀", "second", "", "after"},
		start: 1,
	}
	got, err := (&paragraphReader{}).ReadBlock(input)
	if err != nil {
		t.Fatalf("ReadBlock returned an error: %v", err)
	}
	want := feature.BlockReadResult{
		Matched:  true,
		Consumed: 2,
		Node: &feature.TextBlock{
			Type:          feature.ParagraphBlockType,
			Text:          "日😀\nsecond",
			ContentOrigin: ast.Position{Line: 2, Column: 1},
			Range: ast.Range{
				Start: ast.Position{Line: 2, Column: 1},
				End:   ast.Position{Line: 3, Column: 6},
			},
		},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("result mismatch (-want +got):\n%s", diff)
	}
}

func TestParagraphDefinitionBuildsWithActiveInlineCapability(t *testing.T) {
	node := &feature.TextBlock{
		Type:          feature.ParagraphBlockType,
		Text:          "text",
		ContentOrigin: ast.Position{Line: 2, Column: 3},
		Range:         ast.Range{Start: ast.Position{Line: 2, Column: 3}, End: ast.Position{Line: 2, Column: 6}},
	}
	wantContent := []ast.Inline{&ast.Text{Value: node.Text, Range: node.Range}}
	ctx := &stubBuildContext{
		parseInlines: func(text string, origin ast.Position) ([]ast.Inline, error) {
			if text != node.Text || origin != node.ContentOrigin {
				t.Errorf("ParseInlines(%q, %#v), want paragraph text and content origin", text, origin)
			}
			return wantContent, nil
		},
	}

	got, err := (&paragraphDefinition{}).BuildBlock(ctx, node)
	if err != nil {
		t.Fatalf("BuildBlock returned an error: %v", err)
	}
	want := &ast.Paragraph{Content: wantContent, Range: node.Range}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("paragraph mismatch (-want +got):\n%s", diff)
	}
}

func TestCoreReaderPrecedence(t *testing.T) {
	p := mustCoreParser(t)
	readers := p.spec.getReaders()
	if len(readers) != 4 {
		t.Fatalf("reader count = %d, want 4", len(readers))
	}
	if _, ok := readers[0].(*blockDirectiveReader); !ok {
		t.Errorf("reader 0 type = %T, want *blockDirectiveReader", readers[0])
	}
	if got := readers[1].(feature.BlockSugarDefinition).BlockType(); got != "heading" {
		t.Errorf("reader 1 type = %q, want heading", got)
	}
	if got := readers[2].(feature.BlockSugarDefinition).BlockType(); got != "list" {
		t.Errorf("reader 2 type = %q, want list", got)
	}
	if _, ok := readers[3].(*paragraphReader); !ok {
		t.Errorf("reader 3 type = %T, want *paragraphReader", readers[3])
	}
}

type stubBuildContext struct {
	parseInlines func(string, ast.Position) ([]ast.Inline, error)
}

func (c *stubBuildContext) ParseInlines(text string, origin ast.Position) ([]ast.Inline, error) {
	return c.parseInlines(text, origin)
}

func (*stubBuildContext) BuildBlock(feature.BlockNode) (ast.Block, error) {
	panic("unexpected BuildBlock call")
}
