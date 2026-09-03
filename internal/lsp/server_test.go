package lsp

import (
	"context"
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
