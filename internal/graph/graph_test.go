package graph

import (
	"slices"
	"strings"
	"testing"
)

// buildFarm makes a Graph the way the parser would, from tunnel lines written
// exactly as they appear in an input file.
func buildFarm(t *testing.T, ants int, start, end string, tunnels ...string) *Graph {
	t.Helper()

	g := &Graph{
		Rooms:   make(map[string]*Room),
		Links:   make(map[string][]string),
		NumAnts: ants,
		Start:   start,
		End:     end,
	}

	for _, tunnel := range tunnels {
		a, b, ok := strings.Cut(tunnel, "-")
		if !ok {
			t.Fatalf("malformed tunnel in fixture: %q", tunnel)
		}
		for _, name := range []string{a, b} {
			if _, seen := g.Rooms[name]; !seen {
				g.Rooms[name] = &Room{Name: name}
			}
		}
		g.Links[a] = append(g.Links[a], b)
		g.Links[b] = append(g.Links[b], a)
	}
	return g
}

// checkPathSet asserts what the audit cares about: every path is a real walk
// from start to end, and no two paths share a room in the middle.
func checkPathSet(t *testing.T, g *Graph, paths []Path) {
	t.Helper()

	holder := make(map[string]int) // room -> the path already using it

	for i, p := range paths {
		if len(p) < 2 {
			t.Fatalf("path %d is too short: %v", i, p)
		}
		if p[0] != g.Start || p[len(p)-1] != g.End {
			t.Fatalf("path %d does not run %s -> %s: %v", i, g.Start, g.End, p)
		}

		for step := 1; step < len(p); step++ {
			if !slices.Contains(g.Links[p[step-1]], p[step]) {
				t.Fatalf("path %d walks a tunnel that does not exist: %s-%s",
					i, p[step-1], p[step])
			}
		}

		for _, room := range p[1 : len(p)-1] {
			if other, taken := holder[room]; taken {
				t.Fatalf("paths %d and %d both use room %q", other, i, room)
			}
			holder[room] = i
		}
	}
}

// checkResult runs FindPaths and asserts the path count and the turn count.
//
// It deliberately does not assert which paths came back. Ties are broken by
// room order, so an equally good set in a different order is still correct.
func checkResult(t *testing.T, g *Graph, wantPaths, wantTurns int) {
	t.Helper()

	paths, err := FindPaths(g)
	if err != nil {
		t.Fatalf("FindPaths: %v", err)
	}
	checkPathSet(t, g, paths)

	if len(paths) != wantPaths {
		t.Errorf("got %d paths, want %d: %v", len(paths), wantPaths, paths)
	}
	if got := turnsNeeded(paths, g.NumAnts); got != wantTurns {
		t.Errorf("got %d turns, want %d: %v", got, wantTurns, paths)
	}
}

// example00 from the project brief: one corridor, so every ant queues behind the
// one in front. Three tunnels and four ants take 3+4-1 turns.
func TestFindPathsSingleCorridor(t *testing.T) {
	g := buildFarm(t, 4, "0", "1", "0-2", "2-3", "3-1")
	checkResult(t, g, 1, 6)
}

// Taking the shortest path and deleting its rooms strands the rest of this farm:
// it finds only S-1-2-E and needs 6 turns. Undoing that first choice is what
// frees both paths and brings the ants home in 4.
func TestFindPathsUndoesABlockingChoice(t *testing.T) {
	g := buildFarm(t, 4, "S", "E",
		"S-1", "1-2", "2-E", "S-3", "3-2", "1-4", "4-E")
	checkResult(t, g, 2, 4)
}

// Two paths of different lengths, three tunnels and four, with five ants split
// three and two.
func TestFindPathsTwoUnequalPaths(t *testing.T) {
	g := buildFarm(t, 5, "S", "E",
		"S-a", "a-b", "b-E", "S-c", "c-b",
		"a-d", "d-g", "g-E", "b-h")
	checkResult(t, g, 2, 5)
}

// The long way round costs five tunnels against three, but with six ants the
// short path would jam, so both are worth using.
func TestFindPathsLongDetourStillWorthTaking(t *testing.T) {
	g := buildFarm(t, 6, "S", "E",
		"S-no", "S-so", "no-mid", "mid-E", "so-mid",
		"no-up", "up-far", "far-out", "out-E", "mid-dead")
	checkResult(t, g, 2, 6)
}

// Two ants, a two-tunnel path and a six-tunnel detour. Sending anyone down the
// detour would land them on turn 6, so the right answer is to leave it alone.
func TestFindPathsIgnoresAPathThatWouldSlowThingsDown(t *testing.T) {
	g := buildFarm(t, 2, "S", "E",
		"S-a", "a-E",
		"S-b1", "b1-b2", "b2-b3", "b3-b4", "b4-b5", "b5-E")
	checkResult(t, g, 1, 3)
}

func TestFindPathsRejectsDisconnectedFarm(t *testing.T) {
	g := buildFarm(t, 3, "S", "E", "S-a", "b-E")

	if _, err := FindPaths(g); err == nil {
		t.Fatal("expected an error when start and end are not connected")
	}
}

// Go randomises map iteration, so without a fixed room order the same farm could
// produce different output on every run. The audit fails that outright.
func TestFindPathsIsDeterministic(t *testing.T) {
	farm := func() *Graph {
		return buildFarm(t, 6, "S", "E",
			"S-no", "S-so", "no-mid", "mid-E", "so-mid",
			"no-up", "up-far", "far-out", "out-E", "mid-dead")
	}

	first, err := FindPaths(farm())
	if err != nil {
		t.Fatalf("FindPaths: %v", err)
	}

	for attempt := 0; attempt < 50; attempt++ {
		again, err := FindPaths(farm())
		if err != nil {
			t.Fatalf("attempt %d: %v", attempt, err)
		}
		samePaths := func(a, b Path) bool { return slices.Equal(a, b) }
		if !slices.EqualFunc(first, again, samePaths) {
			t.Fatalf("attempt %d differs:\nfirst %v\nagain %v", attempt, first, again)
		}
	}
}
