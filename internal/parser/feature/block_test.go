package feature_test

import (
	"testing"

	"orische/internal/ast"
	"orische/internal/parser/feature"
)

func TestTextBlockImplementsNeutralNodeContract(t *testing.T) {
	rng := ast.Range{Start: ast.Position{Line: 2, Column: 3}, End: ast.Position{Line: 4, Column: 5}}
	node := &feature.TextBlock{Type: "TyPe", ContentOrigin: ast.Position{Line: 3, Column: 1}, Range: rng}

	if got := node.BlockType(); got != "TyPe" {
		t.Errorf("BlockType() = %q, want TyPe", got)
	}
	if got := node.BlockRange(); got != rng {
		t.Errorf("BlockRange() = %#v, want %#v", got, rng)
	}
}

var _ feature.BlockNode = (*feature.TextBlock)(nil)
