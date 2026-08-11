package graph

import (
	"slices"
	"testing"
)

// The four node helpers have to be exact inverses. If they ever drift apart,
// every path this package produces turns into nonsense, and the end-to-end
// tests could only report "wrong answer" without saying where.
func TestNodeNumbersAreInverses(t *testing.T) {
	for room := 0; room < 100; room++ {
		if got := roomOf(entryNode(room)); got != room {
			t.Fatalf("roomOf(entryNode(%d)) = %d", room, got)
		}
		if got := roomOf(exitNode(room)); got != room {
			t.Fatalf("roomOf(exitNode(%d)) = %d", room, got)
		}
		if !isEntry(entryNode(room)) {
			t.Fatalf("entryNode(%d) is not reported as an entry", room)
		}
		if isEntry(exitNode(room)) {
			t.Fatalf("exitNode(%d) is reported as an entry", room)
		}
	}
}

func TestNodeNumbersNeverCollide(t *testing.T) {
	taken := make(map[int]int) // node number -> the room that claimed it

	for room := 0; room < 100; room++ {
		for _, node := range []int{entryNode(room), exitNode(room)} {
			if other, clash := taken[node]; clash {
				t.Fatalf("node %d claimed by both room %d and room %d", node, other, room)
			}
			taken[node] = room
		}
	}
}

// Go randomises map iteration, so this sort is the one thing standing between
// us and a different answer on every run.
func TestSortedRoomNamesIsStable(t *testing.T) {
	g := buildFarm(t, 1, "S", "E", "S-b", "b-E", "S-a", "a-E")
	want := []string{"E", "S", "a", "b"}

	for attempt := 0; attempt < 20; attempt++ {
		if got := sortedRoomNames(g); !slices.Equal(got, want) {
			t.Fatalf("attempt %d: got %v, want %v", attempt, got, want)
		}
	}
}

func TestNodeCountIsTwoPerRoom(t *testing.T) {
	g := buildFarm(t, 1, "S", "E", "S-a", "a-E") // three rooms

	if got := newNetwork(g).nodeCount(); got != 6 {
		t.Errorf("got %d nodes for 3 rooms, want 6", got)
	}
}

// The search starts past the start room's turnstile and stops before the end
// room's. That is what leaves those two rooms unlimited, with no special case.
func TestNewNetworkPlacesSourceAndSink(t *testing.T) {
	g := buildFarm(t, 1, "S", "E", "S-a", "a-E")
	n := newNetwork(g)

	// The rooms sort to E, S, a, so they are numbered 0, 1, 2.
	if got, want := n.source, exitNode(1); got != want {
		t.Errorf("source = %d, want the exit of S (%d)", got, want)
	}
	if got, want := n.sink, entryNode(0); got != want {
		t.Errorf("sink = %d, want the entry of E (%d)", got, want)
	}
}

func TestEveryRoomGetsATurnstile(t *testing.T) {
	g := buildFarm(t, 1, "S", "E", "S-a", "a-E")
	n := newNetwork(g)

	for room, name := range n.roomNames {
		found := false
		for _, id := range n.leaving[entryNode(room)] {
			if n.edges[id].to == exitNode(room) {
				found = true
			}
		}
		if !found {
			t.Errorf("room %q has no passage between its two halves", name)
		}
	}
}

func TestAddEdgeMakesAPair(t *testing.T) {
	n := &network{leaving: make([][]int, 2)}
	n.addEdge(0, 1)

	if len(n.edges) != 2 {
		t.Fatalf("got %d edges, want a pair", len(n.edges))
	}
	onward, undo := n.edges[0], n.edges[1]

	if onward.from != 0 || onward.to != 1 {
		t.Errorf("onward runs %d -> %d, want 0 -> 1", onward.from, onward.to)
	}
	if undo.from != 1 || undo.to != 0 {
		t.Errorf("undo runs %d -> %d, want 1 -> 0", undo.from, undo.to)
	}

	if !onward.open {
		t.Error("a fresh passage should have room in it")
	}
	if undo.open {
		t.Error("an undo passage opens only once a booking is made")
	}
	if onward.isUndo || !undo.isUndo {
		t.Error("exactly one of the pair should be marked as the undo")
	}
	if onward.reverse != 1 || undo.reverse != 0 {
		t.Errorf("the pair does not point back at each other: %d and %d",
			onward.reverse, undo.reverse)
	}
}

func TestEdgeInUse(t *testing.T) {
	tests := []struct {
		name string
		e    edge
		want bool
	}{
		{"a passage with room left carries nobody", edge{open: true}, false},
		{"a full passage is carrying an ant", edge{open: false}, true},
		{"an undo passage never carries an ant", edge{isUndo: true, open: false}, false},
	}

	for _, tc := range tests {
		if got := tc.e.inUse(); got != tc.want {
			t.Errorf("%s: inUse() = %v, want %v", tc.name, got, tc.want)
		}
	}
}
