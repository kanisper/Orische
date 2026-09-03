package lsp

import (
	"context"
	"fmt"

	"go.lsp.dev/protocol"

	"orische/internal/completion"
)

func (s *server) Completion(
	_ context.Context,
	params *protocol.CompletionParams,
) (protocol.CompletionResult, error) {
	if err := s.requestAllowed(); err != nil {
		return nil, err
	}
	if params == nil {
		return protocol.CompletionItemSlice{}, nil
	}
	document, ok := s.documents.get(params.TextDocument.URI)
	if !ok {
		return protocol.CompletionItemSlice{}, nil
	}

	cursor, err := document.mapper.byteOffset(params.Position)
	if err != nil {
		return nil, fmt.Errorf("completion position: %w", err)
	}
	candidates := completion.Complete(completion.Request{
		Source:       document.Source,
		CursorOffset: cursor,
		AST:          document.analysis.AST,
	})
	items := make(protocol.CompletionItemSlice, 0, len(candidates))
	for _, candidate := range candidates {
		start, err := document.mapper.position(candidate.Replace.Start)
		if err != nil {
			return nil, fmt.Errorf("completion edit start: %w", err)
		}
		end, err := document.mapper.position(candidate.Replace.End)
		if err != nil {
			return nil, fmt.Errorf("completion edit end: %w", err)
		}

		insertTextFormat := protocol.InsertTextFormatPlainText
		if candidate.InsertFormat == completion.InsertFormatSnippet {
			insertTextFormat = protocol.InsertTextFormatSnippet
		}
		items = append(items, protocol.CompletionItem{
			Label:            candidate.Label,
			Kind:             protocol.CompletionItemKindKeyword,
			InsertTextFormat: insertTextFormat,
			TextEdit: &protocol.TextEdit{
				Range:   protocol.Range{Start: start, End: end},
				NewText: candidate.InsertText,
			},
		})
	}

	if !s.documents.isCurrent(document) {
		return protocol.CompletionItemSlice{}, nil
	}
	return items, nil
}
