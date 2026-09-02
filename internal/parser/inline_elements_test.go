package parser

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"orische/internal/ast"
)

func TestParseExplicitStyledInlines(t *testing.T) {
	origin := ast.Position{Line: 3, Column: 4}
	input := ":[strong:x]{強} :[italic:x]{斜} :[bold:x]{太} :[underline:x]{下} :[strike:x]{消}"
	want := []ast.Inline{
		&ast.Strong{
			Content: []ast.Inline{&ast.Text{Value: "強", Range: ast.Range{Start: ast.Position{Line: 3, Column: 16}, End: ast.Position{Line: 3, Column: 16}}}},
			Range:   ast.Range{Start: ast.Position{Line: 3, Column: 4}, End: ast.Position{Line: 3, Column: 17}},
		},
		&ast.Text{Value: " ", Range: ast.Range{Start: ast.Position{Line: 3, Column: 18}, End: ast.Position{Line: 3, Column: 18}}},
		&ast.Italic{
			Content: []ast.Inline{&ast.Text{Value: "斜", Range: ast.Range{Start: ast.Position{Line: 3, Column: 31}, End: ast.Position{Line: 3, Column: 31}}}},
			Range:   ast.Range{Start: ast.Position{Line: 3, Column: 19}, End: ast.Position{Line: 3, Column: 32}},
		},
		&ast.Text{Value: " ", Range: ast.Range{Start: ast.Position{Line: 3, Column: 33}, End: ast.Position{Line: 3, Column: 33}}},
		&ast.Bold{
			Content: []ast.Inline{&ast.Text{Value: "太", Range: ast.Range{Start: ast.Position{Line: 3, Column: 44}, End: ast.Position{Line: 3, Column: 44}}}},
			Range:   ast.Range{Start: ast.Position{Line: 3, Column: 34}, End: ast.Position{Line: 3, Column: 45}},
		},
		&ast.Text{Value: " ", Range: ast.Range{Start: ast.Position{Line: 3, Column: 46}, End: ast.Position{Line: 3, Column: 46}}},
		&ast.Underline{
			Content: []ast.Inline{&ast.Text{Value: "下", Range: ast.Range{Start: ast.Position{Line: 3, Column: 62}, End: ast.Position{Line: 3, Column: 62}}}},
			Range:   ast.Range{Start: ast.Position{Line: 3, Column: 47}, End: ast.Position{Line: 3, Column: 63}},
		},
		&ast.Text{Value: " ", Range: ast.Range{Start: ast.Position{Line: 3, Column: 64}, End: ast.Position{Line: 3, Column: 64}}},
		&ast.Strikethrough{
			Content: []ast.Inline{&ast.Text{Value: "消", Range: ast.Range{Start: ast.Position{Line: 3, Column: 77}, End: ast.Position{Line: 3, Column: 77}}}},
			Range:   ast.Range{Start: ast.Position{Line: 3, Column: 65}, End: ast.Position{Line: 3, Column: 78}},
		},
	}

	got, err := mustCoreParser(t).parseInlines(input, origin)
	if err != nil {
		t.Fatalf("parseInlines returned an error: %v", err)
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("inline mismatch (-want +got):\n%s", diff)
	}
}

func TestParseExplicitStyledInlineRecursively(t *testing.T) {
	input := ":[strong]{outer :[italic]{inner}}"
	want := []ast.Inline{
		&ast.Strong{
			Content: []ast.Inline{
				&ast.Text{Value: "outer ", Range: ast.Range{Start: ast.Position{Line: 1, Column: 11}, End: ast.Position{Line: 1, Column: 16}}},
				&ast.Italic{
					Content: []ast.Inline{&ast.Text{Value: "inner", Range: ast.Range{Start: ast.Position{Line: 1, Column: 27}, End: ast.Position{Line: 1, Column: 31}}}},
					Range:   ast.Range{Start: ast.Position{Line: 1, Column: 17}, End: ast.Position{Line: 1, Column: 32}},
				},
			},
			Range: ast.Range{Start: ast.Position{Line: 1, Column: 1}, End: ast.Position{Line: 1, Column: 33}},
		},
	}

	got, err := mustCoreParser(t).parseInlines(input, ast.Position{Line: 1, Column: 1})
	if err != nil {
		t.Fatalf("parseInlines returned an error: %v", err)
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("inline mismatch (-want +got):\n%s", diff)
	}
}
