package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"orische/internal/parser"
	htmlrenderer "orische/internal/render/html"
)

const (
	exitSuccess = 0
	exitFailure = 1
	exitUsage   = 2
)

func run(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet(args[0], flag.ExitOnError)
	flags.SetOutput(stderr)

	var outputPath string
	flags.StringVar(&outputPath, "o", "", "output HTML file")

	if err := flags.Parse(args[1:]); err != nil {
		return exitUsage
	}
	if flags.NArg() != 1 {
		fmt.Fprintln(stderr, "usage: orische [-o output.html] input.oris")
		return exitUsage
	}

	inputPath := flags.Arg(0)

	if outputPath == "" {
		outputPath = defaultOutputPath(inputPath)
	}

	if err := convertFile(inputPath, outputPath); err != nil {
		fmt.Fprintf(stderr, "orische: %v\n", err)
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
		return fmt.Errorf("parse %q: %w", inputPath, err)
	}

	var output bytes.Buffer

	if err := htmlrenderer.Render(&output, doc); err != nil {
		return fmt.Errorf("render %q: %w", inputPath, err)
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
