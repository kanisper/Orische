package feature

// Language is an immutable-by-convention declaration set for replaceable
// syntax. Blocks preserve Sugar-reader precedence. The parser frontend
// validates the definitions and copies their registries into a private compiled
// specification.
type Language struct {
	Blocks  []BlockDefinition
	Inlines []InlineDirectiveDefinition
}
