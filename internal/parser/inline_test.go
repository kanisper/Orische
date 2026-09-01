package parser

import (
	"testing"
	"unicode/utf8"

	"github.com/google/go-cmp/cmp"

	"orische/internal/ast"
)

func TestParseInline(t *testing.T) {
	input := "plain text 1 :[em]{emphasized text} :[code]{codespan text}:[link:http://example.com]{link text}"

	want := []ast.Inline{
		&ast.Text{
			Value: "plain text 1 ",
			Range: ast.Range{
				Start: ast.Position{Line: 1, Column: 1},
				End:   ast.Position{Line: 1, Column: 13},
			},
		},
		&ast.Emphasis{
			Content: []ast.Inline{
				&ast.Text{
					Value: "emphasized text",
					Range: ast.Range{
						Start: ast.Position{Line: 1, Column: 20},
						End:   ast.Position{Line: 1, Column: 34},
					},
				},
			},
			Range: ast.Range{
				Start: ast.Position{Line: 1, Column: 14},
				End:   ast.Position{Line: 1, Column: 35},
			},
		},
		&ast.Text{
			Value: " ",
			Range: ast.Range{
				Start: ast.Position{Line: 1, Column: 36},
				End:   ast.Position{Line: 1, Column: 36},
			},
		},
		&ast.CodeSpan{
			Value: "codespan text",
			Range: ast.Range{
				Start: ast.Position{Line: 1, Column: 37},
				End:   ast.Position{Line: 1, Column: 58},
			},
		},
		&ast.Link{
			URI: "http://example.com",
			Content: []ast.Inline{
				&ast.Text{
					Value: "link text",
					Range: ast.Range{
						Start: ast.Position{Line: 1, Column: 86},
						End:   ast.Position{Line: 1, Column: 94},
					},
				},
			},
			Range: ast.Range{
				Start: ast.Position{Line: 1, Column: 59},
				End:   ast.Position{Line: 1, Column: 95},
			},
		},
	}

	output, err := mustCoreParser(t).parseInlines(input, ast.Position{Line: 1, Column: 1})

	if err != nil {
		t.Fatalf("parse returned an error: %v", err)
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
				&ast.Text{
					Value: "emphasized text1 ",
					Range: ast.Range{
						Start: ast.Position{Line: 1, Column: 7},
						End:   ast.Position{Line: 1, Column: 23},
					},
				},
				&ast.Link{
					URI: "http://example.com",
					Content: []ast.Inline{
						&ast.Text{
							Value: "linktext",
							Range: ast.Range{
								Start: ast.Position{Line: 1, Column: 51},
								End:   ast.Position{Line: 1, Column: 58},
							},
						},
					},
					Range: ast.Range{
						Start: ast.Position{Line: 1, Column: 24},
						End:   ast.Position{Line: 1, Column: 59},
					},
				},
				&ast.Text{
					Value: " emphasize text2",
					Range: ast.Range{
						Start: ast.Position{Line: 1, Column: 60},
						End:   ast.Position{Line: 1, Column: 75},
					},
				},
			},
			Range: ast.Range{
				Start: ast.Position{Line: 1, Column: 1},
				End:   ast.Position{Line: 1, Column: 76},
			},
		},
	}

	output, err := mustCoreParser(t).parseInlines(input, ast.Position{Line: 1, Column: 1})

	if err != nil {
		t.Fatalf("parse returned an error: %v", err)
	}
	if diff := cmp.Diff(want, output); diff != "" {
		t.Errorf("parsed incorrectly.\n(-want +got)\n%s\n", diff)
	}
}

func TestInlineTypesAreCaseInsensitiveWithoutNormalizingValues(t *testing.T) {
	origin := ast.Position{Line: 2, Column: 3}
	lower := ":[em:Ä]{日 :[link:/X:Y]{界}} :[code:Go]{値}"
	mixed := ":[Em:Ä]{日 :[LiNk:/X:Y]{界}} :[CoDe:Go]{値}"

	want, err := mustCoreParser(t).parseInlines(lower, origin)
	if err != nil {
		t.Fatalf("lowercase parse returned an error: %v", err)
	}
	got, err := mustCoreParser(t).parseInlines(mixed, origin)
	if err != nil {
		t.Fatalf("mixed-case parse returned an error: %v", err)
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("mixed-case directives changed values or ranges (-lower +mixed):\n%s", diff)
	}

	link := got[0].(*ast.Emphasis).Content[1].(*ast.Link)
	if link.URI != "/X:Y" {
		t.Errorf("Link URI = %q, want original /X:Y", link.URI)
	}
	code := got[2].(*ast.CodeSpan)
	if code.Value != "値" {
		t.Errorf("Code value = %q, want original content", code.Value)
	}
}

func TestParseNestedCodespan(t *testing.T) {
	input := ":[code]{codespan :[em]{this directive should be escaped}}"
	want := []ast.Inline{
		&ast.CodeSpan{
			Value: "codespan :[em]{this directive should be escaped",
			Range: ast.Range{
				Start: ast.Position{Line: 1, Column: 1},
				End:   ast.Position{Line: 1, Column: 56},
			},
		},
		&ast.Text{
			Value: "}",
			Range: ast.Range{
				Start: ast.Position{Line: 1, Column: 57},
				End:   ast.Position{Line: 1, Column: 57},
			},
		},
	}

	output, err := mustCoreParser(t).parseInlines(input, ast.Position{Line: 1, Column: 1})

	if err != nil {
		t.Fatalf("parse returned an error: %v", err)
	}
	if diff := cmp.Diff(want, output); diff != "" {
		t.Errorf("parsed incorrectly.\n(-want +got)\n%s\n", diff)
	}
}

func TestParseEmptyInlineContent(t *testing.T) {
	input := ":[em]{}:[code]{}:[link:https://example.com]{}"
	want := []ast.Inline{
		&ast.Emphasis{
			Content: nil,
			Range: ast.Range{
				Start: ast.Position{Line: 1, Column: 1},
				End:   ast.Position{Line: 1, Column: 7},
			},
		},
		&ast.CodeSpan{
			Value: "",
			Range: ast.Range{
				Start: ast.Position{Line: 1, Column: 8},
				End:   ast.Position{Line: 1, Column: 16},
			},
		},
		&ast.Link{
			URI:     "https://example.com",
			Content: nil,
			Range: ast.Range{
				Start: ast.Position{Line: 1, Column: 17},
				End:   ast.Position{Line: 1, Column: 45},
			},
		},
	}

	output, err := mustCoreParser(t).parseInlines(input, ast.Position{Line: 1, Column: 1})
	if err != nil {
		t.Fatalf("parse returned an error: %v", err)
	}
	if diff := cmp.Diff(want, output); diff != "" {
		t.Errorf("parsed incorrectly.\n(-want +got)\n%s\n", diff)
	}
}

func TestParseUnsupportedInlineRemainsLiteralText(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{name: "unsupported type", input: ":[unknown]{text}"},
		{name: "empty type", input: ":[]{text}"},
		{name: "link without URI", input: ":[link]{text}"},
		{name: "link with empty URI", input: ":[link:]{text}"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := mustCoreParser(t).parseInlines(tt.input, ast.Position{Line: 2, Column: 3})
			if err != nil {
				t.Fatalf("parse returned an error: %v", err)
			}

			want := []ast.Inline{
				&ast.Text{
					Value: tt.input,
					Range: ast.Range{
						Start: ast.Position{Line: 2, Column: 3},
						End: ast.Position{
							Line:   2,
							Column: 2 + utf8.RuneCountInString(tt.input),
						},
					},
				},
			}
			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("parsed incorrectly\n(-want +got)\n%s", diff)
			}
		})
	}
}

func TestParseUnsupportedInlineRemainsLiteralWhenNested(t *testing.T) {
	input := ":[em]{a :[unknown]{b} c}"

	want := []ast.Inline{
		&ast.Emphasis{
			Content: []ast.Inline{
				&ast.Text{
					Value: "a :[unknown]{b} c",
					Range: ast.Range{
						Start: ast.Position{Line: 1, Column: 7},
						End:   ast.Position{Line: 1, Column: 23},
					},
				},
			},
			Range: ast.Range{
				Start: ast.Position{Line: 1, Column: 1},
				End:   ast.Position{Line: 1, Column: 24},
			},
		},
	}

	got, err := mustCoreParser(t).parseInlines(input, ast.Position{Line: 1, Column: 1})
	if err != nil {
		t.Fatalf("parse returned an error: %v", err)
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("parsed incorrectly\n(-want +got)\n%s", diff)
	}
}

func TestParseInvalidInlines(t *testing.T) {
	input := ":[em]{There is no end brace."

	want := []ast.Inline{
		&ast.Text{
			Value: ":[em]{There is no end brace.",
			Range: ast.Range{
				Start: ast.Position{Line: 1, Column: 1},
				End:   ast.Position{Line: 1, Column: 28},
			},
		},
	}

	output, err := mustCoreParser(t).parseInlines(input, ast.Position{Line: 1, Column: 1})

	if err != nil {
		t.Fatalf("parse returned an error: %v", err)
	}
	if diff := cmp.Diff(want, output); diff != "" {
		t.Errorf("parsed incorrectly.\n(-want +got)\n%s\n", diff)
	}
}

func TestParseInlines_ContainingNewLine(t *testing.T) {
	input := "Plain text\n:[em]{emphasized text}"

	want := []ast.Inline{
		&ast.Text{
			Value: "Plain text",
			Range: ast.Range{
				Start: ast.Position{Line: 1, Column: 1},
				End:   ast.Position{Line: 1, Column: 10},
			},
		},
		&ast.Emphasis{
			Content: []ast.Inline{
				&ast.Text{
					Value: "emphasized text",
					Range: ast.Range{
						Start: ast.Position{Line: 2, Column: 7},
						End:   ast.Position{Line: 2, Column: 21},
					},
				},
			},
			Range: ast.Range{
				Start: ast.Position{Line: 2, Column: 1},
				End:   ast.Position{Line: 2, Column: 22},
			},
		},
	}

	output, err := mustCoreParser(t).parseInlines(input, ast.Position{Line: 1, Column: 1})

	if err != nil {
		t.Fatalf("parse returned an error: %v", err)
	}

	if diff := cmp.Diff(want, output); diff != "" {
		t.Errorf("parsed incorrectly\n(-want +got)\n%s", diff)
	}
}

func TestParseInlines_UnicodeRangeWithNonDefaultOrigin(t *testing.T) {
	input := "日😀 :[em]{é界} 後"
	origin := ast.Position{Line: 4, Column: 7}

	want := []ast.Inline{
		&ast.Text{
			Value: "日😀 ",
			Range: ast.Range{
				Start: ast.Position{Line: 4, Column: 7},
				End:   ast.Position{Line: 4, Column: 9},
			},
		},
		&ast.Emphasis{
			Content: []ast.Inline{
				&ast.Text{
					Value: "é界",
					Range: ast.Range{
						Start: ast.Position{Line: 4, Column: 16},
						End:   ast.Position{Line: 4, Column: 17},
					},
				},
			},
			Range: ast.Range{
				Start: ast.Position{Line: 4, Column: 10},
				End:   ast.Position{Line: 4, Column: 18},
			},
		},
		&ast.Text{
			Value: " 後",
			Range: ast.Range{
				Start: ast.Position{Line: 4, Column: 19},
				End:   ast.Position{Line: 4, Column: 20},
			},
		},
	}

	got, err := mustCoreParser(t).parseInlines(input, origin)
	if err != nil {
		t.Fatalf("parse returned an error: %v", err)
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("parsed incorrectly\n(-want +got)\n%s", diff)
	}
}
