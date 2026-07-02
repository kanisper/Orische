package parser

type blockContext struct {
	lines []string
	pos   int
}

func newBlockContext(lines []string, start int, parser *Parser) *blockContext {
	return &blockContext{
		lines: lines,
		pos:   start,
	}
}

func (c *blockContext) isEOF() bool {
	return c.pos >= len(c.lines)
}

func (c *blockContext) getLine() string {
	return c.lines[c.pos]
}

func (c *blockContext) advance(n int) {
	c.pos += n
}

func (c *blockContext) setPos(pos int) {
	c.pos = pos
}

func (c *blockContext) getPos() int {
	return c.pos
}
