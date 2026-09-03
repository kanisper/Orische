package completion

import (
	"strings"
	"unicode/utf8"

	"orische/internal/ast"
)

type Request struct {
	Source       string
	CursorOffset int
	AST          *ast.Document
}

type Span struct {
	Start int
	End   int
}

type InsertFormat uint8

const (
	InsertFormatPlainText InsertFormat = iota
	InsertFormatSnippet
)

type Candidate struct {
	Label        string
	Replace      Span
	InsertText   string
	InsertFormat InsertFormat
}

var (
	blockDirectiveTypes  = []string{"heading", "paragraph", "code"}
	inlineDirectiveTypes = []string{"em", "strong", "italic", "bold", "del", "outdated", "link", "code"}
)

// Complete derives directive type edits from source around the cursor. The AST
// supplies optional context for regions where directive syntax is literal.
func Complete(request Request) []Candidate {
	if request.CursorOffset < 0 || request.CursorOffset > len(request.Source) ||
		!utf8.ValidString(request.Source) ||
		(request.CursorOffset < len(request.Source) &&
			request.CursorOffset > 0 && !utf8.RuneStart(request.Source[request.CursorOffset])) {
		return nil
	}

	lineStart := currentLineStart(request.Source, request.CursorOffset)
	if insideCodeBlock(request.AST, sourceLineNumber(request.Source, lineStart)) ||
		insideCodeSpan(request.AST, sourcePosition(request.Source, request.CursorOffset)) {
		return nil
	}
	prefixStart := request.CursorOffset
	for prefixStart > lineStart && isDirectiveTypeByte(request.Source[prefixStart-1]) {
		prefixStart--
	}
	prefix := request.Source[prefixStart:request.CursorOffset]

	var types []string
	if prefixStart == lineStart+len(":::[") && request.Source[lineStart:prefixStart] == ":::[" {
		types = blockDirectiveTypes
	} else if prefixStart >= lineStart+len(":[") && request.Source[prefixStart-2:prefixStart] == ":[" {
		colon := prefixStart - 2
		if (colon >= 2 && request.Source[colon-2:colon] == "::") || isEscaped(request.Source, colon) {
			return nil
		}
		types = inlineDirectiveTypes
	} else {
		return nil
	}

	normalizedPrefix := strings.ToLower(prefix)
	candidates := make([]Candidate, 0, len(types))
	for _, directiveType := range types {
		if !strings.HasPrefix(directiveType, normalizedPrefix) {
			continue
		}
		candidates = append(candidates, Candidate{
			Label:        directiveType,
			Replace:      Span{Start: prefixStart, End: request.CursorOffset},
			InsertText:   directiveType,
			InsertFormat: InsertFormatPlainText,
		})
	}
	return candidates
}

func currentLineStart(source string, cursor int) int {
	for cursor > 0 {
		if source[cursor-1] == '\n' || source[cursor-1] == '\r' {
			break
		}
		cursor--
	}
	return cursor
}

func isDirectiveTypeByte(b byte) bool {
	return b >= 'a' && b <= 'z' || b >= 'A' && b <= 'Z'
}

func isEscaped(source string, offset int) bool {
	backslashes := 0
	for offset > 0 && source[offset-1] == '\\' {
		backslashes++
		offset--
	}
	return backslashes%2 != 0
}

func sourceLineNumber(source string, offset int) int {
	line := 1
	for cursor := 0; cursor < offset; {
		switch source[cursor] {
		case '\r':
			line++
			cursor++
			if cursor < offset && source[cursor] == '\n' {
				cursor++
			}
		case '\n':
			line++
			cursor++
		default:
			_, size := utf8.DecodeRuneInString(source[cursor:offset])
			cursor += size
		}
	}
	return line
}

func insideCodeBlock(document *ast.Document, line int) bool {
	if document == nil {
		return false
	}
	return blocksContainCodeLine(document.Blocks, line)
}

func blocksContainCodeLine(blocks []ast.Block, line int) bool {
	for _, block := range blocks {
		switch block := block.(type) {
		case *ast.CodeBlock:
			if line > block.Range.Start.Line && line < block.Range.End.Line {
				return true
			}
		case *ast.List:
			for _, item := range block.Items {
				if blocksContainCodeLine(item.Blocks, line) {
					return true
				}
			}
		}
	}
	return false
}

func sourcePosition(source string, offset int) ast.Position {
	position := ast.Position{Line: 1, Column: 1}
	for cursor := 0; cursor < offset; {
		switch source[cursor] {
		case '\r':
			position.Line++
			position.Column = 1
			cursor++
			if cursor < offset && source[cursor] == '\n' {
				cursor++
			}
		case '\n':
			position.Line++
			position.Column = 1
			cursor++
		default:
			_, size := utf8.DecodeRuneInString(source[cursor:offset])
			position.Column++
			cursor += size
		}
	}
	return position
}

func insideCodeSpan(document *ast.Document, position ast.Position) bool {
	if document == nil {
		return false
	}
	return blocksContainCodePosition(document.Blocks, position)
}

func blocksContainCodePosition(blocks []ast.Block, position ast.Position) bool {
	for _, block := range blocks {
		switch block := block.(type) {
		case *ast.Heading:
			if inlinesContainCodePosition(block.Content, position) {
				return true
			}
		case *ast.Paragraph:
			if inlinesContainCodePosition(block.Content, position) {
				return true
			}
		case *ast.List:
			for _, item := range block.Items {
				if item != nil && blocksContainCodePosition(item.Blocks, position) {
					return true
				}
			}
		}
	}
	return false
}

func inlinesContainCodePosition(inlines []ast.Inline, position ast.Position) bool {
	for _, inline := range inlines {
		switch inline := inline.(type) {
		case *ast.CodeSpan:
			if positionAfter(position, inline.Range.Start) && !positionAfter(position, inline.Range.End) {
				return true
			}
		case *ast.Emphasis:
			if inlinesContainCodePosition(inline.Content, position) {
				return true
			}
		case *ast.Strong:
			if inlinesContainCodePosition(inline.Content, position) {
				return true
			}
		case *ast.Italic:
			if inlinesContainCodePosition(inline.Content, position) {
				return true
			}
		case *ast.Bold:
			if inlinesContainCodePosition(inline.Content, position) {
				return true
			}
		case *ast.Deleted:
			if inlinesContainCodePosition(inline.Content, position) {
				return true
			}
		case *ast.Outdated:
			if inlinesContainCodePosition(inline.Content, position) {
				return true
			}
		case *ast.Link:
			if inlinesContainCodePosition(inline.Content, position) {
				return true
			}
		}
	}
	return false
}

func positionAfter(left, right ast.Position) bool {
	return left.Line > right.Line || left.Line == right.Line && left.Column > right.Column
}
