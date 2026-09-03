package lsp

import (
	"testing"

	"go.lsp.dev/protocol"
	"go.lsp.dev/uri"
)

func TestServerCompletionUsesCurrentSource(t *testing.T) {
	srv := newServer()
	documentURI := uri.URI("file:///document.oris")
	if err := srv.DidOpen(t.Context(), &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI: documentURI, Version: 1, Text: "日本🍣 :[str",
		},
	}); err != nil {
		t.Fatal(err)
	}

	result, err := srv.Completion(t.Context(), &protocol.CompletionParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{URI: documentURI},
			Position:     protocol.Position{Line: 0, Character: 10},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	items, ok := result.(protocol.CompletionItemSlice)
	if !ok || len(items) != 1 {
		t.Fatalf("completion result = %#v, want one item", result)
	}
	if items[0].Label != "strong" || items[0].InsertTextFormat != protocol.InsertTextFormatPlainText {
		t.Errorf("completion item = %#v", items[0])
	}
	edit, ok := items[0].TextEdit.(*protocol.TextEdit)
	if !ok {
		t.Fatalf("text edit = %#v, want *protocol.TextEdit", items[0].TextEdit)
	}
	wantRange := protocol.Range{
		Start: protocol.Position{Line: 0, Character: 7},
		End:   protocol.Position{Line: 0, Character: 10},
	}
	if edit.Range != wantRange || edit.NewText != "strong" {
		t.Errorf("text edit = %#v, want range %v and text strong", edit, wantRange)
	}

	changeDocument(t, srv, documentURI, 2, ":::[hea")
	result, err = srv.Completion(t.Context(), &protocol.CompletionParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{URI: documentURI},
			Position:     protocol.Position{Line: 0, Character: 7},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	items = result.(protocol.CompletionItemSlice)
	if len(items) != 1 || items[0].Label != "heading" {
		t.Errorf("completion after change = %#v, want heading", items)
	}

	if err := srv.DidClose(t.Context(), &protocol.DidCloseTextDocumentParams{
		TextDocument: protocol.TextDocumentIdentifier{URI: documentURI},
	}); err != nil {
		t.Fatal(err)
	}
	result, err = srv.Completion(t.Context(), &protocol.CompletionParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{URI: documentURI},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if items := result.(protocol.CompletionItemSlice); len(items) != 0 {
		t.Errorf("completion after close = %#v, want empty", items)
	}
}

func TestServerCompletionUsesNegotiatedUTF8(t *testing.T) {
	srv := newServer()
	if _, err := srv.Initialize(t.Context(), &protocol.InitializeParams{
		Capabilities: protocol.ClientCapabilities{General: &protocol.GeneralClientCapabilities{
			PositionEncodings: []protocol.PositionEncodingKind{protocol.PositionEncodingKindUTF8},
		}},
	}); err != nil {
		t.Fatal(err)
	}
	documentURI := uri.URI("file:///utf8.oris")
	source := "日 :[str"
	if err := srv.DidOpen(t.Context(), &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{URI: documentURI, Version: 1, Text: source},
	}); err != nil {
		t.Fatal(err)
	}
	result, err := srv.Completion(t.Context(), &protocol.CompletionParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{URI: documentURI},
			Position:     protocol.Position{Character: uint32(len(source))},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	items := result.(protocol.CompletionItemSlice)
	if len(items) != 1 {
		t.Fatalf("completion = %#v, want one item", items)
	}
	edit := items[0].TextEdit.(*protocol.TextEdit)
	want := protocol.Range{
		Start: protocol.Position{Character: uint32(len("日 :["))},
		End:   protocol.Position{Character: uint32(len(source))},
	}
	if edit.Range != want {
		t.Errorf("UTF-8 edit range = %v, want %v", edit.Range, want)
	}
}

func TestServerCompletionRejectsInvalidPosition(t *testing.T) {
	srv := newServer()
	documentURI := uri.URI("file:///document.oris")
	if err := srv.DidOpen(t.Context(), &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI: documentURI, Version: 1, Text: ":[str",
		},
	}); err != nil {
		t.Fatal(err)
	}
	_, err := srv.Completion(t.Context(), &protocol.CompletionParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{URI: documentURI},
			Position:     protocol.Position{Line: 4},
		},
	})
	if err == nil {
		t.Fatal("completion at invalid position succeeded")
	}
}
