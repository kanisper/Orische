package parser

import (
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"orische/internal/ast"
	"orische/internal/parser/feature"
	"orische/internal/parser/syntax"
)

func TestNewParserValidatesDefinitionsBeforeMutation(t *testing.T) {
	var nilBlock *testDirectiveDefinition
	var nilInline *testInlineDirectiveDefinition

	tests := []struct {
		name   string
		mutate func(*feature.Language)
		want   string
	}{
		{
			name: "nil block",
			mutate: func(language *feature.Language) {
				language.Blocks = append(language.Blocks, nilBlock)
			},
			want: "block definition is nil",
		},
		{
			name: "empty block type",
			mutate: func(language *feature.Language) {
				language.Blocks = append(language.Blocks, &testDirectiveDefinition{})
			},
			want: "block type must not be empty",
		},
		{
			name: "normalized block duplicate",
			mutate: func(language *feature.Language) {
				language.Blocks = append(language.Blocks, &testDirectiveDefinition{typ: "CODE"})
			},
			want: `block definition "code" is already registered`,
		},
		{
			name: "nil inline",
			mutate: func(language *feature.Language) {
				language.Inlines = append(language.Inlines, nilInline)
			},
			want: "inline definition is nil",
		},
		{
			name: "empty inline type",
			mutate: func(language *feature.Language) {
				language.Inlines = append(language.Inlines, &testInlineDirectiveDefinition{})
			},
			want: "inline definition type must not be empty",
		},
		{
			name: "invalid inline policy",
			mutate: func(language *feature.Language) {
				language.Inlines = append(language.Inlines, &testInlineDirectiveDefinition{typ: "invalid", policy: 99})
			},
			want: `inline directive definition "invalid" has invalid content policy 99`,
		},
		{
			name: "normalized inline duplicate",
			mutate: func(language *feature.Language) {
				language.Inlines = append(language.Inlines, &testInlineDirectiveDefinition{typ: "EM", policy: feature.InlineContentNested})
			},
			want: `inline directive definition "em" is already registered`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			language := syntax.Core()
			tt.mutate(&language)
			got, err := NewParser(language)
			if err == nil {
				t.Fatal("NewParser returned no error")
			}
			if got != nil {
				t.Errorf("NewParser returned a parser: %#v", got)
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Errorf("error = %q, want it to contain %q", err, tt.want)
			}
		})
	}
}

func TestNewParserCopiesLanguageSlices(t *testing.T) {
	language := syntax.Core()
	p, err := NewParser(language)
	if err != nil {
		t.Fatalf("NewParser returned an error: %v", err)
	}

	language.Blocks[1] = &testDirectiveDefinition{typ: "replacement"}
	language.Inlines[0] = &testInlineDirectiveDefinition{
		typ:    "replacement",
		policy: feature.InlineContentLiteral,
	}

	got, err := p.Parse("= :[em]{text}")
	if err != nil {
		t.Fatalf("Parse returned an error after source Language mutation: %v", err)
	}
	want := &ast.Document{
		Blocks: []ast.Block{
			&ast.Heading{
				Level: 1,
				Content: []ast.Inline{
					&ast.Emphasis{
						Content: []ast.Inline{
							&ast.Text{Value: "text", Range: ast.Range{Start: ast.Position{Line: 1, Column: 9}, End: ast.Position{Line: 1, Column: 12}}},
						},
						Range: ast.Range{Start: ast.Position{Line: 1, Column: 3}, End: ast.Position{Line: 1, Column: 13}},
					},
				},
				Range: ast.Range{Start: ast.Position{Line: 1, Column: 1}, End: ast.Position{Line: 1, Column: 13}},
			},
		},
		Range: ast.Range{Start: ast.Position{Line: 1, Column: 1}, End: ast.Position{Line: 1, Column: 13}},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("document mismatch (-want +got):\n%s", diff)
	}
}

func TestBlockDirectiveDispatchNormalizesOnlyType(t *testing.T) {
	language := syntax.Core()
	language.Blocks = append(language.Blocks,
		&testDirectiveDefinition{typ: "First"},
		&testDirectiveDefinition{typ: "Second"},
	)
	p, err := NewParser(language)
	if err != nil {
		t.Fatalf("NewParser returned an error: %v", err)
	}

	got, err := p.Parse(":::[fIRSt:Ä:TTR]\n内容😀\n:::\n\n:::[SECOND:x]\nText\n:::")
	if err != nil {
		t.Fatalf("Parse returned an error: %v", err)
	}
	want := &ast.Document{
		Blocks: []ast.Block{
			&ast.CodeBlock{Language: "Ä:TTR", Text: "内容😀", Range: ast.Range{Start: ast.Position{Line: 1, Column: 1}, End: ast.Position{Line: 3, Column: 3}}},
			&ast.CodeBlock{Language: "x", Text: "Text", Range: ast.Range{Start: ast.Position{Line: 5, Column: 1}, End: ast.Position{Line: 7, Column: 3}}},
		},
		Range: ast.Range{Start: ast.Position{Line: 1, Column: 1}, End: ast.Position{Line: 7, Column: 3}},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("document mismatch (-want +got):\n%s", diff)
	}
}

func TestMinimalLanguageUsesFixedFrontendReaders(t *testing.T) {
	core := syntax.Core()
	p, err := NewParser(feature.Language{Paragraph: core.Paragraph})
	if err != nil {
		t.Fatalf("NewParser returned an error: %v", err)
	}

	doc, err := p.Parse("plain")
	if err != nil {
		t.Fatalf("Parse returned an error: %v", err)
	}
	if len(doc.Blocks) != 1 {
		t.Fatalf("Block count = %d, want one Paragraph", len(doc.Blocks))
	}
	paragraph, ok := doc.Blocks[0].(*ast.Paragraph)
	if !ok || len(paragraph.Content) != 1 {
		t.Fatalf("Block = %#v, want one-content Paragraph", doc.Blocks[0])
	}
	text, ok := paragraph.Content[0].(*ast.Text)
	if !ok || text.Value != "plain" {
		t.Errorf("Paragraph content = %#v, want plain text", paragraph.Content[0])
	}
}

func TestDefinitionNamespacesUseUnicodeCaseNormalization(t *testing.T) {
	core := syntax.Core()
	inline := &testInlineDirectiveDefinition{
		typ:    "äbc",
		policy: feature.InlineContentLiteral,
	}
	language := feature.Language{
		Paragraph: core.Paragraph,
		Blocks: []feature.BlockDefinition{
			&testDirectiveDefinition{typ: "ÄBC"},
		},
		Inlines: []feature.InlineDirectiveDefinition{inline},
	}
	p, err := NewParser(language)
	if err != nil {
		t.Fatalf("cross-category definitions returned an error: %v", err)
	}
	if _, err := p.Parse(":::[äBc]\ntext\n:::\n\n:[ÄbC]{inline}"); err != nil {
		t.Fatalf("cross-category definitions failed to parse: %v", err)
	}

	language.Blocks = append(language.Blocks, &testDirectiveDefinition{typ: "äbc"})
	got, err := NewParser(language)
	if err == nil {
		t.Fatal("Unicode case-only Block duplicate returned no error")
	}
	if got != nil {
		t.Errorf("NewParser returned a parser: %#v", got)
	}
	if !strings.Contains(err.Error(), `block definition "äbc" is already registered`) {
		t.Errorf("duplicate error = %q, want normalized Unicode type", err)
	}
}

type testDirectiveDefinition struct {
	typ string
}

func (d *testDirectiveDefinition) BlockType() string {
	return d.typ
}

func (*testDirectiveDefinition) BuildBlock(_ feature.BuildContext, node feature.BlockNode) (ast.Block, error) {
	block, ok := node.(*feature.TextBlock)
	if !ok {
		return nil, fmt.Errorf("expected *feature.TextBlock, got %T", node)
	}
	return &ast.CodeBlock{Language: block.Attr, Text: block.Text, Range: block.Range}, nil
}

type testInlineDirectiveDefinition struct {
	typ      string
	policy   feature.InlineContentPolicy
	validate func(string) (bool, error)
	build    func(feature.InlineDirectiveCandidate) (ast.Inline, error)
}

func (d *testInlineDirectiveDefinition) InlineType() string {
	return d.typ
}

func (d *testInlineDirectiveDefinition) ContentPolicy() feature.InlineContentPolicy {
	return d.policy
}

func (d *testInlineDirectiveDefinition) ValidateAttribute(attribute string) (bool, error) {
	if d.validate == nil {
		return true, nil
	}
	return d.validate(attribute)
}

func (d *testInlineDirectiveDefinition) BuildInline(candidate feature.InlineDirectiveCandidate) (ast.Inline, error) {
	if d.build == nil {
		return &ast.CodeSpan{Value: candidate.LiteralContent, Range: candidate.Range}, nil
	}
	return d.build(candidate)
}
