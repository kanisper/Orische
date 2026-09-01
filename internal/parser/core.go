package parser

func newParser() *Parser {
	return &Parser{spec: newSpec()}
}
