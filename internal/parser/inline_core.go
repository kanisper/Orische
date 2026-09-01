package parser

func coreInlineDefinitions() map[string]inlineDefinition {
	return map[string]inlineDefinition{
		"em":   emphasisDefinition,
		"link": linkDefinition,
		"code": codeDefinition,
	}
}
