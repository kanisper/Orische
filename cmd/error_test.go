package main

import (
	"bytes"
	"errors"
	"testing"

	"orische/internal/ast"
	"orische/internal/diagnostic"
)

func TestPrintErrorFormatsDiagnostic(t *testing.T) {
	diag := &diagnostic.Error{
		Message: "unsupported block directive",
		Range: ast.Range{
			Start: ast.Position{Line: 3, Column: 7},
		},
	}
	var output bytes.Buffer

	printError(&output, "input.oris", diag)

	if got, want := output.String(), "orische: input.oris:3:7 error: unsupported block directive\n"; got != want {
		t.Errorf("output = %q, want %q", got, want)
	}
}

func TestPrintErrorFormatsOrdinaryError(t *testing.T) {
	var output bytes.Buffer

	printError(&output, "input.oris", errors.New("conversion failed"))

	if got, want := output.String(), "orische: conversion failed\n"; got != want {
		t.Errorf("output = %q, want %q", got, want)
	}
}
