package parser

import (
	"testing"

	"medoc/internal/ast"

	"github.com/google/go-cmp/cmp"
)

func TestParseDocument(t *testing.T) {
	input := []string{
		"= Heading1",
		"",
		"paragraph1 line1",
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
					StartLine: 0,
					EndLine:   0,
				},
			},
			&parsedBlock{
				Type: "Paragraph",
				Attr: "",
				Text: "paragraph1 line1",
				Range: ast.Range{
					StartLine: 2,
					EndLine:   3,
				},
			},
			&parsedBlock{
				Type: "code",
				Attr: "go",
				Text: "fmt.Println(\"Hello\")",
				Range: ast.Range{
					StartLine: 4,
					EndLine:   6,
				},
			},
			&parsedBlock{
				Type: "Heading",
				Attr: "level2",
				Text: "Heading2",
				Range: ast.Range{
					StartLine: 8,
					EndLine:   8,
				},
			},
			&parsedBlock{
				Type: "Paragraph",
				Attr: "",
				Text: "paragraph2 line1\nparagraph2 line2",
				Range: ast.Range{
					StartLine: 10,
					EndLine:   12,
				},
			},
		},
		Range: ast.Range{
			StartLine: 0,
			EndLine:   12,
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
