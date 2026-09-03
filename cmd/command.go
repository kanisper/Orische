package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	lspserver "orische/internal/lsp"
	"orische/internal/parser"
	htmlrenderer "orische/internal/render/html"
)

const (
	exitSuccess = 0
	exitFailure = 1
	exitUsage   = 2
)

func run(args []string, stderr io.Writer) int {
	return runWithIO(args, os.Stdin, os.Stdout, stderr)
}

func runWithIO(args []string, stdin io.ReadCloser, stdout, stderr io.Writer) int {
	if len(args) > 0 && args[0] == "lsp" {
		if len(args) != 1 {
			fmt.Fprintln(stderr, "usage: orische lsp")
			return exitUsage
		}
		if err := lspserver.Serve(context.Background(), stdin, stdout); err != nil {
			printError(stderr, "", err)
			return exitFailure
		}
		return exitSuccess
	}

	flags := flag.NewFlagSet("orische", flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.Usage = func() {
		fmt.Fprintln(stderr, "usage: orische [-o output.html] input.oris")
	}

	var outputPath string
	flags.StringVar(&outputPath, "o", "", "output HTML file")

	if err := flags.Parse(args); err != nil {
		return exitUsage
	}
	if flags.NArg() != 1 {
		flags.Usage()
		return exitUsage
	}

	inputPath := flags.Arg(0)

	if outputPath == "" {
		outputPath = defaultOutputPath(inputPath)
	}
	inputAbs, err := filepath.Abs(inputPath)
	if err != nil {
		printError(stderr, "", fmt.Errorf("resolve input path %q: %w", inputPath, err))
		return exitFailure
	}
	outputAbs, err := filepath.Abs(outputPath)
	if err != nil {
		printError(stderr, "", fmt.Errorf("resolve output path %q: %w", outputPath, err))
		return exitFailure
	}
	if inputAbs == outputAbs {
		fmt.Fprintln(stderr, "orische: input and output paths must differ")
		flags.Usage()
		return exitUsage
	}

	if err := convertFile(inputPath, outputPath); err != nil {
		printError(stderr, inputPath, err)
		return exitFailure
	}

	return exitSuccess
}

func convertFile(inputPath, outputPath string) error {
	source, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("read %q: %w", inputPath, err)
	}

	doc, err := parser.Parse(string(source))
	if err != nil {
		return err
	}

	var output bytes.Buffer

	if err := htmlrenderer.Render(&output, doc); err != nil {
		return err
	}

	if err := os.WriteFile(outputPath, output.Bytes(), 0o644); err != nil {
		return fmt.Errorf("write %q: %w", outputPath, err)
	}

	return nil
}

func defaultOutputPath(inputPath string) string {
	ext := filepath.Ext(inputPath)
	if ext == "" {
		return inputPath + ".html"
	}
	return strings.TrimSuffix(inputPath, ext) + ".html"
}
