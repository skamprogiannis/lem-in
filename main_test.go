package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
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

func TestRunOfficialExamples(t *testing.T) {
	tests := []struct {
		file     string
		maxTurns int
	}{
		{"examples/example00.txt", 6},
		{"examples/example01.txt", 8},
		{"examples/example02.txt", 11},
		{"examples/example03.txt", 6},
		{"examples/example04.txt", 6},
		{"examples/example05.txt", 8},
		{"examples/example06.txt", 0},
		{"examples/example07.txt", 0},
	}

	for _, tc := range tests {
		t.Run(tc.file, func(t *testing.T) {
			raw, err := os.ReadFile(tc.file)
			if err != nil {
				t.Fatalf("read fixture: %v", err)
			}

			var stdout bytes.Buffer
			if err := run([]string{tc.file}, &stdout); err != nil {
				t.Fatalf("run: %v", err)
			}

			prefix := string(raw)
			if len(raw) == 0 || raw[len(raw)-1] != '\n' {
				prefix += "\n"
			}
			prefix += "\n"
			if !strings.HasPrefix(stdout.String(), prefix) {
				t.Fatal("output does not preserve the input and blank-line separator")
			}

			moves := strings.TrimSuffix(strings.TrimPrefix(stdout.String(), prefix), "\n")
			if moves == "" {
				t.Fatal("output contains no ant movements")
			}
			if turns := strings.Count(moves, "\n") + 1; tc.maxTurns > 0 && turns > tc.maxTurns {
				t.Errorf("completed in %d turns, want at most %d", turns, tc.maxTurns)
			}
		})
	}
}

func TestRunRejectsOfficialBadExamples(t *testing.T) {
	for _, file := range []string{
		"examples/badexample00.txt",
		"examples/badexample01.txt",
	} {
		t.Run(file, func(t *testing.T) {
			if err := run([]string{file}, &bytes.Buffer{}); err == nil {
				t.Fatal("run accepted invalid input")
			}
		})
	}
}
