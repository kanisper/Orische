package parser

import "medoc/internal/ast"

type parsedDocument struct {
	Blocks []parsedBlockNode
	Range  ast.Range
}

type parsedBlockNode interface {
	isParsedBlockNode()
}

type parsedBlock struct {
	Type  string
	Attr  string
	Text  string
	Range ast.Range
}

type parsedList struct {
	Ordered bool
	Items   []parsedListItem
	Range   ast.Range
}

type parsedListItem struct {
	Blocks []parsedBlockNode
	Range  ast.Range
}

func (*parsedBlock) isParsedBlockNode() {}
func (*parsedList) isParsedBlockNode()  {}
