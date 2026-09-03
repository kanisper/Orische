package main

import (
	"bytes"
	"context"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"go.lsp.dev/jsonrpc2"
	"go.lsp.dev/protocol"
)

func TestRunLSP(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()

	serverConn, clientConn := net.Pipe()
	defer func() { _ = clientConn.Close() }()

	var stderr bytes.Buffer
	exitCode := make(chan int, 1)
	go func() {
		exitCode <- runWithIO([]string{"lsp"}, serverConn, serverConn, &stderr)
	}()

	_, conn, server := protocol.NewClient(ctx, protocol.UnimplementedClient{}, jsonrpc2.NewStream(clientConn))
	defer func() { _ = conn.Close() }()

	result, err := server.Initialize(ctx, &protocol.InitializeParams{})
	if err != nil {
		t.Fatalf("initialize: %v", err)
	}
	if result.ServerInfo.Name != "orische" {
		t.Errorf("server name = %q, want %q", result.ServerInfo.Name, "orische")
	}
	if err := server.Shutdown(ctx); err != nil {
		t.Fatalf("shutdown: %v", err)
	}
	if err := server.Exit(ctx); err != nil {
		t.Fatalf("exit: %v", err)
	}

	select {
	case got := <-exitCode:
		if got != exitSuccess {
			t.Errorf("run exit = %d, want %d; stderr: %s", got, exitSuccess, stderr.String())
		}
	case <-ctx.Done():
		t.Fatal("orische lsp did not stop after exit")
	}
}

func TestRunLSPExitWithoutShutdownFails(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()

	serverConn, clientConn := net.Pipe()
	defer func() { _ = clientConn.Close() }()

	var stderr bytes.Buffer
	exitCode := make(chan int, 1)
	go func() {
		exitCode <- runWithIO([]string{"lsp"}, serverConn, serverConn, &stderr)
	}()

	_, conn, server := protocol.NewClient(ctx, protocol.UnimplementedClient{}, jsonrpc2.NewStream(clientConn))
	defer func() { _ = conn.Close() }()

	if _, err := server.Initialize(ctx, &protocol.InitializeParams{}); err != nil {
		t.Fatalf("initialize: %v", err)
	}
	if err := server.Exit(ctx); err != nil {
		t.Fatalf("exit: %v", err)
	}

	select {
	case got := <-exitCode:
		if got != exitFailure {
			t.Errorf("run exit = %d, want %d", got, exitFailure)
		}
		if !strings.Contains(stderr.String(), "exit received before shutdown") {
			t.Errorf("stderr = %q, want exit-without-shutdown error", stderr.String())
		}
	case <-ctx.Done():
		t.Fatal("orische lsp did not stop after exit")
	}
}

func TestRunLSPRejectsArguments(t *testing.T) {
	var stderr bytes.Buffer
	exitCode := run([]string{"lsp", "extra"}, &stderr)

	if exitCode != exitUsage {
		t.Errorf("run exit = %d, want %d", exitCode, exitUsage)
	}
	if want := "usage: orische lsp"; !strings.Contains(stderr.String(), want) {
		t.Errorf("stderr = %q, want it to contain %q", stderr.String(), want)
	}
}

func TestRunConvertsFile(t *testing.T) {
	dir := t.TempDir()
	inputPath := filepath.Join(dir, "input.oris")
	outputPath := filepath.Join(dir, "output.html")
	writeTestInput(t, inputPath, "= heading 1")

	var stderr bytes.Buffer
	exitCode := run([]string{"-o", outputPath, inputPath}, &stderr)

	if exitCode != exitSuccess {
		t.Fatalf("run exit = %d, want %d; stderr: %s", exitCode, exitSuccess, stderr.String())
	}
	if stderr.Len() != 0 {
		t.Errorf("stderr = %q, want empty", stderr.String())
	}

	got, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatal(err)
	}
	if want := "<h1>heading 1</h1>\n"; string(got) != want {
		t.Errorf("output = %q, want %q", got, want)
	}
}

func TestRunUsesDefaultOutputPath(t *testing.T) {
	dir := t.TempDir()
	inputPath := filepath.Join(dir, "input.oris")
	writeTestInput(t, inputPath, "paragraph")

	var stderr bytes.Buffer
	exitCode := run([]string{inputPath}, &stderr)

	if exitCode != exitSuccess {
		t.Fatalf("run exit = %d, want %d; stderr: %s", exitCode, exitSuccess, stderr.String())
	}
	got, err := os.ReadFile(filepath.Join(dir, "input.html"))
	if err != nil {
		t.Fatal(err)
	}
	if want := "<p>\nparagraph\n</p>\n"; string(got) != want {
		t.Errorf("output = %q, want %q", got, want)
	}
}

func TestRunRejectsUsageErrors(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		wantStderr string
	}{
		{name: "missing input", wantStderr: "usage: orische [-o output.html] input.oris"},
		{name: "extra input", args: []string{"one.oris", "two.oris"}, wantStderr: "usage: orische [-o output.html] input.oris"},
		{name: "unknown flag", args: []string{"-unknown"}, wantStderr: "flag provided but not defined: -unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stderr bytes.Buffer
			exitCode := run(tt.args, &stderr)

			if exitCode != exitUsage {
				t.Errorf("run exit = %d, want %d", exitCode, exitUsage)
			}
			if !strings.Contains(stderr.String(), tt.wantStderr) {
				t.Errorf("stderr = %q, want it to contain %q", stderr.String(), tt.wantStderr)
			}
		})
	}
}

func TestRunRejectsInputOutputCollision(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		args     func(string) []string
	}{
		{
			name:     "explicit output",
			filename: "input.oris",
			args: func(path string) []string {
				separator := string(filepath.Separator)
				outputPath := filepath.Dir(path) + separator + "." + separator + filepath.Base(path)
				return []string{"-o", outputPath, path}
			},
		},
		{
			name:     "default output",
			filename: "input.html",
			args:     func(path string) []string { return []string{path} },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inputPath := filepath.Join(t.TempDir(), tt.filename)
			original := []byte("= must remain source")
			writeTestInput(t, inputPath, string(original))

			var stderr bytes.Buffer
			exitCode := run(tt.args(inputPath), &stderr)

			if exitCode != exitUsage {
				t.Errorf("run exit = %d, want %d", exitCode, exitUsage)
			}
			got, err := os.ReadFile(inputPath)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(got, original) {
				t.Errorf("input was modified: got %q, want %q", got, original)
			}
		})
	}
}

func TestRunReportsInputReadFailure(t *testing.T) {
	inputPath := filepath.Join(t.TempDir(), "missing.oris")
	var stderr bytes.Buffer

	exitCode := run([]string{inputPath}, &stderr)

	if exitCode != exitFailure {
		t.Errorf("run exit = %d, want %d", exitCode, exitFailure)
	}
	if want := "orische: read \"" + inputPath + "\""; !strings.Contains(stderr.String(), want) {
		t.Errorf("stderr = %q, want it to contain %q", stderr.String(), want)
	}
}

func TestRunReportsAbsolutePathFailure(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows does not allow removing the current working directory")
	}

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name       string
		args       []string
		wantStderr string
	}{
		{name: "input", args: []string{"input.oris"}, wantStderr: "resolve input path"},
		{
			name:       "output",
			args:       []string{"-o", "output.html", filepath.Join(originalDir, "input.oris")},
			wantStderr: "resolve output path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			workingDir := t.TempDir()
			if err := os.Chdir(workingDir); err != nil {
				t.Fatal(err)
			}
			defer func() {
				if err := os.Chdir(originalDir); err != nil {
					t.Errorf("restore working directory: %v", err)
				}
			}()
			if err := os.Remove(workingDir); err != nil {
				t.Fatal(err)
			}

			var stderr bytes.Buffer
			exitCode := run(tt.args, &stderr)

			if exitCode != exitFailure {
				t.Errorf("run exit = %d, want %d", exitCode, exitFailure)
			}
			if !strings.Contains(stderr.String(), tt.wantStderr) {
				t.Errorf("stderr = %q, want it to contain %q", stderr.String(), tt.wantStderr)
			}
		})
	}
}

func TestDefaultOutputPath(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"document.oris", "document.html"},
		{"document", "document.html"},
		{"docs/guide.oris", "docs/guide.html"},
		{"guide.ja.oris", "guide.ja.html"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := defaultOutputPath(tt.input); got != tt.want {
				t.Errorf("defaultOutputPath(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func writeTestInput(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
