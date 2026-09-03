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
	exited    chan struct{}
	exitOnce  sync.Once
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

func (s *server) DidOpen(_ context.Context, params *protocol.DidOpenTextDocumentParams) error {
	if params != nil {
		_ = s.documents.open(params.TextDocument)
	}
	return nil
}

func (s *server) DidChange(_ context.Context, params *protocol.DidChangeTextDocumentParams) error {
	if params != nil {
		_, _ = s.documents.change(params.TextDocument, params.ContentChanges)
	}
	return nil
}

func (s *server) DidClose(_ context.Context, params *protocol.DidCloseTextDocumentParams) error {
	if params != nil {
		s.documents.close(params.TextDocument.URI)
	}
	return nil
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
