package main

import (
	"errors"
	"fmt"
	"io"

	"orische/internal/diagnostic"
)

func printErrors(w io.Writer, path string, err error) {
	var diag *diagnostic.Error
	if errors.As(err, &diag) {
		printDiagnostic(w, path, diag)
		return
	}

	fmt.Fprintf(w, "orische: %v\n", err)
}

func printDiagnostic(w io.Writer, path string, diag *diagnostic.Error) {
	line := diag.Range.StartLine

	if path == "" {
		fmt.Fprintf(
			w,
			"orische: line %d: error: %s\n",
			line,
			diag.Message,
		)
	} else {
		fmt.Fprintf(
			w,
			"orische: %s:%d: error: %s\n",
			path,
			line,
			diag.Message,
		)
	}
}
