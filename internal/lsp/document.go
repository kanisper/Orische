package lsp

import (
	"errors"
	"fmt"
	"sync"

	"go.lsp.dev/protocol"
	"go.lsp.dev/uri"

	"orische/internal/ast"
)

var (
	errDocumentAlreadyOpen        = errors.New("document is already open")
	errDocumentNotOpen            = errors.New("document is not open")
	errFullDocumentChangeRequired = errors.New("exactly one full document change is required")
)

type document struct {
	URI        uri.URI
	Version    int32
	Source     string
	mapper     *positionMapper
	generation uint64
	analysis   analysis
}

type analysis struct {
	AST         *ast.Document
	Diagnostics []protocol.Diagnostic
}

type documentStore struct {
	mu        sync.RWMutex
	encoding  protocol.PositionEncodingKind
	documents map[uri.URI]document
	nextID    uint64
}

func newDocumentStore(encoding protocol.PositionEncodingKind) *documentStore {
	return &documentStore{
		encoding:  encoding,
		documents: make(map[uri.URI]document),
	}
}

func (s *documentStore) open(item protocol.TextDocumentItem) error {
	mapper, err := newPositionMapper(item.Text, s.encoding)
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.documents[item.URI]; exists {
		return fmt.Errorf("open %q: %w", item.URI, errDocumentAlreadyOpen)
	}
	s.nextID++
	s.documents[item.URI] = document{
		URI: item.URI, Version: item.Version, Source: item.Text, mapper: mapper,
		generation: s.nextID,
	}
	return nil
}

func (s *documentStore) change(
	identifier protocol.VersionedTextDocumentIdentifier,
	changes []protocol.TextDocumentContentChangeEvent,
) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	current, exists := s.documents[identifier.URI]
	if !exists {
		return false, fmt.Errorf("change %q: %w", identifier.URI, errDocumentNotOpen)
	}
	if identifier.Version <= current.Version {
		return false, nil
	}
	if len(changes) != 1 {
		return false, errFullDocumentChangeRequired
	}
	whole, ok := changes[0].(*protocol.TextDocumentContentChangeWholeDocument)
	if !ok || whole == nil {
		return false, errFullDocumentChangeRequired
	}
	mapper, err := newPositionMapper(whole.Text, s.encoding)
	if err != nil {
		return false, err
	}

	s.documents[identifier.URI] = document{
		URI: identifier.URI, Version: identifier.Version, Source: whole.Text, mapper: mapper,
		generation: current.generation,
	}
	return true, nil
}

func (s *documentStore) close(documentURI uri.URI) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.documents[documentURI]; !exists {
		return false
	}
	delete(s.documents, documentURI)
	return true
}

func (s *documentStore) get(documentURI uri.URI) (document, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	document, ok := s.documents[documentURI]
	return document, ok
}

func (s *documentStore) setAnalysis(snapshot document, result analysis) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	current, ok := s.documents[snapshot.URI]
	if !ok || current.generation != snapshot.generation || current.Version != snapshot.Version {
		return false
	}
	current.analysis = result
	s.documents[snapshot.URI] = current
	return true
}
