package parser

import (
	"reflect"
	"strings"
	"testing"
	"unicode/utf8"

	"orische/internal/ast"
)

func TestInlineSugarBuildsExpectedNodesAndRanges(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		wantType      ast.Inline
		contentOffset int
		wantURI       string
	}{
		{name: "emphasis", input: "*x*", wantType: &ast.Emphasis{}, contentOffset: 1},
		{name: "strong", input: "**x**", wantType: &ast.Strong{}, contentOffset: 2},
		{name: "italic", input: "_x_", wantType: &ast.Italic{}, contentOffset: 1},
		{name: "bold", input: "__x__", wantType: &ast.Bold{}, contentOffset: 2},
		{name: "underline", input: "++x++", wantType: &ast.Underline{}, contentOffset: 2},
		{name: "strikethrough", input: "~~x~~", wantType: &ast.Strikethrough{}, contentOffset: 2},
		{name: "code span", input: "`x`", wantType: &ast.CodeSpan{}},
		{name: "link", input: "[x](https://example.com)", wantType: &ast.Link{}, contentOffset: 1, wantURI: "https://example.com"},
	}

	origin := ast.Position{Line: 2, Column: 3}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := mustCoreParser(t).parseInlines(tt.input, origin)
			if err != nil {
				t.Fatalf("parseInlines returned an error: %v", err)
			}
			if len(got) != 1 {
				t.Fatalf("inline count = %d, want 1", len(got))
			}
			if reflect.TypeOf(got[0]) != reflect.TypeOf(tt.wantType) {
				t.Fatalf("inline type = %T, want %T", got[0], tt.wantType)
			}

			wantRange := ast.Range{
				Start: origin,
				End: ast.Position{
					Line:   origin.Line,
					Column: origin.Column + utf8.RuneCountInString(tt.input) - 1,
				},
			}
			if rng := sugarNodeRange(got[0]); rng != wantRange {
				t.Errorf("Range = %#v, want %#v", rng, wantRange)
			}

			switch node := got[0].(type) {
			case *ast.CodeSpan:
				if node.Value != "x" {
					t.Errorf("CodeSpan.Value = %q, want x", node.Value)
				}
			case *ast.Link:
				if node.URI != tt.wantURI {
					t.Errorf("Link.URI = %q, want %q", node.URI, tt.wantURI)
				}
				assertSugarTextContent(t, node.Content, origin, tt.contentOffset)
			default:
				assertSugarTextContent(t, sugarNodeContent(got[0]), origin, tt.contentOffset)
			}
		})
	}
}

func TestInlineSugarAndExplicitFormsShareSemantics(t *testing.T) {
	tests := []struct {
		sugar    string
		explicit string
	}{
		{sugar: "*x*", explicit: ":[em]{x}"},
		{sugar: "**x**", explicit: ":[strong]{x}"},
		{sugar: "_x_", explicit: ":[italic]{x}"},
		{sugar: "__x__", explicit: ":[bold]{x}"},
		{sugar: "++x++", explicit: ":[underline]{x}"},
		{sugar: "~~x~~", explicit: ":[strike]{x}"},
		{sugar: "`x`", explicit: ":[code]{x}"},
		{sugar: "[x](https://example.com)", explicit: ":[link:https://example.com]{x}"},
	}

	for _, tt := range tests {
		gotSugar, err := mustCoreParser(t).parseInlines(tt.sugar, ast.Position{Line: 1, Column: 1})
		if err != nil {
			t.Fatalf("parse sugar %q: %v", tt.sugar, err)
		}
		gotExplicit, err := mustCoreParser(t).parseInlines(tt.explicit, ast.Position{Line: 1, Column: 1})
		if err != nil {
			t.Fatalf("parse explicit %q: %v", tt.explicit, err)
		}
		if reflect.TypeOf(gotSugar[0]) != reflect.TypeOf(gotExplicit[0]) {
			t.Errorf("%q type = %T, explicit type = %T", tt.sugar, gotSugar[0], gotExplicit[0])
		}
		if sugarSemanticValue(gotSugar[0]) != sugarSemanticValue(gotExplicit[0]) {
			t.Errorf("%q semantics = %q, explicit semantics = %q", tt.sugar, sugarSemanticValue(gotSugar[0]), sugarSemanticValue(gotExplicit[0]))
		}
	}
}

func TestInlineSugarAcceptsConstrainedBoundaries(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantType ast.Inline
	}{
		{name: "line boundaries", input: "*important*", wantType: &ast.Emphasis{}},
		{name: "ASCII spaces", input: "This is **important** text", wantType: &ast.Strong{}},
		{name: "ASCII punctuation", input: "Use (`go test`).", wantType: &ast.CodeSpan{}},
		{name: "Japanese brackets", input: "これは「__重要__」です。", wantType: &ast.Bold{}},
		{name: "Japanese parentheses", input: "これは（++重要++）。", wantType: &ast.Underline{}},
		{name: "Japanese period", input: "前。~~削除~~。後", wantType: &ast.Strikethrough{}},
		{name: "link punctuation", input: "See [Orische](https://example.com).", wantType: &ast.Link{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := mustCoreParser(t).parseInlines(tt.input, ast.Position{Line: 1, Column: 1})
			if err != nil {
				t.Fatalf("parseInlines returned an error: %v", err)
			}
			if !containsSugarType(got, reflect.TypeOf(tt.wantType)) {
				t.Errorf("parsed nodes %#v do not contain %T", got, tt.wantType)
			}
		})
	}
}

func TestInlineSugarRejectsInvalidBoundariesAndDelimiterRuns(t *testing.T) {
	tests := []string{
		"foo**bar**baz",
		"これは**重要**です",
		"Use`go test`now",
		"a\t*x* b",
		"a　*x* b",
		"a\u00a0*x* b",
		"C++ language",
		"foo_bar_baz",
		"***important***",
		"___important___",
		"~~~important~~~",
		"``code``",
		"*x**",
		"**x*",
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			assertOnlyLiteralText(t, input)
		})
	}
}

func TestInlineSugarFallsBackForInvalidOrIncompleteCandidates(t *testing.T) {
	tests := []string{
		"**",
		"* text*",
		"*text *",
		"*unterminated **valid**",
		"* :[em]{nested} *",
		"[](https://example.com)",
		"[x]()",
		"[ x](https://example.com)",
		"[x ](https://example.com)",
		"[x]( https://example.com)",
		"[x](https://example.com )",
		"[x](https://example.com",
		"[x](https://example.com/a(b)c)",
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			assertOnlyLiteralText(t, input)
		})
	}
}

func TestInlineSugarHonorsEscapedDelimiters(t *testing.T) {
	for _, input := range []string{`\*text*`, `*text\*`, `\[x](https://example.com)`} {
		got, err := mustCoreParser(t).parseInlines(input, ast.Position{Line: 1, Column: 1})
		if err != nil {
			t.Fatalf("parseInlines(%q) returned an error: %v", input, err)
		}
		for _, node := range got {
			if _, ok := node.(*ast.Text); !ok {
				t.Errorf("parseInlines(%q) produced %T, want only Text", input, node)
			}
		}
	}
}

func TestInlineSugarRecursivelyParsesDifferentSyntaxes(t *testing.T) {
	input := "**outer _inner_ :[em]{explicit}**"
	got, err := mustCoreParser(t).parseInlines(input, ast.Position{Line: 1, Column: 1})
	if err != nil {
		t.Fatalf("parseInlines returned an error: %v", err)
	}
	strong, ok := got[0].(*ast.Strong)
	if !ok {
		t.Fatalf("inline type = %T, want *ast.Strong", got[0])
	}
	if !containsSugarType(strong.Content, reflect.TypeOf(&ast.Italic{})) ||
		!containsSugarType(strong.Content, reflect.TypeOf(&ast.Emphasis{})) {
		t.Errorf("nested content = %#v, want Italic and Emphasis", strong.Content)
	}
}

func TestCodeSpanSugarKeepsContentLiteral(t *testing.T) {
	input := "`:[em]{x} \\* **y**`"
	got, err := mustCoreParser(t).parseInlines(input, ast.Position{Line: 1, Column: 1})
	if err != nil {
		t.Fatalf("parseInlines returned an error: %v", err)
	}
	code, ok := got[0].(*ast.CodeSpan)
	if !ok || code.Value != `:[em]{x} \* **y**` {
		t.Fatalf("inline = %#v, want literal CodeSpan", got[0])
	}
}

func TestLinkSugarParsesLabelAndUnescapesURI(t *testing.T) {
	input := `[**label**](https://example.com/a\)b)`
	got, err := mustCoreParser(t).parseInlines(input, ast.Position{Line: 1, Column: 1})
	if err != nil {
		t.Fatalf("parseInlines returned an error: %v", err)
	}
	link, ok := got[0].(*ast.Link)
	if !ok {
		t.Fatalf("inline type = %T, want *ast.Link", got[0])
	}
	if link.URI != "https://example.com/a)b" {
		t.Errorf("Link.URI = %q, want %q", link.URI, "https://example.com/a)b")
	}
	if !containsSugarType(link.Content, reflect.TypeOf(&ast.Strong{})) {
		t.Errorf("Link.Content = %#v, want nested Strong", link.Content)
	}
}

func TestInlineSugarDoesNotCrossLogicalLines(t *testing.T) {
	for _, newline := range []string{"\n", "\r\n", "\r"} {
		input := "*first" + newline + "second*"
		got, err := mustCoreParser(t).parseInlines(input, ast.Position{Line: 1, Column: 1})
		if err != nil {
			t.Fatalf("parseInlines returned an error: %v", err)
		}
		for _, node := range got {
			if _, ok := node.(*ast.Text); !ok {
				t.Errorf("newline %q produced %T, want only Text", newline, node)
			}
		}
	}
}

func TestUnterminatedInlineSugarResumesOnNextLogicalLine(t *testing.T) {
	for _, newline := range []string{"\n", "\r\n", "\r"} {
		input := "*unterminated **same line**" + newline + "**next line**"
		got, err := mustCoreParser(t).parseInlines(input, ast.Position{Line: 1, Column: 1})
		if err != nil {
			t.Fatalf("parseInlines returned an error: %v", err)
		}
		if len(got) != 2 {
			t.Fatalf("newline %q inline count = %d, want Text and Strong", newline, len(got))
		}
		if _, ok := got[0].(*ast.Text); !ok {
			t.Errorf("newline %q first inline = %T, want *ast.Text", newline, got[0])
		}
		if _, ok := got[1].(*ast.Strong); !ok {
			t.Errorf("newline %q second inline = %T, want *ast.Strong", newline, got[1])
		}
	}
}

func TestInlineSugarConnectsToInlineCapableBlocks(t *testing.T) {
	doc, err := Parse("= **heading**\n\n* _item_\n\nparagraph ++under++")
	if err != nil {
		t.Fatalf("Parse returned an error: %v", err)
	}
	if len(doc.Blocks) != 3 {
		t.Fatalf("block count = %d, want 3", len(doc.Blocks))
	}

	heading, ok := doc.Blocks[0].(*ast.Heading)
	if !ok || !containsSugarType(heading.Content, reflect.TypeOf(&ast.Strong{})) {
		t.Errorf("heading = %#v, want Strong content", doc.Blocks[0])
	}
	list, ok := doc.Blocks[1].(*ast.List)
	if !ok {
		t.Fatalf("second block = %T, want *ast.List", doc.Blocks[1])
	}
	item := list.Items[0].Blocks[0].(*ast.Paragraph)
	if !containsSugarType(item.Content, reflect.TypeOf(&ast.Italic{})) {
		t.Errorf("list item = %#v, want Italic content", item)
	}
	paragraph, ok := doc.Blocks[2].(*ast.Paragraph)
	if !ok || !containsSugarType(paragraph.Content, reflect.TypeOf(&ast.Underline{})) {
		t.Errorf("paragraph = %#v, want Underline content", doc.Blocks[2])
	}
}

func assertSugarTextContent(t *testing.T, content []ast.Inline, origin ast.Position, offset int) {
	t.Helper()
	if len(content) != 1 {
		t.Fatalf("content count = %d, want 1", len(content))
	}
	text, ok := content[0].(*ast.Text)
	if !ok || text.Value != "x" {
		t.Fatalf("content = %#v, want Text value x", content)
	}
	want := ast.Range{
		Start: ast.Position{Line: origin.Line, Column: origin.Column + offset},
		End:   ast.Position{Line: origin.Line, Column: origin.Column + offset},
	}
	if text.Range != want {
		t.Errorf("content Range = %#v, want %#v", text.Range, want)
	}
}

func assertOnlyLiteralText(t *testing.T, input string) {
	t.Helper()
	got, err := mustCoreParser(t).parseInlines(input, ast.Position{Line: 1, Column: 1})
	if err != nil {
		t.Fatalf("parseInlines returned an error: %v", err)
	}
	var value strings.Builder
	for _, node := range got {
		text, ok := node.(*ast.Text)
		if !ok {
			t.Fatalf("inline type = %T, want only *ast.Text", node)
		}
		value.WriteString(text.Value)
	}
	if value.String() != input {
		t.Errorf("literal text = %q, want %q", value.String(), input)
	}
}

func containsSugarType(nodes []ast.Inline, want reflect.Type) bool {
	for _, node := range nodes {
		if reflect.TypeOf(node) == want {
			return true
		}
	}
	return false
}

func sugarNodeRange(node ast.Inline) ast.Range {
	switch node := node.(type) {
	case *ast.Emphasis:
		return node.Range
	case *ast.Strong:
		return node.Range
	case *ast.Italic:
		return node.Range
	case *ast.Bold:
		return node.Range
	case *ast.Underline:
		return node.Range
	case *ast.Strikethrough:
		return node.Range
	case *ast.CodeSpan:
		return node.Range
	case *ast.Link:
		return node.Range
	default:
		return ast.Range{}
	}
}

func sugarNodeContent(node ast.Inline) []ast.Inline {
	switch node := node.(type) {
	case *ast.Emphasis:
		return node.Content
	case *ast.Strong:
		return node.Content
	case *ast.Italic:
		return node.Content
	case *ast.Bold:
		return node.Content
	case *ast.Underline:
		return node.Content
	case *ast.Strikethrough:
		return node.Content
	default:
		return nil
	}
}

func sugarSemanticValue(node ast.Inline) string {
	switch node := node.(type) {
	case *ast.CodeSpan:
		return node.Value
	case *ast.Link:
		return node.URI + "|" + inlineTextValue(node.Content)
	default:
		return inlineTextValue(sugarNodeContent(node))
	}
}

func inlineTextValue(nodes []ast.Inline) string {
	var value strings.Builder
	for _, node := range nodes {
		if text, ok := node.(*ast.Text); ok {
			value.WriteString(text.Value)
		}
	}
	return value.String()
}
