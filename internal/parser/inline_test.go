package parser

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"orische/internal/ast"
)

func TestParseInline(t *testing.T) {
	input := "plain text 1 :[em]{emphasized text} :[code]{codespan text}:[link:http://example.com]{link text}"

	want := []ast.Inline{
		&ast.Text{
			Value: "plain text 1 ",
		},
		&ast.Emphasis{
			Content: []ast.Inline{
				&ast.Text{Value: "emphasized text"},
			},
		},
		&ast.Text{
			Value: " ",
		},
		&ast.CodeSpan{
			Value: "codespan text",
		},
		&ast.Link{
			URI: "http://example.com",
			Content: []ast.Inline{
				&ast.Text{Value: "link text"},
			},
		},
	}

	output, err := parseInlines(input)

	if err != nil {
		t.Errorf("parse failed.")
	}
	if diff := cmp.Diff(want, output); diff != "" {
		t.Errorf("parsed incorrectly.\n(-want +got)\n%s\n", diff)
	}
}

func TestParseNestedInlines(t *testing.T) {
	input := ":[em]{emphasized text1 :[link:http://example.com]{linktext} emphasize text2}"

	want := []ast.Inline{
		&ast.Emphasis{
			Content: []ast.Inline{
				&ast.Text{Value: "emphasized text1 "},
				&ast.Link{
					URI: "http://example.com",
					Content: []ast.Inline{
						&ast.Text{Value: "linktext"},
					},
				},
				&ast.Text{Value: " emphasize text2"},
			},
		},
	}

	output, err := parseInlines(input)

	if err != nil {
		t.Errorf("parse failed.")
	}
	if diff := cmp.Diff(want, output); diff != "" {
		t.Errorf("parsed incorrectly.\n(-want +got)\n%s\n", diff)
	}
}

func TestParseNestedCodespan(t *testing.T) {
	input := ":[code]{codespan :[em]{this directive should be escaped}}"
	want := []ast.Inline{
		&ast.CodeSpan{
			Value: "codespan :[em]{this directive should be escaped",
		},
		&ast.Text{Value: "}"},
	}

	output, err := parseInlines(input)

	if err != nil {
		t.Errorf("parse failed.")
	}
	if diff := cmp.Diff(want, output); diff != "" {
		t.Errorf("parsed incorrectly.\n(-want +got)\n%s\n", diff)
	}
}

func TestParseInvalidInlines(t *testing.T) {
	input := ":[em]{There is no end brace."

	want := []ast.Inline{
		&ast.Text{Value: ":[em]{There is no end brace."},
	}

	output, err := parseInlines(input)

	if err != nil {
		t.Errorf("parse failed.")
	}
	if diff := cmp.Diff(want, output); diff != "" {
		t.Errorf("parsed incorrectly.\n(-want +got)\n%s\n", diff)
	}
}
