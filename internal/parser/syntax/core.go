// Package syntax assembles the built-in Orische language definitions.
package syntax

import (
	"orische/internal/parser/feature"
	blocksyntax "orische/internal/parser/syntax/block"
	inlinesyntax "orische/internal/parser/syntax/inline"
)

func Core() feature.Language {
	return feature.Language{
		Blocks:  blocksyntax.Definitions(),
		Inlines: inlinesyntax.Definitions(),
	}
}
