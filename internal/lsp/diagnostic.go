package lsp

import (
	"context"
	"errors"
	"fmt"

	"go.lsp.dev/protocol"
	"go.lsp.dev/uri"

	"orische/internal/diagnostic"
	"orische/internal/parser"
)

func protocolDiagnostic(mapper *positionMapper, err error) (protocol.Diagnostic, bool, error) {
	var sourceError *diagnostic.Error
	if !errors.As(err, &sourceError) {
		return protocol.Diagnostic{}, false, nil
	}

	sourceRange, err := mapper.astRange(sourceError.Range)
	if err != nil {
		return protocol.Diagnostic{}, false, fmt.Errorf("map diagnostic range: %w", err)
	}
	return protocol.Diagnostic{
		Range:    sourceRange,
		Severity: protocol.DiagnosticSeverityError,
		Source:   protocol.NewOptional("orische"),
		Message:  protocol.String(sourceError.Message),
	}, true, nil
}

func analyze(document document) (analysis, error) {
	documentAST, err := parser.Parse(document.Source)
	if err == nil {
		return analysis{
			AST:         documentAST,
			Diagnostics: make([]protocol.Diagnostic, 0),
		}, nil
	}

	converted, ok, conversionErr := protocolDiagnostic(document.mapper, err)
	if conversionErr != nil {
		return analysis{Diagnostics: make([]protocol.Diagnostic, 0)}, conversionErr
	}
	if !ok {
		return analysis{Diagnostics: make([]protocol.Diagnostic, 0)}, err
	}
	return analysis{Diagnostics: []protocol.Diagnostic{converted}}, nil
}

func (s *server) analyzeDocument(ctx context.Context, documentURI uri.URI) error {
	document, ok := s.documents.get(documentURI)
	if !ok {
		return nil
	}

	result, analysisErr := analyze(document)
	if !s.documents.setAnalysis(document, result) {
		return nil
	}
	if analysisErr != nil {
		protocol.LoggerFromContext(ctx).Error(
			"analyze document",
			"uri", documentURI,
			"version", document.Version,
			"error", analysisErr,
		)
	}
	return publishDiagnostics(ctx, documentURI, &document.Version, result.Diagnostics)
}

func publishDiagnostics(
	ctx context.Context,
	documentURI uri.URI,
	version *int32,
	diagnostics []protocol.Diagnostic,
) error {
	client, ok := protocol.ClientFromContext(ctx)
	if !ok {
		return nil
	}
	if diagnostics == nil {
		diagnostics = make([]protocol.Diagnostic, 0)
	}
	params := &protocol.PublishDiagnosticsParams{
		URI:         documentURI,
		Diagnostics: diagnostics,
	}
	if version != nil {
		params.Version = protocol.NewOptional(*version)
	}
	return client.PublishDiagnostics(ctx, params)
}
