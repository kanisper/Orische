package inline

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"orische/internal/ast"
	"orische/internal/parser/feature"
)

func TestDefinitions(t *testing.T) {
	definitions := Definitions()
	if len(definitions) != 3 {
		t.Fatalf("definition count = %d, want 3", len(definitions))
	}

	tests := []struct {
		index     int
		typ       string
		policy    feature.InlineContentPolicy
		attribute string
		accepted  bool
		want      ast.Inline
	}{
		{0, "em", feature.InlineContentNested, "ignored", true, &ast.Emphasis{Content: []ast.Inline{&ast.Text{Value: "x"}}}},
		{1, "link", feature.InlineContentNested, "/x", true, &ast.Link{URI: "/x", Content: []ast.Inline{&ast.Text{Value: "x"}}}},
		{2, "code", feature.InlineContentLiteral, "ignored", true, &ast.CodeSpan{Value: "x"}},
	}

	candidate := feature.InlineDirectiveCandidate{
		NestedContent:  []ast.Inline{&ast.Text{Value: "x"}},
		LiteralContent: "x",
	}
	for _, tt := range tests {
		definition, ok := definitions[tt.index].(feature.InlineDirectiveDefinition)
		if !ok {
			t.Fatalf("definition %d type = %T, want InlineDirectiveDefinition", tt.index, definitions[tt.index])
		}
		if got := definition.InlineType(); got != tt.typ {
			t.Errorf("InlineType() = %q, want %q", got, tt.typ)
		}
		if got := definition.ContentPolicy(); got != tt.policy {
			t.Errorf("ContentPolicy() = %d, want %d", got, tt.policy)
		}
		accepted, err := definition.ValidateAttribute(tt.attribute)
		if err != nil || accepted != tt.accepted {
			t.Errorf("ValidateAttribute(%q) = %t, %v, want %t, nil", tt.attribute, accepted, err, tt.accepted)
		}
		candidate.Attribute = tt.attribute
		got, err := definition.BuildInline(candidate)
		if err != nil {
			t.Fatalf("BuildInline returned an error: %v", err)
		}
		if diff := cmp.Diff(tt.want, got); diff != "" {
			t.Errorf("inline mismatch (-want +got):\n%s", diff)
		}
	}
}

func TestLinkRejectsEmptyURI(t *testing.T) {
	accepted, err := (&linkDefinition{}).ValidateAttribute("")
	if err != nil {
		t.Fatalf("ValidateAttribute returned an error: %v", err)
	}
	if accepted {
		t.Error("ValidateAttribute accepted an empty URI")
	}
}
