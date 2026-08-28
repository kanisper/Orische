package inline

import "orische/internal/parser/feature"

func Definitions() []feature.InlineDirectiveDefinition {
	return []feature.InlineDirectiveDefinition{
		&emphasisDefinition{},
		&linkDefinition{},
		&codeDefinition{},
	}
}
