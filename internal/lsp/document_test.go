package lsp

import (
	"errors"
	"testing"

	"go.lsp.dev/protocol"
	"go.lsp.dev/uri"
)

func TestDocumentStoreLifecycle(t *testing.T) {
	store := newDocumentStore(protocol.PositionEncodingKindUTF16)
	documentURI := uri.URI("file:///document.oris")

	if err := store.open(protocol.TextDocumentItem{
		URI: documentURI, Version: 1, Text: "first",
	}); err != nil {
		t.Fatalf("open: %v", err)
	}
	if err := store.open(protocol.TextDocumentItem{
		URI: documentURI, Version: 2, Text: "duplicate",
	}); !errors.Is(err, errDocumentAlreadyOpen) {
		t.Fatalf("duplicate open error = %v, want %v", err, errDocumentAlreadyOpen)
	}
	assertDocument(t, store, documentURI, 1, "first")

	applied, err := store.change(
		protocol.VersionedTextDocumentIdentifier{
			TextDocumentIdentifier: protocol.TextDocumentIdentifier{URI: documentURI},
			Version:                2,
		},
		[]protocol.TextDocumentContentChangeEvent{
			&protocol.TextDocumentContentChangeWholeDocument{Text: "second"},
		},
	)
	if err != nil {
		t.Fatalf("change: %v", err)
	}
	if !applied {
		t.Fatal("change was not applied")
	}
	assertDocument(t, store, documentURI, 2, "second")

	for _, version := range []int32{2, 1} {
		applied, err = store.change(
			protocol.VersionedTextDocumentIdentifier{
				TextDocumentIdentifier: protocol.TextDocumentIdentifier{URI: documentURI},
				Version:                version,
			},
			[]protocol.TextDocumentContentChangeEvent{
				&protocol.TextDocumentContentChangeWholeDocument{Text: "stale"},
			},
		)
		if err != nil {
			t.Fatalf("stale version %d: %v", version, err)
		}
		if applied {
			t.Errorf("stale version %d was applied", version)
		}
	}
	assertDocument(t, store, documentURI, 2, "second")

	if !store.close(documentURI) {
		t.Fatal("close did not remove the document")
	}
	if _, ok := store.get(documentURI); ok {
		t.Fatal("document remains available after close")
	}
	if store.close(documentURI) {
		t.Fatal("second close reported a removal")
	}
}

func TestDocumentStoreRejectsInvalidChanges(t *testing.T) {
	store := newDocumentStore(protocol.PositionEncodingKindUTF16)
	documentURI := uri.URI("file:///document.oris")
	id := protocol.VersionedTextDocumentIdentifier{
		TextDocumentIdentifier: protocol.TextDocumentIdentifier{URI: documentURI},
		Version:                2,
	}

	if _, err := store.change(id, []protocol.TextDocumentContentChangeEvent{
		&protocol.TextDocumentContentChangeWholeDocument{Text: "text"},
	}); !errors.Is(err, errDocumentNotOpen) {
		t.Fatalf("unknown document error = %v, want %v", err, errDocumentNotOpen)
	}

	if err := store.open(protocol.TextDocumentItem{
		URI: documentURI, Version: 1, Text: "first",
	}); err != nil {
		t.Fatal(err)
	}
	invalidChanges := [][]protocol.TextDocumentContentChangeEvent{
		nil,
		{
			&protocol.TextDocumentContentChangeWholeDocument{Text: "one"},
			&protocol.TextDocumentContentChangeWholeDocument{Text: "two"},
		},
		{
			&protocol.TextDocumentContentChangePartial{
				Range: protocol.Range{}, Text: "partial",
			},
		},
	}
	for _, changes := range invalidChanges {
		if _, err := store.change(id, changes); !errors.Is(err, errFullDocumentChangeRequired) {
			t.Errorf("change error = %v, want %v", err, errFullDocumentChangeRequired)
		}
	}
	assertDocument(t, store, documentURI, 1, "first")
}

func TestServerDocumentHandlers(t *testing.T) {
	srv := newServer()
	documentURI := uri.URI("file:///document.oris")
	result, err := srv.Initialize(t.Context(), &protocol.InitializeParams{
		Capabilities: protocol.ClientCapabilities{
			General: &protocol.GeneralClientCapabilities{
				PositionEncodings: []protocol.PositionEncodingKind{protocol.PositionEncodingKindUTF8},
			},
		},
	})
	if err != nil {
		t.Fatalf("initialize: %v", err)
	}
	if result.Capabilities.PositionEncoding != protocol.PositionEncodingKindUTF8 {
		t.Fatalf("position encoding = %q, want UTF-8", result.Capabilities.PositionEncoding)
	}

	if err := srv.DidOpen(t.Context(), &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI: documentURI, Version: 3, Text: "open",
		},
	}); err != nil {
		t.Fatalf("didOpen: %v", err)
	}
	assertDocument(t, srv.documents, documentURI, 3, "open")
	document, _ := srv.documents.get(documentURI)
	if document.mapper.encoding != result.Capabilities.PositionEncoding {
		t.Errorf("mapper encoding = %q, capability = %q", document.mapper.encoding, result.Capabilities.PositionEncoding)
	}

	if err := srv.DidChange(t.Context(), &protocol.DidChangeTextDocumentParams{
		TextDocument: protocol.VersionedTextDocumentIdentifier{
			TextDocumentIdentifier: protocol.TextDocumentIdentifier{URI: documentURI},
			Version:                4,
		},
		ContentChanges: []protocol.TextDocumentContentChangeEvent{
			&protocol.TextDocumentContentChangeWholeDocument{Text: "changed"},
		},
	}); err != nil {
		t.Fatalf("didChange: %v", err)
	}
	assertDocument(t, srv.documents, documentURI, 4, "changed")

	if err := srv.DidClose(t.Context(), &protocol.DidCloseTextDocumentParams{
		TextDocument: protocol.TextDocumentIdentifier{URI: documentURI},
	}); err != nil {
		t.Fatalf("didClose: %v", err)
	}
	if _, ok := srv.documents.get(documentURI); ok {
		t.Fatal("document remains available after didClose")
	}
}

func assertDocument(t *testing.T, store *documentStore, documentURI uri.URI, version int32, source string) {
	t.Helper()
	document, ok := store.get(documentURI)
	if !ok {
		t.Fatal("document not found")
	}
	if document.URI != documentURI || document.Version != version || document.Source != source {
		t.Errorf("document = %#v, want URI %q, version %d, source %q", document, documentURI, version, source)
	}
}
