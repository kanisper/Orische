package feature

// Language is an immutable-by-convention declaration set. Blocks preserve
// Sugar-reader precedence. The parser frontend validates the definitions and
// copies their registries into a private compiled specification.
type Language struct {
	Paragraph ParagraphDefinition
	Blocks    []BlockDefinition
	Inlines   []InlineDirectiveDefinition
}
