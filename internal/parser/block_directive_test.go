package parser

import (
	"strings"
	"testing"

	"orische/internal/ast"
)

func TestHeadingBlockDirective(t *testing.T) {
	doc, err := Parse(":::[HEADING:level2]\nTitle :[em]{日}\n:::")
	if err != nil {
		t.Fatalf("Parse returned an error: %v", err)
	}
	if len(doc.Blocks) != 1 {
		t.Fatalf("Block count = %d, want 1", len(doc.Blocks))
	}

	heading, ok := doc.Blocks[0].(*ast.Heading)
	if !ok {
		t.Fatalf("Block type = %T, want *ast.Heading", doc.Blocks[0])
	}
	if heading.Level != 2 {
		t.Errorf("Heading level = %d, want 2", heading.Level)
	}
	wantRange := ast.Range{
		Start: ast.Position{Line: 1, Column: 1},
		End:   ast.Position{Line: 3, Column: 3},
	}
	if heading.Range != wantRange {
		t.Errorf("Heading range = %#v, want %#v", heading.Range, wantRange)
	}
	if len(heading.Content) != 2 {
		t.Fatalf("Heading inline count = %d, want text and emphasis", len(heading.Content))
	}
	if emphasis, ok := heading.Content[1].(*ast.Emphasis); !ok {
		t.Errorf("Heading inline type = %T, want *ast.Emphasis", heading.Content[1])
	} else if emphasis.Range != (ast.Range{
		Start: ast.Position{Line: 2, Column: 7},
		End:   ast.Position{Line: 2, Column: 14},
	}) {
		t.Errorf("Emphasis range = %#v, want line 2 columns 7-14", emphasis.Range)
	}
}

func TestParagraphBlockDirective(t *testing.T) {
	doc, err := Parse(":::[PARAGRAPH:ignored]\nText :[em]{日}\n:::")
	if err != nil {
		t.Fatalf("Parse returned an error: %v", err)
	}
	if len(doc.Blocks) != 1 {
		t.Fatalf("Block count = %d, want 1", len(doc.Blocks))
	}

	paragraph, ok := doc.Blocks[0].(*ast.Paragraph)
	if !ok {
		t.Fatalf("Block type = %T, want *ast.Paragraph", doc.Blocks[0])
	}
	wantRange := ast.Range{
		Start: ast.Position{Line: 1, Column: 1},
		End:   ast.Position{Line: 3, Column: 3},
	}
	if paragraph.Range != wantRange {
		t.Errorf("Paragraph range = %#v, want %#v", paragraph.Range, wantRange)
	}
	if len(paragraph.Content) != 2 {
		t.Fatalf("Paragraph inline count = %d, want text and emphasis", len(paragraph.Content))
	}
	if _, ok := paragraph.Content[1].(*ast.Emphasis); !ok {
		t.Errorf("Paragraph inline type = %T, want *ast.Emphasis", paragraph.Content[1])
	}
}

func TestHeadingBlockDirectiveRejectsInvalidLevel(t *testing.T) {
	for _, attribute := range []string{"", "level", "other2", "level0", "level7", "levelx"} {
		t.Run(attribute, func(t *testing.T) {
			input := ":::[heading:" + attribute + "]\ntext\n:::"
			if attribute == "" {
				input = ":::[heading]\ntext\n:::"
			}

			doc, err := Parse(input)
			if doc != nil {
				t.Errorf("Parse returned a document: %#v", doc)
			}
			if err == nil || !strings.Contains(err.Error(), "invalid heading attribute") {
				t.Errorf("Parse error = %v, want invalid heading attribute error", err)
			}
		})
	}
}
