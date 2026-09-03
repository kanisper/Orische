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

	exited   chan struct{}
	exitOnce sync.Once
}

func newServer() *server {
	return &server{exited: make(chan struct{})}
}

func (s *server) Initialize(context.Context, *protocol.InitializeParams) (*protocol.InitializeResult, error) {
	return &protocol.InitializeResult{
		ServerInfo: protocol.ServerInfo{Name: "orische"},
	}, nil
}

func (s *server) Initialized(context.Context, *protocol.InitializedParams) error {
	return nil
}

func (s *server) Shutdown(context.Context) error {
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
