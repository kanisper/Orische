package parser

func coreInlineDefinitions() map[string]inlineDefinition {
	return map[string]inlineDefinition{
		"em":        emphasisDefinition,
		"strong":    strongDefinition,
		"italic":    italicDefinition,
		"bold":      boldDefinition,
		"underline": underlineDefinition,
		"strike":    strikethroughDefinition,
		"link":      linkDefinition,
		"code":      codeDefinition,
	}
}
