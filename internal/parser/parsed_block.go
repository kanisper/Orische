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

func (*parsedBlock) isParsedBlockNode() {}

// TODO: Define parsedList struct as parsedBlockNode

// TODO: Define parsedListItem struct, which have Blocks as parsedBlockNode
