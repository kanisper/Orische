package parser

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"orische/internal/ast"
	"orische/internal/parser/feature"
	"orische/internal/parser/syntax"
)

func TestInlineDefinitionUsesActiveLanguageDuringRecursion(t *testing.T) {
	definition := &testInlineDirectiveDefinition{
		typ:    "wrap",
		policy: feature.InlineContentNested,
		build: func(candidate feature.InlineDirectiveCandidate) (ast.Inline, error) {
			return &ast.Emphasis{Content: candidate.NestedContent, Range: candidate.Range}, nil
		},
	}
	p := mustParserWithAdditionalInlines(t, definition)

	got, err := p.parseInlines(":[WRAP]{外 :[wrap]{内}}", ast.Position{Line: 3, Column: 4})
	if err != nil {
		t.Fatalf("parseInlines returned an error: %v", err)
	}
	want := []ast.Inline{
		&ast.Emphasis{
			Content: []ast.Inline{
				&ast.Text{Value: "外 ", Range: ast.Range{Start: ast.Position{Line: 3, Column: 12}, End: ast.Position{Line: 3, Column: 13}}},
				&ast.Emphasis{
					Content: []ast.Inline{
						&ast.Text{Value: "内", Range: ast.Range{Start: ast.Position{Line: 3, Column: 22}, End: ast.Position{Line: 3, Column: 22}}},
					},
					Range: ast.Range{Start: ast.Position{Line: 3, Column: 14}, End: ast.Position{Line: 3, Column: 23}},
				},
			},
			Range: ast.Range{Start: ast.Position{Line: 3, Column: 4}, End: ast.Position{Line: 3, Column: 24}},
		},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("inline mismatch (-want +got):\n%s", diff)
	}
}

func TestLiteralInlineDefinitionDoesNotParseNestedSyntax(t *testing.T) {
	definition := &testInlineDirectiveDefinition{
		typ:    "literal",
		policy: feature.InlineContentLiteral,
		build: func(candidate feature.InlineDirectiveCandidate) (ast.Inline, error) {
			return &ast.CodeSpan{Value: candidate.LiteralContent, Range: candidate.Range}, nil
		},
	}
	p := mustParserWithAdditionalInlines(t, definition)

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
	reject := &testInlineDirectiveDefinition{
		typ:    "reject",
		policy: feature.InlineContentNested,
		validate: func(string) (bool, error) {
			return false, nil
		},
	}
	p := mustParserWithAdditionalInlines(t, reject)

	got, err := p.parseInlines(":[reject]{no} :[EM]{日}", ast.Position{Line: 1, Column: 1})
	if err != nil {
		t.Fatalf("parseInlines returned an error: %v", err)
	}
	want := []ast.Inline{
		&ast.Text{Value: ":[reject]{no} ", Range: ast.Range{Start: ast.Position{Line: 1, Column: 1}, End: ast.Position{Line: 1, Column: 14}}},
		&ast.Emphasis{
			Content: []ast.Inline{
				&ast.Text{Value: "日", Range: ast.Range{Start: ast.Position{Line: 1, Column: 21}, End: ast.Position{Line: 1, Column: 21}}},
			},
			Range: ast.Range{Start: ast.Position{Line: 1, Column: 15}, End: ast.Position{Line: 1, Column: 22}},
		},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("inline mismatch (-want +got):\n%s", diff)
	}
}

func TestInlineValidationErrorStopsWithoutFallback(t *testing.T) {
	wantErr := errors.New("validation failure")
	definition := &testInlineDirectiveDefinition{
		typ:    "broken",
		policy: feature.InlineContentNested,
		validate: func(string) (bool, error) {
			return false, wantErr
		},
	}
	p := mustParserWithAdditionalInlines(t, definition)

	got, err := p.parseInlines(":[broken]{text} :[em]{later}", ast.Position{Line: 1, Column: 1})
	if got != nil {
		t.Errorf("parseInlines returned nodes: %#v", got)
	}
	if !errors.Is(err, wantErr) {
		t.Errorf("parseInlines error = %v, want %v", err, wantErr)
	}
}

func TestInlineValidatesOnlyClosedCandidates(t *testing.T) {
	calls := 0
	definition := &testInlineDirectiveDefinition{
		typ:    "probe",
		policy: feature.InlineContentNested,
		validate: func(string) (bool, error) {
			calls++
			return true, nil
		},
		build: func(candidate feature.InlineDirectiveCandidate) (ast.Inline, error) {
			return &ast.Emphasis{Content: candidate.NestedContent, Range: candidate.Range}, nil
		},
	}
	p := mustParserWithAdditionalInlines(t, definition)

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

func TestInlineConstructionErrorIsNotFallback(t *testing.T) {
	wantErr := errors.New("construction failure")
	definition := &testInlineDirectiveDefinition{
		typ:    "broken",
		policy: feature.InlineContentNested,
		build: func(feature.InlineDirectiveCandidate) (ast.Inline, error) {
			return nil, wantErr
		},
	}
	p := mustParserWithAdditionalInlines(t, definition)

	got, err := p.parseInlines(":[broken]{text}", ast.Position{Line: 1, Column: 1})
	if got != nil {
		t.Errorf("parseInlines returned nodes: %#v", got)
	}
	if !errors.Is(err, wantErr) || !strings.Contains(err.Error(), `build inline directive "broken"`) {
		t.Errorf("parseInlines error = %v, want wrapped construction error", err)
	}
}

func TestCoreInlineTypesAreCaseInsensitiveWithoutNormalizingValues(t *testing.T) {
	p := mustCoreParser(t)
	origin := ast.Position{Line: 2, Column: 3}
	lower := ":[em:Ä]{日 :[link:/X:Y]{界}} :[code:Go]{値}"
	mixed := ":[Em:Ä]{日 :[LiNk:/X:Y]{界}} :[CoDe:Go]{値}"

	want, err := p.parseInlines(lower, origin)
	if err != nil {
		t.Fatalf("lowercase parseInlines returned an error: %v", err)
	}
	got, err := p.parseInlines(mixed, origin)
	if err != nil {
		t.Fatalf("mixed-case parseInlines returned an error: %v", err)
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

func TestUnterminatedLiteralCandidateSkipsDefinitionCallbacks(t *testing.T) {
	validateCalls := 0
	buildCalls := 0
	definition := &testInlineDirectiveDefinition{
		typ:    "literal",
		policy: feature.InlineContentLiteral,
		validate: func(string) (bool, error) {
			validateCalls++
			return true, nil
		},
		build: func(feature.InlineDirectiveCandidate) (ast.Inline, error) {
			buildCalls++
			return &ast.CodeSpan{}, nil
		},
	}
	p := mustParserWithAdditionalInlines(t, definition)
	input := ":[literal]{unterminated"

	got, err := p.parseInlines(input, ast.Position{Line: 1, Column: 1})
	if err != nil {
		t.Fatalf("parseInlines returned an error: %v", err)
	}
	if validateCalls != 0 || buildCalls != 0 {
		t.Errorf("callbacks = validate %d, build %d; want zero", validateCalls, buildCalls)
	}
	want := []ast.Inline{
		&ast.Text{
			Value: input,
			Range: ast.Range{Start: ast.Position{Line: 1, Column: 1}, End: ast.Position{Line: 1, Column: len([]rune(input))}},
		},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("inline mismatch (-want +got):\n%s", diff)
	}
}

func TestInlineDefinitionReturningNilNodeIsInternalError(t *testing.T) {
	tests := []struct {
		name  string
		build func(feature.InlineDirectiveCandidate) (ast.Inline, error)
	}{
		{
			name: "nil interface",
			build: func(feature.InlineDirectiveCandidate) (ast.Inline, error) {
				return nil, nil
			},
		},
		{
			name: "typed nil",
			build: func(feature.InlineDirectiveCandidate) (ast.Inline, error) {
				var node *ast.Emphasis
				return node, nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			definition := &testInlineDirectiveDefinition{
				typ:    "nil",
				policy: feature.InlineContentNested,
				build:  tt.build,
			}
			p := mustParserWithAdditionalInlines(t, definition)
			got, err := p.parseInlines(":[nil]{text}", ast.Position{Line: 1, Column: 1})
			if got != nil {
				t.Errorf("parseInlines returned nodes: %#v", got)
			}
			if err == nil || !strings.Contains(err.Error(), `build inline directive "nil": definition returned a nil node`) {
				t.Errorf("parseInlines error = %v, want nil-node internal error", err)
			}
		})
	}
}

func mustParserWithAdditionalInlines(t testing.TB, definitions ...feature.InlineDirectiveDefinition) *Parser {
	t.Helper()
	language := syntax.Core()
	language.Inlines = append(language.Inlines, definitions...)
	p, err := NewParser(language)
	if err != nil {
		t.Fatalf("NewParser returned an error: %v", err)
	}
	return p
}
