package ast

type Range struct {
	StartLine int
	EndLine   int
}

type Root interface {
	isRoot()
}

type Block interface {
	isBlock()
}

type Inline interface {
	isInline()
}

type Document struct {
	Blocks []Block
	Range  Range
}

type Heading struct {
	Level   int
	Content []Inline
	Range   Range
}

type Paragraph struct {
	Content []Inline
	Range   Range
}

type List struct {
	Ordered bool
	Items   []*ListItem
	Range   Range
}

type ListItem struct {
	Blocks []Block
	Range  Range
}

type CodeBlock struct {
	Language string
	Text     string
	Range    Range
}

type Text struct {
	Value string
}

type Emphasis struct {
	Content []Inline
}

type CodeSpan struct {
	Value string
}

type Link struct {
	URI     string
	Content []Inline
}

func (*Document) isRoot() {}

func (*Heading) isBlock()   {}
func (*Paragraph) isBlock() {}
func (*List) isBlock()      {}
func (*CodeBlock) isBlock() {}

func (*Text) isInline()     {}
func (*Emphasis) isInline() {}
func (*CodeSpan) isInline() {}
func (*Link) isInline()     {}
