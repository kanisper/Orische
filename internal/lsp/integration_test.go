package lsp

import (
	"context"
	"testing"
	"time"

	"go.lsp.dev/jsonrpc2"
	"go.lsp.dev/protocol"
	"go.lsp.dev/uri"
)

func TestProtocolWorkflow(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()

	clientStream, serverStream := jsonrpc2.NewChannelStreamPair(16)
	serveErr := make(chan error, 1)
	go func() { serveErr <- serveStream(ctx, serverStream) }()

	recorder := &diagnosticClient{published: make(chan protocol.PublishDiagnosticsParams, 4)}
	_, clientConn, server := protocol.NewClient(ctx, recorder, clientStream)
	defer func() { _ = clientConn.Close() }()

	initialized, err := server.Initialize(ctx, &protocol.InitializeParams{})
	if err != nil {
		t.Fatalf("initialize: %v", err)
	}
	if initialized.Capabilities.CompletionProvider == nil {
		t.Fatal("initialize did not advertise completion")
	}
	if err := server.Initialized(ctx, &protocol.InitializedParams{}); err != nil {
		t.Fatalf("initialized: %v", err)
	}

	documentURI := uri.URI("file:///workflow.oris")
	if err := server.DidOpen(ctx, &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI: documentURI, Version: 1,
			Text: ":::[unsupported]\ntext\n:::",
		},
	}); err != nil {
		t.Fatalf("didOpen: %v", err)
	}
	if diagnostics := receiveDiagnostics(t, ctx, recorder.published); len(diagnostics.Diagnostics) != 1 {
		t.Fatalf("didOpen diagnostics = %#v, want one", diagnostics.Diagnostics)
	}

	source := "日本🍣 :[str"
	if err := server.DidChange(ctx, &protocol.DidChangeTextDocumentParams{
		TextDocument: protocol.VersionedTextDocumentIdentifier{
			TextDocumentIdentifier: protocol.TextDocumentIdentifier{URI: documentURI},
			Version:                2,
		},
		ContentChanges: []protocol.TextDocumentContentChangeEvent{
			&protocol.TextDocumentContentChangeWholeDocument{Text: source},
		},
	}); err != nil {
		t.Fatalf("didChange: %v", err)
	}
	if diagnostics := receiveDiagnostics(t, ctx, recorder.published); len(diagnostics.Diagnostics) != 0 {
		t.Fatalf("didChange diagnostics = %#v, want empty", diagnostics.Diagnostics)
	}

	result, err := server.Completion(ctx, &protocol.CompletionParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{URI: documentURI},
			Position:     protocol.Position{Line: 0, Character: 10},
		},
	})
	if err != nil {
		t.Fatalf("completion: %v", err)
	}
	items, ok := result.(protocol.CompletionItemSlice)
	if !ok || len(items) != 1 || items[0].Label != "strong" {
		t.Fatalf("completion = %#v, want strong", result)
	}
	edit, ok := items[0].TextEdit.(*protocol.TextEdit)
	if !ok {
		t.Fatalf("completion edit = %#v, want TextEdit", items[0].TextEdit)
	}
	wantRange := protocol.Range{
		Start: protocol.Position{Line: 0, Character: 7},
		End:   protocol.Position{Line: 0, Character: 10},
	}
	if edit.Range != wantRange || edit.NewText != "strong" {
		t.Errorf("completion edit = %#v, want range %v and text strong", edit, wantRange)
	}

	if err := server.DidClose(ctx, &protocol.DidCloseTextDocumentParams{
		TextDocument: protocol.TextDocumentIdentifier{URI: documentURI},
	}); err != nil {
		t.Fatalf("didClose: %v", err)
	}
	if diagnostics := receiveDiagnostics(t, ctx, recorder.published); len(diagnostics.Diagnostics) != 0 {
		t.Fatalf("didClose diagnostics = %#v, want empty", diagnostics.Diagnostics)
	}

	if err := server.Shutdown(ctx); err != nil {
		t.Fatalf("shutdown: %v", err)
	}
	if err := server.Exit(ctx); err != nil {
		t.Fatalf("exit: %v", err)
	}
	select {
	case err := <-serveErr:
		if err != nil {
			t.Fatalf("serve: %v", err)
		}
	case <-ctx.Done():
		t.Fatal("server did not stop")
	}
}
