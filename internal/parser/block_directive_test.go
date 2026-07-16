package parser

import (
	"reflect"
	"testing"

	"orische/internal/ast"
)

func TestParseBlockDirective(t *testing.T) {
	input := &blockContext{
		lines: []string{
			":::[code:go]",
			"fmt.Println(\"Hello, world!\")",
			":::",
		},
		pos: 0,
	}

	output, ok, err := (&blockDirectiveParser{}).parse(input)

	want := &parsedBlock{
		Type: "code",
		Attr: "go",
		Text: "fmt.Println(\"Hello, world!\")",
		Range: ast.Range{
			StartLine: 1,
			EndLine:   3,
		},
	}

	ctx_pos_want := 2

	if !ok || err != nil {
		t.Errorf("parse failed.")
	}
	if !reflect.DeepEqual(output, want) {
		t.Errorf("parsed incorrectly.\nwant:\n%v\ngot:\n%v", want, output)
	}
	if input.pos != ctx_pos_want {
		t.Errorf("position in context is not updated correctly. want: %d, got: %d", ctx_pos_want, input.pos)
	}
}

func TestParseBlockDirective_NoAttr(t *testing.T) {
	input := &blockContext{
		lines: []string{
			":::[code]",
			"fmt.Println(\"Hello, world!\")",
			":::",
		},
		pos: 0,
	}

	output, ok, err := (&blockDirectiveParser{}).parse(input)

	want := &parsedBlock{
		Type: "code",
		Attr: "",
		Text: "fmt.Println(\"Hello, world!\")",
		Range: ast.Range{
			StartLine: 1,
			EndLine:   3,
		},
	}

	if !ok || err != nil {
		t.Errorf("parse failed.")
	}
	if !reflect.DeepEqual(output, want) {
		t.Errorf("parsed incorrectly.\nwant:\n%v\ngot:\n%v", want, output)
	}
}

func TestParseBlockDirective_NoClosing(t *testing.T) {
	input := &blockContext{
		lines: []string{
			":::[code:go]",
			"fmt.Println(\"Hello, world!\")",
		},
		pos: 0,
	}

	output, ok, err := (&blockDirectiveParser{}).parse(input)

	if err != nil {
		t.Fatalf("parse returned an error: %v", err)
	}
	if ok {
		t.Error("unterminated directive was parsed successfully")
	}
	if output != nil {
		t.Errorf("unterminated directive returned a node: %v", output)
	}
	if input.pos != 0 {
		t.Errorf("position in context was not restored. want: 0, got: %d", input.pos)
	}

	doc, err := NewParser(nil).parseDocument(input.lines)
	if err != nil {
		t.Fatalf("parse document returned an error: %v", err)
	}

	wantDoc := &parsedDocument{
		Blocks: []parsedBlockNode{
			&parsedBlock{
				Type: "Paragraph",
				Attr: "",
				Text: ":::[code:go]\nfmt.Println(\"Hello, world!\")",
				Range: ast.Range{
					StartLine: 1,
					EndLine:   2,
				},
			},
		},
		Range: ast.Range{
			StartLine: 1,
			EndLine:   2,
		},
	}
	if !reflect.DeepEqual(doc, wantDoc) {
		t.Errorf("document parsed incorrectly.\nwant:\n%v\ngot:\n%v", wantDoc, doc)
	}
}

func TestParseBlockDirective_NoType(t *testing.T) {
	tests := []string{
		":::[:go]",
		":::[]",
	}

	for _, opener := range tests {
		t.Run(opener, func(t *testing.T) {
			input := &blockContext{
				lines: []string{
					opener,
					"fmt.Println(\"Hello, world!\")",
					":::",
				},
				pos: 0,
			}

			output, ok, err := (&blockDirectiveParser{}).parse(input)
			if err != nil {
				t.Fatalf("parse returned an error: %v", err)
			}
			if ok {
				t.Error("directive without a type was parsed successfully")
			}
			if output != nil {
				t.Errorf("directive without a type returned a node: %v", output)
			}
			if input.pos != 0 {
				t.Errorf("position in context changed. want 0, got %d", input.pos)
			}
		})
	}
}
