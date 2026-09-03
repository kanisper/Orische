package lsp

import (
	"context"
	"fmt"
	"io"
	"sync"

	"go.lsp.dev/jsonrpc2"
	"go.lsp.dev/protocol"
)

type server struct {
	protocol.UnimplementedServer

	documents *documentStore
	encoding  protocol.PositionEncodingKind
	// documentActions keeps document state changes and their diagnostic
	// publications in protocol order without holding the store lock during I/O.
	documentActions sync.Mutex
	exited          chan struct{}
	exitOnce        sync.Once
}

func newServer() *server {
	encoding := protocol.PositionEncodingKindUTF16
	return &server{
		documents: newDocumentStore(encoding),
		encoding:  encoding,
		exited:    make(chan struct{}),
	}
}

func (s *server) Initialize(_ context.Context, params *protocol.InitializeParams) (*protocol.InitializeResult, error) {
	var encodings []protocol.PositionEncodingKind
	if params != nil && params.Capabilities.General != nil {
		encodings = params.Capabilities.General.PositionEncodings
	}
	s.encoding = negotiatePositionEncoding(encodings)
	s.documents = newDocumentStore(s.encoding)

	openClose := true
	full := protocol.TextDocumentSyncKindFull
	return &protocol.InitializeResult{
		Capabilities: protocol.ServerCapabilities{
			PositionEncoding: s.encoding,
			TextDocumentSync: &protocol.TextDocumentSyncOptions{
				OpenClose: &openClose,
				Change:    &full,
			},
			CompletionProvider: &protocol.CompletionOptions{
				TriggerCharacters: []string{"["},
			},
		},
		ServerInfo: protocol.ServerInfo{Name: "orische"},
	}, nil
}

func (s *server) Initialized(context.Context, *protocol.InitializedParams) error {
	return nil
}

func (s *server) Shutdown(context.Context) error {
	return nil
}

func (s *server) DidOpen(ctx context.Context, params *protocol.DidOpenTextDocumentParams) error {
	if params == nil {
		return nil
	}
	s.documentActions.Lock()
	defer s.documentActions.Unlock()
	if err := s.documents.open(params.TextDocument); err != nil {
		return nil
	}
	return s.analyzeDocument(ctx, params.TextDocument.URI)
}

func (s *server) DidChange(ctx context.Context, params *protocol.DidChangeTextDocumentParams) error {
	if params == nil {
		return nil
	}
	s.documentActions.Lock()
	defer s.documentActions.Unlock()
	applied, err := s.documents.change(params.TextDocument, params.ContentChanges)
	if err != nil || !applied {
		return nil
	}
	return s.analyzeDocument(ctx, params.TextDocument.URI)
}

func (s *server) DidClose(ctx context.Context, params *protocol.DidCloseTextDocumentParams) error {
	if params == nil {
		return nil
	}
	s.documentActions.Lock()
	defer s.documentActions.Unlock()
	if !s.documents.close(params.TextDocument.URI) {
		return nil
	}
	return publishDiagnostics(ctx, params.TextDocument.URI, nil, nil)
}

func (s *server) Exit(context.Context) error {
	s.exitOnce.Do(func() { close(s.exited) })
	return nil
}

// Serve runs an LSP server over an LSP Content-Length framed byte stream.
func Serve(ctx context.Context, input io.ReadCloser, output io.Writer) error {
	if input == nil {
		return fmt.Errorf("serve LSP: input is nil")
	}
	if output == nil {
		return fmt.Errorf("serve LSP: output is nil")
	}

	stream := jsonrpc2.NewStream(&readWriteCloser{
		Reader: input,
		Writer: output,
		Closer: input,
	})
	return serveStream(ctx, stream)
}

func serveStream(ctx context.Context, stream jsonrpc2.Stream) error {
	srv := newServer()
	_, conn, _ := protocol.NewServer(ctx, srv, stream)

	select {
	case <-srv.exited:
	case <-conn.Done():
		return conn.Err()
	case <-ctx.Done():
	}

	if err := conn.Close(); err != nil {
		return fmt.Errorf("close LSP connection: %w", err)
	}
	if err := conn.Err(); err != nil {
		return fmt.Errorf("serve LSP: %w", err)
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	return nil
}

type readWriteCloser struct {
	io.Reader
	io.Writer
	io.Closer
}
