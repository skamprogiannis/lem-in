package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestRunPreservesInputBeforeMoves(t *testing.T) {
	input := "4\n" +
		"##start\n" +
		"0 0 3\n" +
		"2 2 5\n" +
		"3 4 0\n" +
		"##end\n" +
		"1 8 3\n" +
		"0-2\n" +
		"2-3\n" +
		"3-1\n\n"
	want := input + "\n" +
		"L1-2\n" +
		"L1-3 L2-2\n" +
		"L1-1 L2-3 L3-2\n" +
		"L2-1 L3-3 L4-2\n" +
		"L3-1 L4-3\n" +
		"L4-1\n"

	filePath := filepath.Join(t.TempDir(), "example00.txt")
	if err := os.WriteFile(filePath, []byte(input), 0o600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	var stdout bytes.Buffer
	if err := run([]string{filePath}, &stdout); err != nil {
		t.Fatalf("run: %v", err)
	}
	if got := stdout.String(); got != want {
		t.Errorf("unexpected output\ngot:\n%q\nwant:\n%q", got, want)
	}
}
