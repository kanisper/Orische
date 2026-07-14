package parser

import (
	"testing"

	"medoc/internal/ast"

	"github.com/google/go-cmp/cmp"
)

func TestBuildCodeBlock(t *testing.T) {
	input := &parsedBlock{
		Type:  "code",
		Attr:  "c",
		Text:  "#include <stdio.h>\n\nint main()\n{\nprintf(\"Hello, world!\");\n}",
		Range: ast.Range{StartLine: 1, EndLine: 6},
	}

	want := &ast.CodeBlock{
		Language: "c",
		Text:     "#include <stdio.h>\n\nint main()\n{\nprintf(\"Hello, world!\");\n}",
		Range:    ast.Range{StartLine: 1, EndLine: 6},
	}

	got, err := (&codeBlockBuilder{}).build(input)
	if err != nil {
		t.Errorf("build failed: %v", err)
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("build incorrectly.\n(-want +got)\n%s", diff)
	}
}
