package simulation

import (
	"slices"
	"strconv"
	"strings"
	"testing"

	"lem-in/internal/graph"
	"lem-in/internal/parser"
)

func TestSimulateRejectsNoPaths(t *testing.T) {
	if _, err := Simulate(nil, 3); err == nil {
		t.Fatal("expected an error when there are no paths")
	}
}

func TestSimulateRejectsNoAnts(t *testing.T) {
	paths := []graph.Path{{"S", "E"}}
	if _, err := Simulate(paths, 0); err == nil {
		t.Fatal("expected an error when there are no ants")
	}
}

// The brief's worked example (PRD section 6.1): a single corridor, four ants
// queuing behind each other through three tunnels, six turns.
func TestSimulateMatchesTheBriefsWorkedExample(t *testing.T) {
	paths := []graph.Path{{"0", "2", "3", "1"}}

	got, err := Simulate(paths, 4)
	if err != nil {
		t.Fatalf("Simulate: %v", err)
	}

	want := []string{
		"L1-2",
		"L1-3 L2-2",
		"L1-1 L2-3 L3-2",
		"L2-1 L3-3 L4-2",
		"L3-1 L4-3",
		"L4-1",
	}
	assertLines(t, got, want)
}

// Same worked example distribute_test.go checks the counts for: a
// three-tunnel path and a five-tunnel one, six ants split 4/2 by AssignAnts.
func TestSimulateSplitsAntsAcrossTwoUnequalPaths(t *testing.T) {
	paths := []graph.Path{
		{"S", "a", "b", "E"},
		{"S", "c", "d", "e", "f", "E"},
	}

	got, err := Simulate(paths, 6)
	if err != nil {
		t.Fatalf("Simulate: %v", err)
	}

	want := []string{
		"L1-a L4-c",
		"L1-b L2-a L4-d L6-c",
		"L1-E L2-b L3-a L4-e L6-d",
		"L2-E L3-b L4-f L5-a L6-e",
		"L3-E L4-E L5-b L6-f",
		"L5-E L6-E",
	}
	assertLines(t, got, want)
}

// One short path carrying every ant: the classic single-corridor queue.
func TestSimulateOneShortPathCarriesAllAnts(t *testing.T) {
	paths := []graph.Path{{"S", "m", "E"}}

	got, err := Simulate(paths, 3)
	if err != nil {
		t.Fatalf("Simulate: %v", err)
	}

	want := []string{
		"L1-m",
		"L1-E L2-m",
		"L2-E L3-m",
		"L3-E",
	}
	assertLines(t, got, want)
}

func TestSimulateLeavesAnUnhelpfulPathIdle(t *testing.T) {
	paths := []graph.Path{
		{"S", "a", "E"},
		{"S", "b", "c", "d", "e", "f", "E"},
	}

	got, err := Simulate(paths, 2)
	if err != nil {
		t.Fatalf("Simulate: %v", err)
	}

	want := []string{
		"L1-a",
		"L1-E L2-a",
		"L2-E",
	}
	assertLines(t, got, want)
}

func TestSimulateIsDeterministic(t *testing.T) {
	paths := []graph.Path{
		{"S", "a", "b", "E"},
		{"S", "c", "d", "e", "f", "E"},
	}

	first, err := Simulate(paths, 6)
	if err != nil {
		t.Fatalf("Simulate: %v", err)
	}

	for attempt := 0; attempt < 20; attempt++ {
		again, err := Simulate(paths, 6)
		if err != nil {
			t.Fatalf("attempt %d: %v", attempt, err)
		}
		if !slices.Equal(first, again) {
			t.Fatalf("attempt %d differs:\nfirst %v\nagain %v", attempt, first, again)
		}
	}
}

// End-to-end through the real pipeline (parser -> graph -> simulation) on the
// example fixtures, checking the audit's core simulation properties rather
// than exact output: every ant reaches an end room, no ant moves twice in a
// turn, and no two ants enter the same non-terminal room in the same turn.
func TestSimulateOnExampleFixtures(t *testing.T) {
	tests := []struct {
		file     string
		maxTurns int
	}{
		{"../../examples/example00.txt", 6}, // PRD: must finish in six turns
		{"../../examples/detour.txt", 6},    // hand-solved: the detour is worth it, in six
	}

	for _, tc := range tests {
		t.Run(tc.file, func(t *testing.T) {
			g, err := parser.Parse(tc.file)
			if err != nil {
				t.Fatalf("Parse: %v", err)
			}
			paths, err := graph.FindPaths(g)
			if err != nil {
				t.Fatalf("FindPaths: %v", err)
			}
			lines, err := Simulate(paths, g.NumAnts)
			if err != nil {
				t.Fatalf("Simulate: %v", err)
			}

			if len(lines) > tc.maxTurns {
				t.Errorf("took %d turns, want at most %d:\n%s",
					len(lines), tc.maxTurns, strings.Join(lines, "\n"))
			}
			checkInvariants(t, paths, g.NumAnts, lines)
		})
	}
}

func assertLines(t *testing.T, got, want []string) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("got %d turns, want %d\ngot:  %v\nwant: %v", len(got), len(want), got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("turn %d: got %q, want %q", i+1, got[i], want[i])
		}
	}
}

// checkInvariants verifies the audit's core simulation properties: no ant
// moves twice in one turn, no two ants enter the same non-terminal room in one
// turn, and every ant's last move lands it in an end room.
func checkInvariants(t *testing.T, paths []graph.Path, numAnts int, lines []string) {
	t.Helper()

	ends := make(map[string]bool)
	lastRoom := make(map[int]string)
	for _, p := range paths {
		ends[p[len(p)-1]] = true
	}
	for p, queue := range graph.AssignAnts(paths, numAnts) {
		for _, ant := range queue {
			lastRoom[ant] = paths[p][0]
		}
	}

	for turn, line := range lines {
		seenAnt := make(map[int]bool)
		seenRoom := make(map[string]bool)
		seenTunnel := make(map[string]bool)

		for _, token := range strings.Fields(line) {
			ant, room := parseMove(t, token)

			if seenAnt[ant] {
				t.Fatalf("turn %d: ant %d moves twice: %q", turn+1, ant, line)
			}
			seenAnt[ant] = true

			if !ends[room] && seenRoom[room] {
				t.Fatalf("turn %d: two ants enter room %q: %q", turn+1, room, line)
			}
			seenRoom[room] = true

			from, to := lastRoom[ant], room
			if from > to {
				from, to = to, from
			}
			tunnel := from + "\x00" + to
			if seenTunnel[tunnel] {
				t.Fatalf("turn %d: tunnel %q-%q is used twice: %q", turn+1, from, to, line)
			}
			seenTunnel[tunnel] = true

			lastRoom[ant] = room
		}
	}

	for ant := 1; ant <= numAnts; ant++ {
		room, moved := lastRoom[ant]
		if !moved {
			t.Errorf("ant %d never moved", ant)
			continue
		}
		if !ends[room] {
			t.Errorf("ant %d's last move was into %q, not an end room", ant, room)
		}
	}
}

// parseMove splits an "Lant-room" token into its ant number and room name.
func parseMove(t *testing.T, token string) (int, string) {
	t.Helper()

	body, ok := strings.CutPrefix(token, "L")
	if !ok {
		t.Fatalf("move %q does not start with L", token)
	}
	antStr, room, ok := strings.Cut(body, "-")
	if !ok {
		t.Fatalf("move %q is not of the form Lant-room", token)
	}
	ant, err := strconv.Atoi(antStr)
	if err != nil {
		t.Fatalf("move %q has a non-numeric ant: %v", token, err)
	}
	return ant, room
}
