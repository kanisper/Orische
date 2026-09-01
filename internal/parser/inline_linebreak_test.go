package parser

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"orische/internal/ast"
)

func TestParseExplicitLineBreakAcrossLogicalNewlines(t *testing.T) {
	want := []ast.Inline{
		testText("a", 3, 4, 3, 4),
		&ast.LineBreak{
			Range: testRange(3, 5, 3, 6),
		},
		testText("b", 4, 1, 4, 1),
	}

	for _, lineEnding := range []string{"\n", "\r\n", "\r"} {
		got, err := mustCoreParser(t).parseInlines(
			"a +"+lineEnding+"b",
			ast.Position{Line: 3, Column: 4},
		)
		if err != nil {
			t.Fatalf("parseInlines(%q) returned an error: %v", lineEnding, err)
		}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("line ending %q mismatch (-want +got):\n%s", lineEnding, diff)
		}
	}
}

func TestTrailingLineBreakReachesInlineCapableBlocks(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{name: "paragraph fallback", input: "a +\n"},
		{name: "heading sugar", input: "= a +\n"},
		{name: "list item", input: "* a +\n"},
		{name: "paragraph directive", input: ":::[paragraph]\na +\n:::"},
		{name: "heading directive", input: ":::[heading:level1]\na +\n:::"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := Parse(tt.input)
			if err != nil {
				t.Fatalf("Parse returned an error: %v", err)
			}
			content := firstInlineContent(t, doc.Blocks[0])
			if len(content) != 2 {
				t.Fatalf("inline count = %d, want Text and LineBreak", len(content))
			}
			if _, ok := content[1].(*ast.LineBreak); !ok {
				t.Errorf("last inline type = %T, want *ast.LineBreak", content[1])
			}
		})
	}
}

func TestExplicitLineBreakRequiresSpacePlusAndLogicalNewline(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "plus in line", input: "a + b", want: "a + b"},
		{name: "plus without space", input: "a+\nb", want: "a+b"},
		{name: "marker at EOF", input: "a +", want: "a +"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := mustCoreParser(t).parseInlines(tt.input, ast.Position{Line: 1, Column: 1})
			if err != nil {
				t.Fatalf("parseInlines returned an error: %v", err)
			}

			var value string
			for _, node := range got {
				text, ok := node.(*ast.Text)
				if !ok {
					t.Fatalf("node type = %T, want only Text nodes", node)
				}
				value += text.Value
			}
			if value != tt.want {
				t.Errorf("visible text = %q, want %q", value, tt.want)
			}
		})
	}
}

func TestParseExplicitLineBreakInNestedContent(t *testing.T) {
	got, err := mustCoreParser(t).parseInlines(
		":[em]{a +\nb}",
		ast.Position{Line: 1, Column: 1},
	)
	if err != nil {
		t.Fatalf("parseInlines returned an error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("inline count = %d, want emphasis", len(got))
	}

	emphasis, ok := got[0].(*ast.Emphasis)
	if !ok {
		t.Fatalf("first inline type = %T, want *ast.Emphasis", got[0])
	}
	assertInlineBreakContent(t, emphasis.Content, "a", "b")
}

func TestParseExplicitLineBreakUnicodeRange(t *testing.T) {
	got, err := mustCoreParser(t).parseInlines(
		"日😀 +\n界",
		ast.Position{Line: 2, Column: 3},
	)
	if err != nil {
		t.Fatalf("parseInlines returned an error: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("inline count = %d, want 3", len(got))
	}

	lineBreak, ok := got[1].(*ast.LineBreak)
	if !ok {
		t.Fatalf("middle inline type = %T, want *ast.LineBreak", got[1])
	}
	want := testRange(2, 5, 2, 6)
	if diff := cmp.Diff(want, lineBreak.Range); diff != "" {
		t.Errorf("LineBreak range mismatch (-want +got):\n%s", diff)
	}
}

func TestCodeSpanKeepsLineBreakMarkerLiteral(t *testing.T) {
	got, err := mustCoreParser(t).parseInlines(":[code]{a +\nb}", ast.Position{Line: 1, Column: 1})
	if err != nil {
		t.Fatalf("parseInlines returned an error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("inline count = %d, want 1", len(got))
	}
	code, ok := got[0].(*ast.CodeSpan)
	if !ok {
		t.Fatalf("inline type = %T, want *ast.CodeSpan", got[0])
	}
	if code.Value != "a +\nb" {
		t.Errorf("CodeSpan value = %q, want literal marker and newline", code.Value)
	}
}

func TestUnsupportedDirectiveKeepsNestedLineBreakMarkerLiteral(t *testing.T) {
	got, err := mustCoreParser(t).parseInlines(
		":[unknown]{a +\nb} +\nc",
		ast.Position{Line: 1, Column: 1},
	)
	if err != nil {
		t.Fatalf("parseInlines returned an error: %v", err)
	}
	if len(got) != 4 {
		t.Fatalf("inline count = %d, want two Text nodes, LineBreak, and Text", len(got))
	}

	first, firstOK := got[0].(*ast.Text)
	second, secondOK := got[1].(*ast.Text)
	_, breakOK := got[2].(*ast.LineBreak)
	last, lastOK := got[3].(*ast.Text)
	if !firstOK || !secondOK || !breakOK || !lastOK {
		t.Fatalf("inline types = (%T, %T, %T, %T), want Text, Text, LineBreak, Text", got[0], got[1], got[2], got[3])
	}
	if first.Value != ":[unknown]{a +" || second.Value != "b}" || last.Value != "c" {
		t.Errorf("text values = (%q, %q, %q), want unsupported candidate literal around one outer break", first.Value, second.Value, last.Value)
	}
}

func assertInlineBreakContent(t testing.TB, content []ast.Inline, before, after string) {
	t.Helper()
	if len(content) != 3 {
		t.Fatalf("nested inline count = %d, want 3", len(content))
	}
	first, firstOK := content[0].(*ast.Text)
	lineBreak, breakOK := content[1].(*ast.LineBreak)
	last, lastOK := content[2].(*ast.Text)
	if !firstOK || !breakOK || !lastOK {
		t.Fatalf("nested inline types = (%T, %T, %T), want Text, LineBreak, Text", content[0], content[1], content[2])
	}
	if first.Value != before || last.Value != after {
		t.Errorf("nested text = (%q, %q), want (%q, %q)", first.Value, last.Value, before, after)
	}
	if lineBreak.Range.Start.Line == 0 {
		t.Error("LineBreak range is empty")
	}
}

func firstInlineContent(t testing.TB, block ast.Block) []ast.Inline {
	t.Helper()
	switch block := block.(type) {
	case *ast.Heading:
		return block.Content
	case *ast.Paragraph:
		return block.Content
	case *ast.List:
		paragraph, ok := block.Items[0].Blocks[0].(*ast.Paragraph)
		if !ok {
			t.Fatalf("list item block type = %T, want *ast.Paragraph", block.Items[0].Blocks[0])
		}
		return paragraph.Content
	default:
		t.Fatalf("block type = %T, want inline-capable block", block)
		return nil
	}
}
