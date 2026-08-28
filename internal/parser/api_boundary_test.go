package parser_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"orische/internal/ast"
	"orische/internal/parser"
	"orische/internal/parser/feature"
)

func TestParserPackageBoundary_AcceptsExternalBlockDefinition(t *testing.T) {
	language := feature.Language{
		Paragraph: &boundaryParagraphDefinition{},
		Blocks: []feature.BlockDefinition{
			&boundarySugarDefinition{},
		},
		Inlines: []feature.InlineDirectiveDefinition{
			&boundaryInlineDefinition{},
		},
	}

	p, err := parser.NewParser(language)
	if err != nil {
		t.Fatalf("NewParser returned an error: %v", err)
	}

	got, err := p.Parse("! :[mark]{hello}")
	if err != nil {
		t.Fatalf("Parse returned an error: %v", err)
	}

	want := &ast.Document{
		Blocks: []ast.Block{
			&ast.Paragraph{
				Content: []ast.Inline{
					&ast.Emphasis{
						Content: []ast.Inline{
							&ast.Text{
								Value: "hello",
								Range: ast.Range{
									Start: ast.Position{Line: 1, Column: 11},
									End:   ast.Position{Line: 1, Column: 15},
								},
							},
						},
						Range: ast.Range{
							Start: ast.Position{Line: 1, Column: 3},
							End:   ast.Position{Line: 1, Column: 16},
						},
					},
				},
				Range: ast.Range{
					Start: ast.Position{Line: 1, Column: 1},
					End:   ast.Position{Line: 1, Column: 16},
				},
			},
		},
		Range: ast.Range{
			Start: ast.Position{Line: 1, Column: 1},
			End:   ast.Position{Line: 1, Column: 16},
		},
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Parse returned an unexpected document (-want +got):\n%s", diff)
	}
}

func TestParserPackageBoundary_DirectiveReceivesContentOrigin(t *testing.T) {
	language := feature.Language{
		Paragraph: &boundaryParagraphDefinition{},
		Blocks: []feature.BlockDefinition{
			&boundaryInlineBlockDefinition{},
		},
		Inlines: []feature.InlineDirectiveDefinition{
			&boundaryInlineDefinition{},
		},
	}
	p, err := parser.NewParser(language)
	if err != nil {
		t.Fatalf("NewParser returned an error: %v", err)
	}

	got, err := p.Parse(":::[note]\n:[mark]{hello}\n:::")
	if err != nil {
		t.Fatalf("Parse returned an error: %v", err)
	}
	want := &ast.Document{
		Blocks: []ast.Block{
			&ast.Paragraph{
				Content: []ast.Inline{
					&ast.Emphasis{
						Content: []ast.Inline{
							&ast.Text{
								Value: "hello",
								Range: ast.Range{
									Start: ast.Position{Line: 2, Column: 9},
									End:   ast.Position{Line: 2, Column: 13},
								},
							},
						},
						Range: ast.Range{
							Start: ast.Position{Line: 2, Column: 1},
							End:   ast.Position{Line: 2, Column: 14},
						},
					},
				},
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
}

func TestParserPackageBoundary_RejectsInvalidReaderResults(t *testing.T) {
	rng := ast.Range{
		Start: ast.Position{Line: 1, Column: 1},
		End:   ast.Position{Line: 1, Column: 4},
	}
	node := &boundaryBlockNode{typ: "probe", text: "text", rng: rng}

	tests := []struct {
		name   string
		result feature.BlockReadResult
		want   string
	}{
		{
			name:   "rejected result consumes input",
			result: feature.BlockReadResult{Consumed: 1},
			want:   "unmatched block reader consumed 1 lines",
		},
		{
			name:   "rejected result returns node",
			result: feature.BlockReadResult{Node: node},
			want:   "unmatched block reader returned a node",
		},
		{
			name:   "matched result consumes no input",
			result: feature.BlockReadResult{Matched: true, Node: node},
			want:   "matched block reader consumed 0 lines",
		},
		{
			name:   "matched result returns no node",
			result: feature.BlockReadResult{Matched: true, Consumed: 1},
			want:   "matched block reader returned no node",
		},
		{
			name:   "matched result consumes beyond input",
			result: feature.BlockReadResult{Matched: true, Consumed: 2, Node: node},
			want:   "matched block reader consumed 2 lines with only 1 available",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			language := feature.Language{
				Paragraph: &boundaryParagraphDefinition{},
				Blocks: []feature.BlockDefinition{
					&boundaryResultDefinition{result: tt.result},
				},
			}

			p, err := parser.NewParser(language)
			if err != nil {
				t.Fatalf("NewParser returned an error: %v", err)
			}

			got, err := p.Parse("text")
			if err == nil {
				t.Fatal("Parse returned no error")
			}
			if got != nil {
				t.Errorf("Parse returned a document: %#v", got)
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Errorf("Parse error = %q, want it to contain %q", err, tt.want)
			}
		})
	}
}

func TestParserPackageBoundary_AdvancesByConsumedLineCount(t *testing.T) {
	language := feature.Language{
		Paragraph: &boundaryParagraphDefinition{},
		Blocks: []feature.BlockDefinition{
			&boundaryMultilineDefinition{},
		},
	}
	p, err := parser.NewParser(language)
	if err != nil {
		t.Fatalf("NewParser returned an error: %v", err)
	}

	got, err := p.Parse("%%\nignored\nplain")
	if err != nil {
		t.Fatalf("Parse returned an error: %v", err)
	}
	if len(got.Blocks) != 2 {
		t.Fatalf("Block count = %d, want custom block followed by paragraph", len(got.Blocks))
	}
	first, ok := got.Blocks[0].(*ast.CodeBlock)
	if !ok || first.Text != "ignored" {
		t.Errorf("first Block = %#v, want custom CodeBlock", got.Blocks[0])
	}
	paragraph, ok := got.Blocks[1].(*ast.Paragraph)
	if !ok || len(paragraph.Content) != 1 {
		t.Fatalf("second Block = %#v, want external Paragraph", got.Blocks[1])
	}
	textNode, ok := paragraph.Content[0].(*ast.Text)
	if !ok || textNode.Value != "plain" || textNode.Range.Start.Line != 3 {
		t.Errorf("paragraph content = %#v, want plain text on line 3", paragraph.Content[0])
	}
}

func TestParserPackageBoundary_ReaderErrorAndTypeMismatchDoNotFallback(t *testing.T) {
	t.Run("reader error", func(t *testing.T) {
		wantErr := errors.New("reader failure")
		language := feature.Language{
			Paragraph: &boundaryParagraphDefinition{},
			Blocks: []feature.BlockDefinition{
				&boundaryErrorReaderDefinition{err: wantErr},
			},
		}
		p, err := parser.NewParser(language)
		if err != nil {
			t.Fatalf("NewParser returned an error: %v", err)
		}
		got, err := p.Parse("plain")
		if got != nil {
			t.Errorf("Parse returned a document: %#v", got)
		}
		if err != wantErr {
			t.Errorf("Parse error = %v, want original reader error %v", err, wantErr)
		}
	})

	t.Run("node type mismatch", func(t *testing.T) {
		language := feature.Language{
			Paragraph: &boundaryParagraphDefinition{},
			Blocks: []feature.BlockDefinition{
				&boundaryResultDefinition{
					result: feature.BlockReadResult{
						Matched:  true,
						Consumed: 1,
						Node:     &boundaryBlockNode{typ: "actual"},
					},
				},
			},
		}
		p, err := parser.NewParser(language)
		if err != nil {
			t.Fatalf("NewParser returned an error: %v", err)
		}
		got, err := p.Parse("plain")
		if got != nil {
			t.Errorf("Parse returned a document: %#v", got)
		}
		if err == nil || !strings.Contains(err.Error(), `declared block type "probe" but produced "actual"`) {
			t.Errorf("Parse error = %v, want normalized type mismatch", err)
		}
	})
}

func TestParserPackageBoundary_ValidatesLanguageBeforeParsing(t *testing.T) {
	tests := []struct {
		name     string
		language feature.Language
		want     string
	}{
		{
			name: "missing paragraph",
			want: "paragraph definition is required",
		},
		{
			name: "wrong paragraph type",
			language: feature.Language{
				Paragraph: &boundaryWrongParagraphDefinition{typ: "other"},
			},
			want: `paragraph definition declares block type "other"`,
		},
		{
			name: "paragraph also implements reader",
			language: feature.Language{
				Paragraph: &boundaryResultDefinition{typ: feature.ParagraphBlockType},
			},
			want: "paragraph definition must not implement BlockReader",
		},
		{
			name: "reserved paragraph in general definitions",
			language: feature.Language{
				Paragraph: &boundaryParagraphDefinition{},
				Blocks: []feature.BlockDefinition{
					&boundaryDirectiveDefinition{typ: "PARAGRAPH"},
				},
			},
			want: "paragraph is fixed parser infrastructure",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parser.NewParser(tt.language)
			if err == nil {
				t.Fatal("NewParser returned no error")
			}
			if got != nil {
				t.Errorf("NewParser returned a parser: %#v", got)
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Errorf("NewParser error = %q, want it to contain %q", err, tt.want)
			}
		})
	}
}

type boundaryBlockNode struct {
	typ  string
	text string
	rng  ast.Range
}

func (n *boundaryBlockNode) BlockType() string {
	return n.typ
}

func (n *boundaryBlockNode) BlockRange() ast.Range {
	return n.rng
}

type boundaryParagraphDefinition struct{}

func (*boundaryParagraphDefinition) BlockType() string {
	return feature.ParagraphBlockType
}

func (*boundaryParagraphDefinition) BuildParagraph(_ feature.BuildContext, node feature.BlockNode) (*ast.Paragraph, error) {
	block, ok := node.(*feature.TextBlock)
	if !ok {
		return nil, fmt.Errorf("expected *feature.TextBlock, got %T", node)
	}

	content := make([]ast.Inline, 0, 1)
	if block.Text != "" {
		content = append(content, &ast.Text{Value: block.Text, Range: block.Range})
	}
	return &ast.Paragraph{Content: content, Range: block.Range}, nil
}

type boundarySugarDefinition struct{}

func (*boundarySugarDefinition) BlockType() string {
	return "probe"
}

func (*boundarySugarDefinition) ReadBlock(input feature.BlockInput) (feature.BlockReadResult, error) {
	line, ok := input.Line(0)
	if !ok || !strings.HasPrefix(line.Text, "! ") {
		return feature.BlockReadResult{}, nil
	}

	return feature.BlockReadResult{
		Matched:  true,
		Consumed: 1,
		Node: &boundaryBlockNode{
			typ:  "probe",
			text: strings.TrimPrefix(line.Text, "! "),
			rng: ast.Range{
				Start: ast.Position{Line: line.Number, Column: 1},
				End:   ast.Position{Line: line.Number, Column: len([]rune(line.Text))},
			},
		},
	}, nil
}

func (*boundarySugarDefinition) BuildBlock(ctx feature.BuildContext, node feature.BlockNode) (ast.Block, error) {
	block, ok := node.(*boundaryBlockNode)
	if !ok {
		return nil, fmt.Errorf("expected *boundaryBlockNode, got %T", node)
	}

	content, err := ctx.ParseInlines(block.text, ast.Position{
		Line:   block.rng.Start.Line,
		Column: block.rng.Start.Column + 2,
	})
	if err != nil {
		return nil, err
	}
	return &ast.Paragraph{
		Content: content,
		Range:   block.rng,
	}, nil
}

type boundaryInlineDefinition struct{}

func (*boundaryInlineDefinition) InlineType() string {
	return "mark"
}

func (*boundaryInlineDefinition) ContentPolicy() feature.InlineContentPolicy {
	return feature.InlineContentNested
}

func (*boundaryInlineDefinition) ValidateAttribute(string) (bool, error) {
	return true, nil
}

func (*boundaryInlineDefinition) BuildInline(candidate feature.InlineDirectiveCandidate) (ast.Inline, error) {
	return &ast.Emphasis{Content: candidate.NestedContent, Range: candidate.Range}, nil
}

type boundaryInlineBlockDefinition struct{}

func (*boundaryInlineBlockDefinition) BlockType() string {
	return "note"
}

func (*boundaryInlineBlockDefinition) BuildBlock(ctx feature.BuildContext, node feature.BlockNode) (ast.Block, error) {
	block, ok := node.(*feature.TextBlock)
	if !ok {
		return nil, fmt.Errorf("expected *feature.TextBlock, got %T", node)
	}
	content, err := ctx.ParseInlines(block.Text, block.ContentOrigin)
	if err != nil {
		return nil, err
	}
	return &ast.Paragraph{Content: content, Range: block.Range}, nil
}

type boundaryResultDefinition struct {
	typ    string
	result feature.BlockReadResult
}

func (d *boundaryResultDefinition) BlockType() string {
	if d.typ != "" {
		return d.typ
	}
	return "probe"
}

func (d *boundaryResultDefinition) ReadBlock(feature.BlockInput) (feature.BlockReadResult, error) {
	return d.result, nil
}

func (*boundaryResultDefinition) BuildBlock(feature.BuildContext, feature.BlockNode) (ast.Block, error) {
	return nil, nil
}

func (*boundaryResultDefinition) BuildParagraph(feature.BuildContext, feature.BlockNode) (*ast.Paragraph, error) {
	return nil, nil
}

type boundaryDirectiveDefinition struct {
	typ string
}

type boundaryWrongParagraphDefinition struct {
	typ string
}

func (d *boundaryWrongParagraphDefinition) BlockType() string {
	return d.typ
}

func (*boundaryWrongParagraphDefinition) BuildParagraph(feature.BuildContext, feature.BlockNode) (*ast.Paragraph, error) {
	return nil, nil
}

type boundaryMultilineDefinition struct{}

func (*boundaryMultilineDefinition) BlockType() string {
	return "multiline"
}

func (*boundaryMultilineDefinition) ReadBlock(input feature.BlockInput) (feature.BlockReadResult, error) {
	first, ok := input.Line(0)
	if !ok || first.Text != "%%" || input.Len() < 2 {
		return feature.BlockReadResult{}, nil
	}
	second, _ := input.Line(1)
	return feature.BlockReadResult{
		Matched:  true,
		Consumed: 2,
		Node: &boundaryBlockNode{
			typ:  "multiline",
			text: second.Text,
			rng: ast.Range{
				Start: ast.Position{Line: first.Number, Column: 1},
				End:   ast.Position{Line: second.Number, Column: len([]rune(second.Text))},
			},
		},
	}, nil
}

func (*boundaryMultilineDefinition) BuildBlock(_ feature.BuildContext, node feature.BlockNode) (ast.Block, error) {
	block, ok := node.(*boundaryBlockNode)
	if !ok {
		return nil, fmt.Errorf("expected *boundaryBlockNode, got %T", node)
	}
	return &ast.CodeBlock{Text: block.text, Range: block.rng}, nil
}

type boundaryErrorReaderDefinition struct {
	err error
}

func (*boundaryErrorReaderDefinition) BlockType() string {
	return "error"
}

func (d *boundaryErrorReaderDefinition) ReadBlock(feature.BlockInput) (feature.BlockReadResult, error) {
	return feature.BlockReadResult{}, d.err
}

func (*boundaryErrorReaderDefinition) BuildBlock(feature.BuildContext, feature.BlockNode) (ast.Block, error) {
	return nil, nil
}

func (d *boundaryDirectiveDefinition) BlockType() string {
	return d.typ
}

func (*boundaryDirectiveDefinition) BuildBlock(feature.BuildContext, feature.BlockNode) (ast.Block, error) {
	return nil, nil
}

var (
	_ feature.BlockNode                 = (*boundaryBlockNode)(nil)
	_ feature.ParagraphDefinition       = (*boundaryParagraphDefinition)(nil)
	_ feature.BlockSugarDefinition      = (*boundarySugarDefinition)(nil)
	_ feature.BlockDefinition           = (*boundaryInlineBlockDefinition)(nil)
	_ feature.InlineDirectiveDefinition = (*boundaryInlineDefinition)(nil)
)
