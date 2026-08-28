package syntax_test

import (
	"testing"

	"orische/internal/parser/syntax"
)

func TestCoreAssemblesBuiltInLanguage(t *testing.T) {
	language := syntax.Core()
	wantBlocks := []string{"code", "heading", "list"}
	if len(language.Blocks) != len(wantBlocks) {
		t.Fatalf("Block count = %d, want %d", len(language.Blocks), len(wantBlocks))
	}
	for i, want := range wantBlocks {
		if got := language.Blocks[i].BlockType(); got != want {
			t.Errorf("Block %d type = %q, want %q", i, got, want)
		}
	}

	wantInlines := []string{"em", "link", "code"}
	if len(language.Inlines) != len(wantInlines) {
		t.Fatalf("Inline count = %d, want %d", len(language.Inlines), len(wantInlines))
	}
	for i, want := range wantInlines {
		if got := language.Inlines[i].InlineType(); got != want {
			t.Errorf("Inline %d type = %q, want %q", i, got, want)
		}
	}
}
