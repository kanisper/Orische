package parser

import "testing"

func TestSpecIncludesBuiltInInlineDefinitions(t *testing.T) {
	definitions := coreInlineDefinitions()
	wantInlines := []string{"em", "strong", "italic", "bold", "del", "outdated", "link", "code"}
	if len(definitions) != len(wantInlines) {
		t.Fatalf("Inline count = %d, want %d", len(definitions), len(wantInlines))
	}
	for _, want := range wantInlines {
		if _, ok := definitions[want]; !ok {
			t.Errorf("Inline %q is missing", want)
		}
	}
}

func TestSpecLookupNormalizesInlineTypes(t *testing.T) {
	s := newSpec()

	for _, typ := range []string{"em", "EM", "eM"} {
		definition, ok := s.getInlineDirectiveDefinition(typ)
		if !ok {
			t.Fatalf("lookup %q did not find a definition", typ)
		}
		if definition.policy != inlineContentNested {
			t.Errorf("lookup %q policy = %d, want nested", typ, definition.policy)
		}
	}
}

func TestSpecLookupUsesUnicodeCaseNormalization(t *testing.T) {
	s := newSpec()
	s.inlineDefinitions[normalizeSyntaxType("ÄBC")] = inlineDefinition{}

	if _, ok := s.getInlineDirectiveDefinition("äBc"); !ok {
		t.Fatal("Unicode case-insensitive inline lookup failed")
	}
}
