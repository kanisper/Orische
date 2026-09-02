package parser

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"orische/internal/ast"
)

func TestBuiltInInlineDefinitions(t *testing.T) {
	definitions := coreInlineDefinitions()
	if len(definitions) != 8 {
		t.Fatalf("definition count = %d, want 8", len(definitions))
	}

	tests := []struct {
		typ       string
		policy    inlineContentPolicy
		attribute string
		accepted  bool
		want      ast.Inline
	}{
		{"em", inlineContentNested, "ignored", true, &ast.Emphasis{Content: []ast.Inline{&ast.Text{Value: "x"}}}},
		{"strong", inlineContentNested, "ignored", true, &ast.Strong{Content: []ast.Inline{&ast.Text{Value: "x"}}}},
		{"italic", inlineContentNested, "ignored", true, &ast.Italic{Content: []ast.Inline{&ast.Text{Value: "x"}}}},
		{"bold", inlineContentNested, "ignored", true, &ast.Bold{Content: []ast.Inline{&ast.Text{Value: "x"}}}},
		{"del", inlineContentNested, "ignored", true, &ast.Deleted{Content: []ast.Inline{&ast.Text{Value: "x"}}}},
		{"outdated", inlineContentNested, "ignored", true, &ast.Outdated{Content: []ast.Inline{&ast.Text{Value: "x"}}}},
		{"link", inlineContentNested, "/x", true, &ast.Link{URI: "/x", Content: []ast.Inline{&ast.Text{Value: "x"}}}},
		{"code", inlineContentLiteral, "ignored", true, &ast.CodeSpan{Value: "x"}},
	}

	candidate := inlineCandidate{
		nestedContent:  []ast.Inline{&ast.Text{Value: "x"}},
		literalContent: "x",
	}
	for _, tt := range tests {
		definition, ok := definitions[tt.typ]
		if !ok {
			t.Fatalf("definition %q is missing", tt.typ)
		}
		if got := definition.policy; got != tt.policy {
			t.Errorf("definition policy = %d, want %d", got, tt.policy)
		}
		accepted := true
		if definition.validate != nil {
			accepted = definition.validate(tt.attribute)
		}
		if accepted != tt.accepted {
			t.Errorf("validate(%q) = %t, want %t", tt.attribute, accepted, tt.accepted)
		}
		candidate.attribute = tt.attribute
		got := definition.build(candidate)
		if diff := cmp.Diff(tt.want, got); diff != "" {
			t.Errorf("inline mismatch (-want +got):\n%s", diff)
		}
	}
}

func TestLinkRejectsEmptyURI(t *testing.T) {
	if linkDefinition.validate == nil || linkDefinition.validate("") {
		t.Error("ValidateAttribute accepted an empty URI")
	}
}
