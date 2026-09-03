package main

import (
	"errors"
	"fmt"
	"io"

	"orische/internal/diagnostic"
)

func printError(w io.Writer, path string, err error) {
	var diag *diagnostic.Error
	if errors.As(err, &diag) {
		printDiagnostic(w, path, diag)
		return
	}

	fmt.Fprintf(w, "orische: %v\n", err)
}

func printDiagnostic(w io.Writer, path string, diag *diagnostic.Error) {
	diagPosition := diag.Range.Start

	if path == "" {
		fmt.Fprintf(
			w,
			"orische: line %d:%d error: %s\n",
			diagPosition.Line,
			diagPosition.Column,
			diag.Message,
		)
	} else {
		fmt.Fprintf(
			w,
			"orische: %s:%d:%d error: %s\n",
			path,
			diagPosition.Line,
			diagPosition.Column,
			diag.Message,
		)
	}
}
