package parser

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"

	"orische/internal/ast"
)

func TestSpec_InlineRegistrationRejectsNormalizedDuplicateWithoutOverwrite(t *testing.T) {
	spec := newSpec()
	first := &testInlineDefinition{policy: inlineContentNested}
	second := &testInlineDefinition{policy: inlineContentLiteral}

	if err := spec.registerInlineDirectiveDefinition("ÄBC", first); err != nil {
		t.Fatalf("first registration returned an error: %v", err)
	}
	if err := spec.registerInlineDirectiveDefinition("äbc", second); err == nil {
		t.Fatal("case-only duplicate registration returned no error")
	}

	got, ok := spec.getInlineDirectiveDefinition("ÄbC")
	if !ok {
		t.Fatal("normalized definition lookup failed")
	}
	if got != first {
		t.Errorf("duplicate registration replaced definition with %T", got)
	}
}

func TestSpec_InlineRegistrationRejectsIncompleteDefinition(t *testing.T) {
	spec := newSpec()
	if err := spec.registerInlineDirectiveDefinition("", &testInlineDefinition{}); err == nil {
		t.Error("empty directive type registration returned no error")
	}
	if err := spec.registerInlineDirectiveDefinition("missing", nil); err == nil {
		t.Error("nil definition registration returned no error")
	}
	if err := spec.registerInlineDirectiveDefinition("invalid-policy", &testInlineDefinition{
		policy: inlineContentPolicy(99),
	}); err == nil {
		t.Error("invalid content policy registration returned no error")
	}
	if _, ok := spec.getInlineDirectiveDefinition("missing"); ok {
		t.Error("nil definition registration installed a definition")
	}
	if _, ok := spec.getInlineDirectiveDefinition("invalid-policy"); ok {
		t.Error("invalid content policy registration installed a definition")
	}
}

func TestParserParseInlines_UsesActiveSpecForNestedDefinitions(t *testing.T) {
	spec := newSpec()
	definition := &testInlineDefinition{
		policy: inlineContentNested,
		build: func(candidate inlineDirectiveCandidate) (ast.Inline, error) {
			return &ast.Emphasis{
				Content: candidate.NestedContent,
				Range:   candidate.Range,
			}, nil
		},
	}
	if err := spec.registerInlineDirectiveDefinition("wrap", definition); err != nil {
		t.Fatalf("register wrap definition: %v", err)
	}

	got, err := NewParser(spec).parseInlines(":[WRAP]{外 :[wrap]{内}}", ast.Position{Line: 3, Column: 4})
	if err != nil {
		t.Fatalf("parseInlines returned an error: %v", err)
	}

	want := []ast.Inline{
		&ast.Emphasis{
			Content: []ast.Inline{
				&ast.Text{
					Value: "外 ",
					Range: ast.Range{
						Start: ast.Position{Line: 3, Column: 12},
						End:   ast.Position{Line: 3, Column: 13},
					},
				},
				&ast.Emphasis{
					Content: []ast.Inline{
						&ast.Text{
							Value: "内",
							Range: ast.Range{
								Start: ast.Position{Line: 3, Column: 22},
								End:   ast.Position{Line: 3, Column: 22},
							},
						},
					},
					Range: ast.Range{
						Start: ast.Position{Line: 3, Column: 14},
						End:   ast.Position{Line: 3, Column: 23},
					},
				},
			},
			Range: ast.Range{
				Start: ast.Position{Line: 3, Column: 4},
				End:   ast.Position{Line: 3, Column: 24},
			},
		},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("parseInlines returned unexpected nodes (-want +got):\n%s", diff)
	}
}

func TestParserParseInlines_LiteralDefinitionDoesNotParseNestedDirectives(t *testing.T) {
	spec := newSpec()
	definition := &testInlineDefinition{
		policy: inlineContentLiteral,
		build: func(candidate inlineDirectiveCandidate) (ast.Inline, error) {
			return &ast.CodeSpan{
				Value: candidate.LiteralContent,
				Range: candidate.Range,
			}, nil
		},
	}
	if err := spec.registerInlineDirectiveDefinition("literal", definition); err != nil {
		t.Fatalf("register literal definition: %v", err)
	}
	if err := spec.registerInlineDirectiveDefinition("wrap", &testInlineDefinition{
		policy: inlineContentNested,
		build: func(candidate inlineDirectiveCandidate) (ast.Inline, error) {
			return &ast.Emphasis{Content: candidate.NestedContent, Range: candidate.Range}, nil
		},
	}); err != nil {
		t.Fatalf("register wrap definition: %v", err)
	}

	got, err := NewParser(spec).parseInlines(":[literal]{:[wrap]{内}}", ast.Position{Line: 1, Column: 1})
	if err != nil {
		t.Fatalf("parseInlines returned an error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("inline count = %d, want 2", len(got))
	}
	code, ok := got[0].(*ast.CodeSpan)
	if !ok {
		t.Fatalf("inline 0 type = %T, want *ast.CodeSpan", got[0])
	}
	if code.Value != ":[wrap]{内" {
		t.Errorf("literal value = %q, want %q", code.Value, ":[wrap]{内")
	}
	text, ok := got[1].(*ast.Text)
	if !ok || text.Value != "}" {
		t.Errorf("remaining inline = %#v, want literal closing brace", got[1])
	}
}

func TestCoreInlineDefinitions_MixedCaseMatchesAndPreservesValuesAndRanges(t *testing.T) {
	lower := "日 :[em]{é :[link:/XyZ]{界}} :[code]{内 :[EM]{外}}"
	mixed := "日 :[EM]{é :[LiNk:/XyZ]{界}} :[CoDe]{内 :[EM]{外}}"
	origin := ast.Position{Line: 5, Column: 7}

	want, err := NewParser(nil).parseInlines(lower, origin)
	if err != nil {
		t.Fatalf("parse lowercase input: %v", err)
	}
	got, err := NewParser(nil).parseInlines(mixed, origin)
	if err != nil {
		t.Fatalf("parse mixed-case input: %v", err)
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("mixed-case definitions differ from lowercase definitions (-want +got):\n%s", diff)
	}

	emphasis := got[1].(*ast.Emphasis)
	link := emphasis.Content[1].(*ast.Link)
	if link.URI != "/XyZ" {
		t.Errorf("link URI = %q, want %q", link.URI, "/XyZ")
	}
	code := got[3].(*ast.CodeSpan)
	if code.Value != "内 :[EM]{外" {
		t.Errorf("code value = %q, want %q", code.Value, "内 :[EM]{外")
	}
}

func TestParserParseInlines_SemanticRejectionRemainsLiteralAndScanningContinues(t *testing.T) {
	spec := newSpec()
	reject := &testInlineDefinition{
		policy: inlineContentNested,
		validate: func(string) (bool, error) {
			return false, nil
		},
	}
	if err := spec.registerInlineDirectiveDefinition("reject", reject); err != nil {
		t.Fatalf("register rejecting definition: %v", err)
	}
	if err := spec.registerInlineDirectiveDefinition("em", &emphasisInlineDefinition{}); err != nil {
		t.Fatalf("register emphasis definition: %v", err)
	}

	got, err := NewParser(spec).parseInlines(":[reject]{no} :[]{bad} :[EM]{日}", ast.Position{Line: 1, Column: 1})
	if err != nil {
		t.Fatalf("parseInlines returned an error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("inline count = %d, want 2", len(got))
	}
	text, ok := got[0].(*ast.Text)
	if !ok || text.Value != ":[reject]{no} :[]{bad} " {
		t.Errorf("literal fallback = %#v, want rejected and malformed candidates as text", got[0])
	}
	if emphasis, ok := got[1].(*ast.Emphasis); !ok || emphasis.Content[0].(*ast.Text).Value != "日" {
		t.Errorf("later valid candidate = %#v, want emphasis", got[1])
	}
}

func TestParserParseInlines_UnterminatedCandidateStillAllowsLaterValidCandidate(t *testing.T) {
	got, err := NewParser(nil).parseInlines(":[em]{broken :[link:/X]{界}", ast.Position{Line: 2, Column: 3})
	if err != nil {
		t.Fatalf("parseInlines returned an error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("inline count = %d, want 2", len(got))
	}

	wantText := &ast.Text{
		Value: ":[em]{broken ",
		Range: ast.Range{
			Start: ast.Position{Line: 2, Column: 3},
			End:   ast.Position{Line: 2, Column: 15},
		},
	}
	if diff := cmp.Diff(wantText, got[0]); diff != "" {
		t.Errorf("unterminated fallback differs (-want +got):\n%s", diff)
	}
	wantLink := &ast.Link{
		URI: "/X",
		Content: []ast.Inline{
			&ast.Text{
				Value: "界",
				Range: ast.Range{
					Start: ast.Position{Line: 2, Column: 27},
					End:   ast.Position{Line: 2, Column: 27},
				},
			},
		},
		Range: ast.Range{
			Start: ast.Position{Line: 2, Column: 16},
			End:   ast.Position{Line: 2, Column: 28},
		},
	}
	if diff := cmp.Diff(wantLink, got[1]); diff != "" {
		t.Errorf("later valid directive differs (-want +got):\n%s", diff)
	}
}

func TestParserParseInlines_ConstructionErrorIsNotLiteralFallback(t *testing.T) {
	wantErr := errors.New("construction failed")
	spec := newSpec()
	definition := &testInlineDefinition{
		policy: inlineContentLiteral,
		build: func(inlineDirectiveCandidate) (ast.Inline, error) {
			return nil, wantErr
		},
	}
	if err := spec.registerInlineDirectiveDefinition("broken", definition); err != nil {
		t.Fatalf("register broken definition: %v", err)
	}

	got, err := NewParser(spec).parseInlines(":[broken]{text}", ast.Position{Line: 1, Column: 1})
	if got != nil {
		t.Errorf("parseInlines returned nodes: %#v", got)
	}
	if !errors.Is(err, wantErr) {
		t.Errorf("error = %v, want %v", err, wantErr)
	}
}

type testInlineDefinition struct {
	policy   inlineContentPolicy
	validate func(string) (bool, error)
	build    func(inlineDirectiveCandidate) (ast.Inline, error)
}

func (d *testInlineDefinition) contentPolicy() inlineContentPolicy {
	return d.policy
}

func (d *testInlineDefinition) validateAttribute(attribute string) (bool, error) {
	if d.validate == nil {
		return true, nil
	}
	return d.validate(attribute)
}

func (d *testInlineDefinition) buildInline(candidate inlineDirectiveCandidate) (ast.Inline, error) {
	if d.build == nil {
		return &ast.CodeSpan{Range: candidate.Range}, nil
	}
	return d.build(candidate)
}
