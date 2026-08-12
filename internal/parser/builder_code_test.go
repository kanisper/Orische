package parser

import (
	"testing"

	"orische/internal/ast"

	"github.com/google/go-cmp/cmp"
)

func TestBuildCodeBlock(t *testing.T) {
	input := &parsedBlock{
		Type: "code",
		Attr: "c",
		Text: `#include <stdio.h>

int main()
{
	printf("Hello, world!");
}`,
		Range: ast.Range{
			Start: ast.Position{Line: 1, Column: 1},
			End:   ast.Position{Line: 6, Column: 1},
		},
	}

	want := &ast.CodeBlock{
		Language: "c",
		Text: `#include <stdio.h>

int main()
{
	printf("Hello, world!");
}`,
		Range: ast.Range{
			Start: ast.Position{Line: 1, Column: 1},
			End:   ast.Position{Line: 6, Column: 1},
		},
	}

	got, err := (&codeBlockBuilder{}).build(input)
	if err != nil {
		t.Fatalf("build returned an error: %v", err)
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("build incorrectly.\n(-want +got)\n%s", diff)
	}
}

func TestBuildCodeBlock_RejectsWrongNodeType(t *testing.T) {
	got, err := (&codeBlockBuilder{}).build(&parsedList{})
	if err == nil {
		t.Fatal("build accepted an unexpected node type")
	}
	if got != nil {
		t.Errorf("build returned a block: %v", got)
	}
}
