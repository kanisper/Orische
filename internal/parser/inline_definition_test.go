package parser

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"orische/internal/ast"
)

func TestNestedInlineUsesConfiguredDefinitionDuringRecursion(t *testing.T) {
	p := parserWithInlineDefinitions(t, map[string]inlineDefinition{
		"wrap": {
			policy: inlineContentNested,
			build: func(candidate inlineCandidate) ast.Inline {
				return &ast.Emphasis{Content: candidate.nestedContent, Range: candidate.rng}
			},
		},
	})

	got, err := p.parseInlines(":[WRAP]{外 :[wrap]{内}}", ast.Position{Line: 3, Column: 4})
	if err != nil {
		t.Fatalf("parseInlines returned an error: %v", err)
	}
	want := []ast.Inline{
		&ast.Emphasis{
			Content: []ast.Inline{
				&ast.Text{Value: "外 ", Range: ast.Range{Start: ast.Position{Line: 3, Column: 12}, End: ast.Position{Line: 3, Column: 13}}},
				&ast.Emphasis{
					Content: []ast.Inline{&ast.Text{Value: "内", Range: ast.Range{Start: ast.Position{Line: 3, Column: 22}, End: ast.Position{Line: 3, Column: 22}}}},
					Range:   ast.Range{Start: ast.Position{Line: 3, Column: 14}, End: ast.Position{Line: 3, Column: 23}},
				},
			},
			Range: ast.Range{Start: ast.Position{Line: 3, Column: 4}, End: ast.Position{Line: 3, Column: 24}},
		},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("inline mismatch (-want +got):\n%s", diff)
	}
}

func TestLiteralInlineContentDoesNotParseNestedSyntax(t *testing.T) {
	p := parserWithInlineDefinitions(t, map[string]inlineDefinition{
		"literal": {
			policy: inlineContentLiteral,
			build: func(candidate inlineCandidate) ast.Inline {
				return &ast.CodeSpan{Value: candidate.literalContent, Range: candidate.rng}
			},
		},
	})

	got, err := p.parseInlines(":[literal]{:[em]{内}} tail", ast.Position{Line: 1, Column: 1})
	if err != nil {
		t.Fatalf("parseInlines returned an error: %v", err)
	}
	want := []ast.Inline{
		&ast.CodeSpan{Value: ":[em]{内", Range: ast.Range{Start: ast.Position{Line: 1, Column: 1}, End: ast.Position{Line: 1, Column: 19}}},
		&ast.Text{Value: "} tail", Range: ast.Range{Start: ast.Position{Line: 1, Column: 20}, End: ast.Position{Line: 1, Column: 25}}},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("inline mismatch (-want +got):\n%s", diff)
	}
}

func TestInlineSemanticRejectionFallsBackAndContinues(t *testing.T) {
	p := parserWithInlineDefinitions(t, map[string]inlineDefinition{
		"reject": {
			policy: inlineContentNested,
			validate: func(string) bool {
				return false
			},
		},
	})

	got, err := p.parseInlines(":[reject]{no} :[EM]{日}", ast.Position{Line: 1, Column: 1})
	if err != nil {
		t.Fatalf("parseInlines returned an error: %v", err)
	}
	want := []ast.Inline{
		&ast.Text{Value: ":[reject]{no} ", Range: ast.Range{Start: ast.Position{Line: 1, Column: 1}, End: ast.Position{Line: 1, Column: 14}}},
		&ast.Emphasis{
			Content: []ast.Inline{&ast.Text{Value: "日", Range: ast.Range{Start: ast.Position{Line: 1, Column: 21}, End: ast.Position{Line: 1, Column: 21}}}},
			Range:   ast.Range{Start: ast.Position{Line: 1, Column: 15}, End: ast.Position{Line: 1, Column: 22}},
		},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("inline mismatch (-want +got):\n%s", diff)
	}
}

func TestInlineValidationOnlyRunsForClosedCandidates(t *testing.T) {
	calls := 0
	p := parserWithInlineDefinitions(t, map[string]inlineDefinition{
		"probe": {
			policy: inlineContentNested,
			validate: func(string) bool {
				calls++
				return true
			},
			build: func(candidate inlineCandidate) ast.Inline {
				return &ast.Emphasis{Content: candidate.nestedContent, Range: candidate.rng}
			},
		},
	})

	got, err := p.parseInlines(":[probe]{broken :[em]{later}", ast.Position{Line: 1, Column: 1})
	if err != nil {
		t.Fatalf("parseInlines returned an error: %v", err)
	}
	if calls != 0 {
		t.Errorf("validator calls = %d, want 0", calls)
	}
	if len(got) != 2 {
		t.Fatalf("inline count = %d, want literal text and later emphasis", len(got))
	}
	if _, ok := got[1].(*ast.Emphasis); !ok {
		t.Errorf("later inline type = %T, want *ast.Emphasis", got[1])
	}
}

func TestInlineWithoutValidatorAcceptsAttribute(t *testing.T) {
	p := parserWithInlineDefinitions(t, map[string]inlineDefinition{
		"accept": {
			policy: inlineContentLiteral,
			build: func(candidate inlineCandidate) ast.Inline {
				return &ast.CodeSpan{Value: candidate.literalContent, Range: candidate.rng}
			},
		},
	})

	got, err := p.parseInlines(":[accept:arbitrary]{value}", ast.Position{Line: 1, Column: 1})
	if err != nil {
		t.Fatalf("parseInlines returned an error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("inline count = %d, want 1", len(got))
	}
	if span, ok := got[0].(*ast.CodeSpan); !ok || span.Value != "value" {
		t.Errorf("inline = %#v, want CodeSpan with value", got[0])
	}
}
