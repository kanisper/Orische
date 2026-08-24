package parser

import "orische/internal/ast"

type parsedDocument struct {
	Blocks []parsedBlockNode
	Range  ast.Range
}

type parsedBlockNode interface {
	isParsedBlockNode()
	getBuilderKey() string
	getBlockRange() ast.Range
}

type parsedBlock struct {
	Type  string
	Attr  string
	Text  string
	Range ast.Range
}

func (pb *parsedBlock) getBuilderKey() string {
	return pb.Type
}

func (pb *parsedBlock) getBlockRange() ast.Range {
	return pb.Range
}

type parsedList struct {
	Ordered bool
	Items   []parsedListItem
	Range   ast.Range
}

func (pl *parsedList) getBuilderKey() string {
	return blockBuilderKeyList
}

func (pl *parsedList) getBlockRange() ast.Range {
	return pl.Range
}

type parsedListItem struct {
	Blocks []parsedBlockNode
	Range  ast.Range
}

func (*parsedBlock) isParsedBlockNode() {}
func (*parsedList) isParsedBlockNode()  {}
