package parser

func coreInlineDefinitions() map[string]inlineDefinition {
	return map[string]inlineDefinition{
		"em":       emphasisDefinition,
		"strong":   strongDefinition,
		"italic":   italicDefinition,
		"bold":     boldDefinition,
		"del":      deletedDefinition,
		"outdated": outdatedDefinition,
		"link":     linkDefinition,
		"code":     codeDefinition,
	}
}
