package graph

import (
	"slices"
	"testing"
)

// pathOfLength makes a path with the given number of tunnels. Nothing in
// distribute.go looks at room names, only at how many rooms a path lists, so
// filler names are enough and make that fact obvious.
func pathOfLength(tunnels int) Path {
	rooms := make(Path, tunnels+1)
	for i := range rooms {
		rooms[i] = "room"
	}
	return rooms
}

func TestTunnelCount(t *testing.T) {
	// Four rooms in a row are joined by three tunnels.
	if got := tunnelCount(Path{"S", "a", "b", "E"}); got != 3 {
		t.Errorf("got %d tunnels, want 3", got)
	}
}

func TestArrivalTurn(t *testing.T) {
	three := pathOfLength(3)

	tests := []struct {
		queuedAhead int
		want        int
	}{
		{0, 3}, // nobody ahead: straight through, one turn per tunnel
		{1, 4}, // one ant ahead: one turn of waiting
		{4, 7},
	}

	for _, tc := range tests {
		if got := arrivalTurn(three, tc.queuedAhead); got != tc.want {
			t.Errorf("arrivalTurn(L=3, %d ahead) = %d, want %d",
				tc.queuedAhead, got, tc.want)
		}
	}
}

// The farm worked out by hand: a three-tunnel path and a five-tunnel one, six
// ants. The short path fills up until the long one becomes the faster bet.
func TestAntsPerPathMatchesTheWorkedExample(t *testing.T) {
	paths := []Path{pathOfLength(3), pathOfLength(5)}

	got := antsPerPath(paths, 6)
	want := []int{4, 2}

	if !slices.Equal(got, want) {
		t.Errorf("got %v ants per path, want %v", got, want)
	}
}

func TestAntsPerPathSpreadsEvenlyOverEqualPaths(t *testing.T) {
	paths := []Path{pathOfLength(4), pathOfLength(4), pathOfLength(4)}

	got := antsPerPath(paths, 9)
	want := []int{3, 3, 3}

	if !slices.Equal(got, want) {
		t.Errorf("got %v ants per path, want %v", got, want)
	}
}

// Two ants, a two-tunnel path and a fifty-tunnel detour. Nobody should ever be
// sent the long way: they would arrive on turn 50 instead of 3.
func TestAntsPerPathLeavesASlowPathEmpty(t *testing.T) {
	paths := []Path{pathOfLength(2), pathOfLength(50)}

	got := antsPerPath(paths, 2)
	want := []int{2, 0}

	if !slices.Equal(got, want) {
		t.Errorf("got %v ants per path, want %v", got, want)
	}
}

// Same worked example as TestAntsPerPathMatchesTheWorkedExample, but checking
// which ant numbers land where, not just how many.
func TestAssignAntsMatchesTheWorkedExample(t *testing.T) {
	paths := []Path{pathOfLength(3), pathOfLength(5)}

	got := AssignAnts(paths, 6)
	want := [][]int{{1, 2, 3, 5}, {4, 6}}

	for p := range want {
		if !slices.Equal(got[p], want[p]) {
			t.Errorf("path %d: got %v, want %v", p, got[p], want[p])
		}
	}
}

// A path that carries nobody must still show up as an (empty) queue, so
// callers can index it by path position.
func TestAssignAntsLeavesASlowPathEmpty(t *testing.T) {
	paths := []Path{pathOfLength(2), pathOfLength(50)}

	got := AssignAnts(paths, 2)
	want := [][]int{{1, 2}, {}}

	for p := range want {
		if !slices.Equal(got[p], want[p]) {
			t.Errorf("path %d: got %v, want %v", p, got[p], want[p])
		}
	}
}

// Every ant assigned must appear exactly once across all paths, in ascending
// order overall, since ants are handed out 1..N in that order.
func TestAssignAntsCoversEveryAntExactlyOnce(t *testing.T) {
	paths := []Path{pathOfLength(4), pathOfLength(4), pathOfLength(4)}

	queues := AssignAnts(paths, 9)

	var all []int
	for _, q := range queues {
		all = append(all, q...)
	}
	slices.Sort(all)

	want := []int{1, 2, 3, 4, 5, 6, 7, 8, 9}
	if !slices.Equal(all, want) {
		t.Errorf("got ants %v, want %v", all, want)
	}
}

// An empty path set must return an empty (not nil-panicking) set of queues.
func TestAssignAntsOnNoPaths(t *testing.T) {
	if got := AssignAnts(nil, 5); len(got) != 0 {
		t.Errorf("got %v, want no queues", got)
	}
}

func TestTurnsNeededOnOneCorridor(t *testing.T) {
	// Four ants queue behind each other down three tunnels: 3 + 4 - 1.
	if got := turnsNeeded([]Path{pathOfLength(3)}, 4); got != 6 {
		t.Errorf("got %d turns, want 6", got)
	}
}

// The fifty-tunnel path carries nobody, so it must not count towards the total.
// Without the guard on empty paths this would report 49.
func TestTurnsNeededIgnoresAnEmptyPath(t *testing.T) {
	paths := []Path{pathOfLength(2), pathOfLength(50)}

	if got := turnsNeeded(paths, 2); got != 3 {
		t.Errorf("got %d turns, want 3", got)
	}
}

// An empty set must not reach antsPerPath, which indexes paths[0].
func TestTurnsNeededOnNoPaths(t *testing.T) {
	if got := turnsNeeded(nil, 5); got != 0 {
		t.Errorf("got %d turns, want 0", got)
	}
}
