package parser

// blockContext is the mutable line cursor shared by block readers. Readers
// either leave pos on their last consumed line or restore it when they reject.
type blockContext struct {
	lines []string
	pos   int
}

func (c *blockContext) isEOF() bool {
	return c.pos >= len(c.lines)
}

func (c *blockContext) line() string {
	return c.lines[c.pos]
}
