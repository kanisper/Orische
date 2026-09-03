package lsp

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"go.lsp.dev/jsonrpc2"
	"go.lsp.dev/protocol"
)

func TestServeLifecycle(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()

	clientStream, serverStream := jsonrpc2.NewChannelStreamPair(8)
	serveErr := make(chan error, 1)
	go func() {
		serveErr <- serveStream(ctx, serverStream)
	}()

	_, clientConn, server := protocol.NewClient(ctx, protocol.UnimplementedClient{}, clientStream)
	defer func() { _ = clientConn.Close() }()

	result, err := server.Initialize(ctx, &protocol.InitializeParams{})
	if err != nil {
		t.Fatalf("initialize: %v", err)
	}
	if result.ServerInfo.Name != "orische" {
		t.Errorf("server name = %q, want %q", result.ServerInfo.Name, "orische")
	}
	openClose := true
	full := protocol.TextDocumentSyncKindFull
	wantCapabilities := protocol.ServerCapabilities{
		PositionEncoding: protocol.PositionEncodingKindUTF16,
		TextDocumentSync: &protocol.TextDocumentSyncOptions{
			OpenClose: &openClose,
			Change:    &full,
		},
		CompletionProvider: &protocol.CompletionOptions{
			TriggerCharacters: []string{"["},
		},
	}
	if diff := cmp.Diff(wantCapabilities, result.Capabilities); diff != "" {
		t.Errorf("capabilities mismatch (-want +got):\n%s", diff)
	}

	if err := server.Initialized(ctx, &protocol.InitializedParams{}); err != nil {
		t.Fatalf("initialized: %v", err)
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
		t.Fatal("server did not stop after exit")
	}
}

func TestServeExitWithoutShutdownFails(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()

	clientStream, serverStream := jsonrpc2.NewChannelStreamPair(8)
	serveErr := make(chan error, 1)
	go func() { serveErr <- serveStream(ctx, serverStream) }()

	_, clientConn, server := protocol.NewClient(ctx, protocol.UnimplementedClient{}, clientStream)
	defer func() { _ = clientConn.Close() }()

	if _, err := server.Initialize(ctx, &protocol.InitializeParams{}); err != nil {
		t.Fatalf("initialize: %v", err)
	}
	if err := server.Exit(ctx); err != nil {
		t.Fatalf("exit: %v", err)
	}

	select {
	case err := <-serveErr:
		if !errors.Is(err, errExitWithoutShutdown) {
			t.Fatalf("serve error = %v, want %v", err, errExitWithoutShutdown)
		}
	case <-ctx.Done():
		t.Fatal("server did not stop after exit")
	}
}

func TestServeRejectsRequestAfterShutdown(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()

	clientStream, serverStream := jsonrpc2.NewChannelStreamPair(8)
	serveErr := make(chan error, 1)
	go func() { serveErr <- serveStream(ctx, serverStream) }()

	_, clientConn, server := protocol.NewClient(ctx, protocol.UnimplementedClient{}, clientStream)
	defer func() { _ = clientConn.Close() }()

	if _, err := server.Initialize(ctx, &protocol.InitializeParams{}); err != nil {
		t.Fatalf("initialize: %v", err)
	}
	if err := server.Shutdown(ctx); err != nil {
		t.Fatalf("shutdown: %v", err)
	}
	if _, err := server.Completion(ctx, &protocol.CompletionParams{}); !errors.Is(err, jsonrpc2.ErrInvalidRequest) {
		t.Fatalf("completion error = %v, want %v", err, jsonrpc2.ErrInvalidRequest)
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
		t.Fatal("server did not stop after exit")
	}
}

func TestServerIgnoresDocumentNotificationsAfterShutdown(t *testing.T) {
	srv := newServer()
	if err := srv.Shutdown(t.Context()); err != nil {
		t.Fatal(err)
	}
	params := &protocol.DidOpenTextDocumentParams{TextDocument: protocol.TextDocumentItem{
		URI: "file:///shutdown.oris", Version: 1, Text: "text",
	}}
	if err := srv.DidOpen(t.Context(), params); err != nil {
		t.Fatalf("didOpen: %v", err)
	}
	if _, ok := srv.documents.get(params.TextDocument.URI); ok {
		t.Fatal("didOpen after shutdown changed document state")
	}
}
