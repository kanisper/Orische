package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestRunConvertFile(t *testing.T) {
	dir := t.TempDir()

	inputPath := filepath.Join(dir, "input.oris")
	outputPath := filepath.Join(dir, "output.html")

	if err := os.WriteFile(
		inputPath,
		[]byte("= heading 1"),
		0o664,
	); err != nil {
		t.Fatal(err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := run(
		[]string{"orische", "-o", outputPath, inputPath},
		&stdout,
		&stderr,
	)

	if exitCode != 0 {
		t.Fatalf("run exit %d, stderr: %s", exitCode, stderr.String())
	}

	got, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatal(err)
	}

	want := "<h1>heading 1</h1>\n"
	if string(got) != want {
		t.Fatalf("got: %s, want: %s", string(got), want)
	}
}

func TestRunRequiresInputFile(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := run([]string{"orische"}, &stdout, &stderr)
	if exitCode != 2 {
		t.Fatalf("run exit %d expected, but want 2", exitCode)
	}

	if stderr.Len() == 0 {
		t.Fatal("expected usage message")
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
			got := defaultOutputPath(tt.input)
			if got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}
