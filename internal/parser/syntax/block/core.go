package block

import "orische/internal/parser/feature"

// Definitions returns built-in block definitions in sugar-reader precedence.
// Directive-only definitions may appear anywhere in the slice.
func Definitions() []feature.BlockDefinition {
	return []feature.BlockDefinition{
		&codeDefinition{},
		&headingDefinition{},
		&listDefinition{},
	}
}
