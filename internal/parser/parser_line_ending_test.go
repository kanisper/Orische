package parser

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParseTreatsLineEndingsAsLogicalNewlines(t *testing.T) {
	lf := "= Heading\n\nparagraph line 1 :[em]{inline\nacross} tail\nparagraph line 2\n\n* item one\n# item two\n\n:::[code:go]\nline one\nline two\n:::\n"
	tests := []struct {
		name  string
		input string
	}{
		{name: "LF", input: lf},
		{name: "CRLF", input: replaceLineEndings(lf, "\r\n")},
		{name: "CR", input: replaceLineEndings(lf, "\r")},
		{name: "mixed", input: "= Heading\r\n\nparagraph line 1 :[em]{inline\racross} tail\nparagraph line 2\n\r\n* item one\n# item two\r\n\r:::[code:go]\rline one\nline two\r\n:::\r"},
	}

	want, err := Parse(lf)
	if err != nil {
		t.Fatalf("Parse(LF) returned an error: %v", err)
	}

	for _, tt := range tests[1:] {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.input)
			if err != nil {
				t.Fatalf("Parse(%s) returned an error: %v", tt.name, err)
			}
			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("Parse(%s) differs from LF input (-want +got):\n%s", tt.name, diff)
			}
		})
	}
}

func replaceLineEndings(input, lineEnding string) string {
	var result []byte
	for i := 0; i < len(input); i++ {
		if input[i] == '\n' {
			result = append(result, lineEnding...)
			continue
		}
		result = append(result, input[i])
	}
	return string(result)
}
