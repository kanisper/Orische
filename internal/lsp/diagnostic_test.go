package lsp

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"go.lsp.dev/jsonrpc2"
	"go.lsp.dev/protocol"
	"go.lsp.dev/uri"

	"orische/internal/ast"
	"orische/internal/diagnostic"
)

func TestProtocolDiagnostic(t *testing.T) {
	mapper, err := newPositionMapper("🍣 text", protocol.PositionEncodingKindUTF16)
	if err != nil {
		t.Fatal(err)
	}
	sourceError := &diagnostic.Error{
		Message: "problem",
		Range: ast.Range{
			Start: ast.Position{Line: 1, Column: 1},
			End:   ast.Position{Line: 1, Column: 1},
		},
	}

	got, ok, err := protocolDiagnostic(mapper, fmt.Errorf("wrapped: %w", sourceError))
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("wrapped diagnostic was not recognized")
	}
	if want := (protocol.Range{
		Start: protocol.Position{Line: 0, Character: 0},
		End:   protocol.Position{Line: 0, Character: 2},
	}); got.Range != want {
		t.Errorf("range = %v, want %v", got.Range, want)
	}
	if got.Severity != protocol.DiagnosticSeverityError {
		t.Errorf("severity = %v, want Error", got.Severity)
	}
	if source, present := got.Source.Get(); !present || source != "orische" {
		t.Errorf("source = %q, present %v; want orische", source, present)
	}
	if message, ok := got.Message.(protocol.String); !ok || message != "problem" {
		t.Errorf("message = %#v, want %q", got.Message, "problem")
	}
}

func TestProtocolDiagnosticIgnoresOrdinaryErrors(t *testing.T) {
	mapper, err := newPositionMapper("text", protocol.PositionEncodingKindUTF16)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok, err := protocolDiagnostic(mapper, errors.New("ordinary")); err != nil || ok {
		t.Errorf("ordinary error conversion = (ok %v, err %v), want (false, nil)", ok, err)
	}
}

func TestServerAnalysisTracksCurrentDocument(t *testing.T) {
	srv := newServer()
	documentURI := uri.URI("file:///document.oris")
	if err := srv.DidOpen(t.Context(), &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI: documentURI, Version: 1, Text: "paragraph",
		},
	}); err != nil {
		t.Fatal(err)
	}
	document, _ := srv.documents.get(documentURI)
	if document.analysis.AST == nil || len(document.analysis.Diagnostics) != 0 {
		t.Fatalf("successful analysis = %#v, want AST and no diagnostics", document.analysis)
	}

	changeDocument(t, srv, documentURI, 2, ":::[unsupported]\ntext\n:::")
	document, _ = srv.documents.get(documentURI)
	if document.analysis.AST != nil || len(document.analysis.Diagnostics) != 1 {
		t.Fatalf("diagnostic analysis = %#v, want nil AST and one diagnostic", document.analysis)
	}

	changeDocument(t, srv, documentURI, 3, ":::[heading:level7]\ntext\n:::")
	document, _ = srv.documents.get(documentURI)
	if document.analysis.AST != nil || len(document.analysis.Diagnostics) != 0 {
		t.Fatalf("ordinary-error analysis = %#v, want nil AST and no fake diagnostic", document.analysis)
	}

	changeDocument(t, srv, documentURI, 4, "current")
	changeDocument(t, srv, documentURI, 3, ":::[unsupported]\nstale\n:::")
	document, _ = srv.documents.get(documentURI)
	if document.Version != 4 || document.Source != "current" || document.analysis.AST == nil {
		t.Errorf("stale change replaced current analysis: %#v", document)
	}
}

func TestLiveDiagnosticsProtocol(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()

	clientStream, serverStream := jsonrpc2.NewChannelStreamPair(16)
	serveErr := make(chan error, 1)
	go func() { serveErr <- serveStream(ctx, serverStream) }()

	recorder := &diagnosticClient{published: make(chan protocol.PublishDiagnosticsParams, 4)}
	_, clientConn, server := protocol.NewClient(ctx, recorder, clientStream)
	defer func() { _ = clientConn.Close() }()

	if _, err := server.Initialize(ctx, &protocol.InitializeParams{}); err != nil {
		t.Fatalf("initialize: %v", err)
	}
	if err := server.Initialized(ctx, &protocol.InitializedParams{}); err != nil {
		t.Fatalf("initialized: %v", err)
	}

	documentURI := uri.URI("file:///document.oris")
	if err := server.DidOpen(ctx, &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI: documentURI, Version: 1,
			Text: ":::[unsupported]\n🍣\n:::",
		},
	}); err != nil {
		t.Fatalf("didOpen: %v", err)
	}
	opened := receiveDiagnostics(t, ctx, recorder.published)
	if opened.URI != documentURI || len(opened.Diagnostics) != 1 {
		t.Fatalf("didOpen diagnostics = %#v, want one diagnostic for %q", opened, documentURI)
	}
	if version, ok := opened.Version.Get(); !ok || version != 1 {
		t.Errorf("didOpen diagnostic version = %d, present %v; want 1", version, ok)
	}

	if err := server.DidChange(ctx, &protocol.DidChangeTextDocumentParams{
		TextDocument: protocol.VersionedTextDocumentIdentifier{
			TextDocumentIdentifier: protocol.TextDocumentIdentifier{URI: documentURI},
			Version:                2,
		},
		ContentChanges: []protocol.TextDocumentContentChangeEvent{
			&protocol.TextDocumentContentChangeWholeDocument{Text: "valid"},
		},
	}); err != nil {
		t.Fatalf("didChange: %v", err)
	}
	changed := receiveDiagnostics(t, ctx, recorder.published)
	if len(changed.Diagnostics) != 0 {
		t.Errorf("valid change diagnostics = %#v, want empty", changed.Diagnostics)
	}
	if version, ok := changed.Version.Get(); !ok || version != 2 {
		t.Errorf("didChange diagnostic version = %d, present %v; want 2", version, ok)
	}

	if err := server.DidClose(ctx, &protocol.DidCloseTextDocumentParams{
		TextDocument: protocol.TextDocumentIdentifier{URI: documentURI},
	}); err != nil {
		t.Fatalf("didClose: %v", err)
	}
	closed := receiveDiagnostics(t, ctx, recorder.published)
	if closed.URI != documentURI || len(closed.Diagnostics) != 0 {
		t.Errorf("didClose diagnostics = %#v, want empty for %q", closed, documentURI)
	}
	if _, ok := closed.Version.Get(); ok {
		t.Error("didClose diagnostic clear unexpectedly included a version")
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

type diagnosticClient struct {
	protocol.UnimplementedClient
	published chan protocol.PublishDiagnosticsParams
}

func (c *diagnosticClient) PublishDiagnostics(_ context.Context, params *protocol.PublishDiagnosticsParams) error {
	copyParams := *params
	copyParams.Diagnostics = append([]protocol.Diagnostic(nil), params.Diagnostics...)
	c.published <- copyParams
	return nil
}

func receiveDiagnostics(
	t *testing.T,
	ctx context.Context,
	from <-chan protocol.PublishDiagnosticsParams,
) protocol.PublishDiagnosticsParams {
	t.Helper()
	select {
	case diagnostics := <-from:
		return diagnostics
	case <-ctx.Done():
		t.Fatal("timed out waiting for diagnostics")
		return protocol.PublishDiagnosticsParams{}
	}
}

func changeDocument(t *testing.T, srv *server, documentURI uri.URI, version int32, source string) {
	t.Helper()
	if err := srv.DidChange(t.Context(), &protocol.DidChangeTextDocumentParams{
		TextDocument: protocol.VersionedTextDocumentIdentifier{
			TextDocumentIdentifier: protocol.TextDocumentIdentifier{URI: documentURI},
			Version:                version,
		},
		ContentChanges: []protocol.TextDocumentContentChangeEvent{
			&protocol.TextDocumentContentChangeWholeDocument{Text: source},
		},
	}); err != nil {
		t.Fatalf("change to version %d: %v", version, err)
	}
}
