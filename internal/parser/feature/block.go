// Package feature defines the syntax-neutral contracts shared by the parser
// frontend and syntax implementations.
package feature

import "orische/internal/ast"

// ParagraphBlockType is reserved for the fixed Paragraph fallback.
const ParagraphBlockType = "paragraph"

// BlockNode is the parsed-block IR transported by the parser frontend.
// Implementations may remain private to their syntax package.
type BlockNode interface {
	BlockType() string
	BlockRange() ast.Range
}

// TextBlock is the common IR for text-backed blocks such as paragraphs,
// headings, and block directives. ContentOrigin locates Text independently of
// the complete block Range.
type TextBlock struct {
	Type          string
	Attr          string
	Text          string
	ContentOrigin ast.Position
	Range         ast.Range
}

func (b *TextBlock) BlockType() string {
	return b.Type
}

func (b *TextBlock) BlockRange() ast.Range {
	return b.Range
}

// BlockLine is a source line exposed to a block reader.
type BlockLine struct {
	Number int
	Text   string
}

// BlockInput is an immutable view of source lines starting at the current
// document position. Offsets are zero-based.
type BlockInput interface {
	Len() int
	Line(offset int) (BlockLine, bool)
}

// BlockReadResult reports whether a reader matched and how many lines the
// frontend should consume. A non-match has zero Consumed and a nil Node. A
// match has a non-nil Node and consumes between one and BlockInput.Len lines.
// The frontend validates these invariants.
type BlockReadResult struct {
	Matched  bool
	Consumed int
	Node     BlockNode
}

// BlockReader recognizes source without mutating frontend cursor state.
type BlockReader interface {
	ReadBlock(BlockInput) (BlockReadResult, error)
}

// BuildContext exposes only the recursive parser capabilities required by
// block syntax implementations. Its methods return a non-nil AST node whenever
// they return a nil error.
type BuildContext interface {
	ParseInlines(text string, origin ast.Position) ([]ast.Inline, error)
	BuildBlock(BlockNode) (ast.Block, error)
}

// BlockDefinition builds one registered Block Type.
type BlockDefinition interface {
	BlockType() string
	BuildBlock(BuildContext, BlockNode) (ast.Block, error)
}

// ParagraphDefinition builds the AST node for the fixed Paragraph fallback.
type ParagraphDefinition interface {
	BlockType() string
	BuildParagraph(BuildContext, BlockNode) (*ast.Paragraph, error)
}

// BlockSugarDefinition combines source recognition with AST construction.
type BlockSugarDefinition interface {
	BlockDefinition
	BlockReader
}
