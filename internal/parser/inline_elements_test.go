package parser

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"orische/internal/ast"
)

func TestParseExplicitStyledInlines(t *testing.T) {
	origin := ast.Position{Line: 3, Column: 4}
	input := ":[em:x]{強} :[strong:x]{重} :[italic:x]{斜} :[bold:x]{太} :[del:x]{削} :[outdated:x]{古}"
	want := []ast.Inline{
		&ast.Emphasis{
			Content: []ast.Inline{&ast.Text{Value: "強", Range: ast.Range{Start: ast.Position{Line: 3, Column: 12}, End: ast.Position{Line: 3, Column: 12}}}},
			Range:   ast.Range{Start: ast.Position{Line: 3, Column: 4}, End: ast.Position{Line: 3, Column: 13}},
		},
		&ast.Text{Value: " ", Range: ast.Range{Start: ast.Position{Line: 3, Column: 14}, End: ast.Position{Line: 3, Column: 14}}},
		&ast.Strong{
			Content: []ast.Inline{&ast.Text{Value: "重", Range: ast.Range{Start: ast.Position{Line: 3, Column: 27}, End: ast.Position{Line: 3, Column: 27}}}},
			Range:   ast.Range{Start: ast.Position{Line: 3, Column: 15}, End: ast.Position{Line: 3, Column: 28}},
		},
		&ast.Text{Value: " ", Range: ast.Range{Start: ast.Position{Line: 3, Column: 29}, End: ast.Position{Line: 3, Column: 29}}},
		&ast.Italic{
			Content: []ast.Inline{&ast.Text{Value: "斜", Range: ast.Range{Start: ast.Position{Line: 3, Column: 42}, End: ast.Position{Line: 3, Column: 42}}}},
			Range:   ast.Range{Start: ast.Position{Line: 3, Column: 30}, End: ast.Position{Line: 3, Column: 43}},
		},
		&ast.Text{Value: " ", Range: ast.Range{Start: ast.Position{Line: 3, Column: 44}, End: ast.Position{Line: 3, Column: 44}}},
		&ast.Bold{
			Content: []ast.Inline{&ast.Text{Value: "太", Range: ast.Range{Start: ast.Position{Line: 3, Column: 55}, End: ast.Position{Line: 3, Column: 55}}}},
			Range:   ast.Range{Start: ast.Position{Line: 3, Column: 45}, End: ast.Position{Line: 3, Column: 56}},
		},
		&ast.Text{Value: " ", Range: ast.Range{Start: ast.Position{Line: 3, Column: 57}, End: ast.Position{Line: 3, Column: 57}}},
		&ast.Deleted{
			Content: []ast.Inline{&ast.Text{Value: "削", Range: ast.Range{Start: ast.Position{Line: 3, Column: 67}, End: ast.Position{Line: 3, Column: 67}}}},
			Range:   ast.Range{Start: ast.Position{Line: 3, Column: 58}, End: ast.Position{Line: 3, Column: 68}},
		},
		&ast.Text{Value: " ", Range: ast.Range{Start: ast.Position{Line: 3, Column: 69}, End: ast.Position{Line: 3, Column: 69}}},
		&ast.Outdated{
			Content: []ast.Inline{&ast.Text{Value: "古", Range: ast.Range{Start: ast.Position{Line: 3, Column: 84}, End: ast.Position{Line: 3, Column: 84}}}},
			Range:   ast.Range{Start: ast.Position{Line: 3, Column: 70}, End: ast.Position{Line: 3, Column: 85}},
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
	input := ":[del]{outer :[outdated]{inner}}"
	want := []ast.Inline{
		&ast.Deleted{
			Content: []ast.Inline{
				&ast.Text{Value: "outer ", Range: ast.Range{Start: ast.Position{Line: 1, Column: 8}, End: ast.Position{Line: 1, Column: 13}}},
				&ast.Outdated{
					Content: []ast.Inline{&ast.Text{Value: "inner", Range: ast.Range{Start: ast.Position{Line: 1, Column: 26}, End: ast.Position{Line: 1, Column: 30}}}},
					Range:   ast.Range{Start: ast.Position{Line: 1, Column: 14}, End: ast.Position{Line: 1, Column: 31}},
				},
			},
			Range: ast.Range{Start: ast.Position{Line: 1, Column: 1}, End: ast.Position{Line: 1, Column: 32}},
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

func TestRemovedStyledDirectivesFallBackToLiteralText(t *testing.T) {
	for _, input := range []string{":[underline]{x}", ":[strike]{x}"} {
		t.Run(input, func(t *testing.T) {
			assertOnlyLiteralText(t, input)
		})
	}
}
