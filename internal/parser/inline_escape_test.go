package parser

import (
	"testing"

	"orische/internal/ast"
)

func TestInlineEscapeAcceptsASCIIPunctuation(t *testing.T) {
	punctuation := "!\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~"

	for _, char := range punctuation {
		t.Run(string(char), func(t *testing.T) {
			got, err := mustCoreParser(t).parseInlines("\\"+string(char), ast.Position{Line: 2, Column: 3})
			if err != nil {
				t.Fatalf("parseInlines returned an error: %v", err)
			}
			if len(got) != 1 {
				t.Fatalf("inline count = %d, want 1", len(got))
			}
			text, ok := got[0].(*ast.Text)
			if !ok {
				t.Fatalf("inline type = %T, want *ast.Text", got[0])
			}
			if text.Value != string(char) {
				t.Errorf("Text.Value = %q, want %q", text.Value, string(char))
			}
			wantRange := ast.Range{
				Start: ast.Position{Line: 2, Column: 3},
				End:   ast.Position{Line: 2, Column: 4},
			}
			if text.Range != wantRange {
				t.Errorf("Text.Range = %#v, want %#v", text.Range, wantRange)
			}
		})
	}
}

func TestInlineEscapeRejectsNonPunctuation(t *testing.T) {
	tests := []struct {
		name  string
		input string
		value string
	}{
		{name: "ASCII letter", input: `\a`, value: `\a`},
		{name: "Unicode punctuation", input: `\。`, value: `\。`},
		{name: "trailing backslash", input: `\`, value: `\`},
		{name: "tab", input: "\\\t", value: "\\\t"},
		{name: "before LF", input: "\\\n", value: `\`},
		{name: "before CRLF", input: "\\\r\n", value: `\`},
		{name: "before CR", input: "\\\r", value: `\`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := mustCoreParser(t).parseInlines(tt.input, ast.Position{Line: 1, Column: 1})
			if err != nil {
				t.Fatalf("parseInlines returned an error: %v", err)
			}
			if len(got) != 1 {
				t.Fatalf("inline count = %d, want 1", len(got))
			}
			text, ok := got[0].(*ast.Text)
			if !ok || text.Value != tt.value {
				t.Errorf("inline = %#v, want Text value %q", got[0], tt.value)
			}
		})
	}
}

func TestInlineEscapeUsesRecursiveReaderDispatch(t *testing.T) {
	got, err := mustCoreParser(t).parseInlines(`:[em]{\*text\*}`, ast.Position{Line: 1, Column: 1})
	if err != nil {
		t.Fatalf("parseInlines returned an error: %v", err)
	}
	emphasis, ok := got[0].(*ast.Emphasis)
	if !ok {
		t.Fatalf("inline type = %T, want *ast.Emphasis", got[0])
	}
	want := []string{"*", "text", "*"}
	if len(emphasis.Content) != len(want) {
		t.Fatalf("nested inline count = %d, want %d", len(emphasis.Content), len(want))
	}
	for i, value := range want {
		text, ok := emphasis.Content[i].(*ast.Text)
		if !ok || text.Value != value {
			t.Errorf("nested inline %d = %#v, want Text value %q", i, emphasis.Content[i], value)
		}
	}
}

func TestInlineEscapeDoesNotAffectLiteralCodeSpan(t *testing.T) {
	got, err := mustCoreParser(t).parseInlines(`:[code]{\*}`, ast.Position{Line: 1, Column: 1})
	if err != nil {
		t.Fatalf("parseInlines returned an error: %v", err)
	}
	code, ok := got[0].(*ast.CodeSpan)
	if !ok || code.Value != `\*` {
		t.Fatalf("inline = %#v, want CodeSpan value %q", got[0], `\*`)
	}
}

func TestEscapedPlusDoesNotCreateLineBreak(t *testing.T) {
	for _, newline := range []string{"\n", "\r\n", "\r"} {
		got, err := mustCoreParser(t).parseInlines(" \\+"+newline, ast.Position{Line: 1, Column: 1})
		if err != nil {
			t.Fatalf("parseInlines returned an error: %v", err)
		}
		if len(got) != 2 {
			t.Fatalf("inline count = %d, want space and escaped plus", len(got))
		}
		for _, node := range got {
			if _, ok := node.(*ast.LineBreak); ok {
				t.Fatalf("escaped plus before %q produced a LineBreak", newline)
			}
		}
		if text, ok := got[1].(*ast.Text); !ok || text.Value != "+" {
			t.Errorf("escaped plus = %#v, want Text value +", got[1])
		}
	}
}

func TestEscapedDirectiveOpeningDoesNotStartDirective(t *testing.T) {
	got, err := mustCoreParser(t).parseInlines(`\:[em]{text}`, ast.Position{Line: 1, Column: 1})
	if err != nil {
		t.Fatalf("parseInlines returned an error: %v", err)
	}
	for _, node := range got {
		if _, ok := node.(*ast.Emphasis); ok {
			t.Fatal("escaped colon started an Inline Directive")
		}
	}
}

func TestInlineEscapeRangeCountsUnicodeCodePoints(t *testing.T) {
	got, err := mustCoreParser(t).parseInlines(`日 \*`, ast.Position{Line: 4, Column: 7})
	if err != nil {
		t.Fatalf("parseInlines returned an error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("inline count = %d, want 2", len(got))
	}
	text, ok := got[1].(*ast.Text)
	if !ok {
		t.Fatalf("inline type = %T, want *ast.Text", got[1])
	}
	want := ast.Range{
		Start: ast.Position{Line: 4, Column: 9},
		End:   ast.Position{Line: 4, Column: 10},
	}
	if text.Range != want {
		t.Errorf("Text.Range = %#v, want %#v", text.Range, want)
	}
}
