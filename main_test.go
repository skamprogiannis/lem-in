package main

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"

	"lem-in/internal/graph"
	"lem-in/internal/parser"
)

func TestRunRequiresExactlyOneInputFile(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{name: "missing argument"},
		{name: "extra argument", args: []string{"example.txt", "extra"}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var stdout bytes.Buffer

			err := run(tc.args, &stdout)
			if err == nil {
				t.Fatal("run accepted invalid argument count")
			}
			if got, want := err.Error(), "expected exactly one input file"; got != want {
				t.Errorf("run error = %q, want %q", got, want)
			}
			if stdout.Len() != 0 {
				t.Errorf("run wrote stdout on failure: %q", stdout.String())
			}
		})
	}
}

func TestCLIReportsInvalidInput(t *testing.T) {
	binary := buildCLI(t)
	fixtureDir := t.TempDir()
	missingFile := filepath.Join(fixtureDir, "missing.txt")
	emptyFile := filepath.Join(fixtureDir, "empty.txt")
	disconnectedFile := filepath.Join(fixtureDir, "disconnected.txt")

	if err := os.WriteFile(emptyFile, nil, 0o600); err != nil {
		t.Fatalf("write empty fixture: %v", err)
	}
	disconnected := "2\n" +
		"##start\n" +
		"start 0 0\n" +
		"middle 1 0\n" +
		"##end\n" +
		"end 2 0\n" +
		"start-middle\n"
	if err := os.WriteFile(disconnectedFile, []byte(disconnected), 0o600); err != nil {
		t.Fatalf("write disconnected fixture: %v", err)
	}

	tests := []struct {
		name       string
		args       []string
		wantStderr string
		exact      bool
	}{
		{
			name:       "missing argument",
			wantStderr: "ERROR: invalid data format, expected exactly one input file\n",
			exact:      true,
		},
		{
			name:       "extra argument",
			args:       []string{"examples/example00.txt", "extra"},
			wantStderr: "ERROR: invalid data format, expected exactly one input file\n",
			exact:      true,
		},
		{
			name:       "missing file",
			args:       []string{missingFile},
			wantStderr: "ERROR: invalid data format,",
		},
		{
			name:       "empty file",
			args:       []string{emptyFile},
			wantStderr: "ERROR: invalid data format,",
		},
		{
			name:       "disconnected farm",
			args:       []string{disconnectedFile},
			wantStderr: "ERROR: invalid data format,",
		},
		{
			name:       "bad example 00",
			args:       []string{"examples/badexample00.txt"},
			wantStderr: "ERROR: invalid data format,",
		},
		{
			name:       "bad example 01",
			args:       []string{"examples/badexample01.txt"},
			wantStderr: "ERROR: invalid data format,",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cmd := exec.Command(binary, tc.args...)
			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			err := cmd.Run()
			var exitErr *exec.ExitError
			if !errors.As(err, &exitErr) {
				t.Fatalf("CLI error = %v, want nonzero exit", err)
			}
			if exitErr.ExitCode() != 1 {
				t.Errorf("CLI exit code = %d, want 1", exitErr.ExitCode())
			}
			if stdout.Len() != 0 {
				t.Errorf("CLI wrote stdout on failure: %q", stdout.String())
			}
			if tc.exact && stderr.String() != tc.wantStderr {
				t.Errorf("CLI stderr = %q, want %q", stderr.String(), tc.wantStderr)
			}
			if !tc.exact && !strings.HasPrefix(stderr.String(), tc.wantStderr) {
				t.Errorf("CLI stderr = %q, want prefix %q", stderr.String(), tc.wantStderr)
			}
		})
	}
}

func buildCLI(t *testing.T) string {
	t.Helper()

	binary := filepath.Join(t.TempDir(), "lem-in")
	cmd := exec.Command("go", "build", "-o", binary, ".")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build CLI: %v\n%s", err, output)
	}
	return binary
}

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

			g, err := parser.Parse(tc.file)
			if err != nil {
				t.Fatalf("parse official example for audit checks: %v", err)
			}
			verifyMovementRules(t, g, strings.Split(moves, "\n"))
		})
	}
}

// verifyMovementRules independently checks the movement transcript emitted by
// the public CLI seam. It deliberately does not rely on the paths selected by
// the graph package: the parsed farm and the official movement rules are the
// source of truth.
func verifyMovementRules(t *testing.T, g *graph.Graph, turns []string) {
	t.Helper()

	positions := make(map[int]string, g.NumAnts)
	for ant := 1; ant <= g.NumAnts; ant++ {
		positions[ant] = g.Start
	}

	for turnIndex, line := range turns {
		if strings.TrimSpace(line) == "" {
			t.Fatalf("turn %d contains no movements", turnIndex+1)
		}
		if normalized := strings.Join(strings.Fields(line), " "); normalized != line {
			t.Fatalf("turn %d is not single-space separated: %q", turnIndex+1, line)
		}

		moved := make(map[int]string)
		destinations := make(map[string]int)
		tunnels := make(map[[2]string]int)

		for _, token := range strings.Fields(line) {
			ant, destination := parseAuditMove(t, token)
			if ant < 1 || ant > g.NumAnts {
				t.Fatalf("turn %d: ant %d is outside 1..%d", turnIndex+1, ant, g.NumAnts)
			}
			if _, duplicate := moved[ant]; duplicate {
				t.Fatalf("turn %d: ant %d moves more than once", turnIndex+1, ant)
			}

			from := positions[ant]
			if from == g.End {
				t.Fatalf("turn %d: ant %d moves after reaching the end", turnIndex+1, ant)
			}
			if !slices.Contains(g.Links[from], destination) {
				t.Fatalf("turn %d: ant %d uses nonexistent tunnel %q-%q",
					turnIndex+1, ant, from, destination)
			}

			tunnel := [2]string{from, destination}
			if tunnel[1] < tunnel[0] {
				tunnel[0], tunnel[1] = tunnel[1], tunnel[0]
			}
			if other, used := tunnels[tunnel]; used {
				t.Fatalf("turn %d: ants %d and %d both use tunnel %q-%q",
					turnIndex+1, other, ant, tunnel[0], tunnel[1])
			}
			tunnels[tunnel] = ant

			if destination != g.Start && destination != g.End {
				if other, occupied := destinations[destination]; occupied {
					t.Fatalf("turn %d: ants %d and %d both enter room %q",
						turnIndex+1, other, ant, destination)
				}
				destinations[destination] = ant
			}
			moved[ant] = destination
		}

		// A room may be entered as its current occupant leaves during the same
		// turn, but it may not be entered while that occupant stays put.
		for ant, room := range positions {
			if room == g.Start || room == g.End {
				continue
			}
			if _, leaves := moved[ant]; leaves {
				continue
			}
			if entering, collision := destinations[room]; collision {
				t.Fatalf("turn %d: ant %d enters room %q while ant %d remains there",
					turnIndex+1, entering, room, ant)
			}
		}

		for ant, room := range moved {
			positions[ant] = room
		}
	}

	for ant := 1; ant <= g.NumAnts; ant++ {
		if positions[ant] != g.End {
			t.Errorf("ant %d finishes in %q, want end room %q", ant, positions[ant], g.End)
		}
	}
}

func TestRunIsDeterministic(t *testing.T) {
	const input = "examples/example05.txt"

	var first bytes.Buffer
	if err := run([]string{input}, &first); err != nil {
		t.Fatalf("first run: %v", err)
	}

	for attempt := 1; attempt <= 20; attempt++ {
		var again bytes.Buffer
		if err := run([]string{input}, &again); err != nil {
			t.Fatalf("attempt %d: %v", attempt, err)
		}
		if got, want := again.String(), first.String(); got != want {
			t.Fatalf("attempt %d produced different output", attempt)
		}
	}
}

func parseAuditMove(t *testing.T, token string) (int, string) {
	t.Helper()

	body, ok := strings.CutPrefix(token, "L")
	if !ok {
		t.Fatalf("movement %q does not start with L", token)
	}
	antText, room, ok := strings.Cut(body, "-")
	if !ok || antText == "" || room == "" {
		t.Fatalf("movement %q is not formatted as Lx-y", token)
	}
	ant, err := strconv.Atoi(antText)
	if err != nil {
		t.Fatalf("movement %q has an invalid ant number: %v", token, err)
	}
	return ant, room
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
