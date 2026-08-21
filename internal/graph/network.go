package graph

import "slices"

// edge is a one-way passage between two nodes. Every real passage fits exactly
// one ant, so "is there room left" is a yes or no question rather than a count.
type edge struct {
	from, to int
	open     bool // still has room for an ant
	isUndo   bool // exists only to cancel a booking; never carries an ant itself
	reverse  int  // the id of the passage that cancels this one
}

// inUse reports whether an ant is currently booked through this passage.
func (e edge) inUse() bool { return !e.isUndo && !e.open }

// network rewrites room capacity as edge capacity by splitting every room into
// entry and exit nodes joined by a one-ant passage. The search starts at the
// start room's exit and stops at the end room's entry, leaving both unlimited.
type network struct {
	roomNames []string       // room number -> room name
	roomIndex map[string]int // room name -> room number

	edges   []edge
	leaving [][]int // leaving[v] holds the ids of the passages out of node v

	source int // the start room's exit
	sink   int // the end room's entry
}

// nodeCount is how many entry and exit halves the farm was split into.
func (n *network) nodeCount() int { return len(n.leaving) }

// Room r owns nodes 2r and 2r+1. These four are exact inverses of each other,
// so a node number on its own says everything about where it sits.
func entryNode(room int) int { return 2 * room }
func exitNode(room int) int  { return 2*room + 1 }
func roomOf(node int) int    { return node / 2 }
func isEntry(node int) bool  { return node%2 == 0 }

// newNetwork builds the split-room model of a parsed ant farm.
func newNetwork(g *Graph) *network {
	names := sortedRoomNames(g)

	n := &network{
		roomNames: names,
		roomIndex: make(map[string]int, len(names)),
		leaving:   make([][]int, 2*len(names)),
	}

	// Every room has to be numbered before any tunnel is read, since a tunnel
	// can name a room that comes later in the list.
	for room, name := range names {
		n.roomIndex[name] = room
	}

	for room, name := range names {
		n.addTurnstile(room)
		n.addTunnelsFrom(room, g.Links[name])
	}

	n.source = exitNode(n.roomIndex[g.Start])
	n.sink = entryNode(n.roomIndex[g.End])
	return n
}

// sortedRoomNames lists every room in a fixed order.
//
// Go walks maps in a random order, and the audit requires that the same input
// always produces the same output. Sorting is what pins that down.
func sortedRoomNames(g *Graph) []string {
	names := make([]string, 0, len(g.Rooms))
	for name := range g.Rooms {
		names = append(names, name)
	}
	slices.Sort(names)
	return names
}

// addTurnstile joins a room's two halves with a passage that fits one ant. This
// single line is the entire one-ant-per-room rule.
func (n *network) addTurnstile(room int) {
	n.addEdge(entryNode(room), exitNode(room))
}

// addTunnelsFrom opens a passage from one room's exit to each neighbour's entry.
// Every tunnel is listed under both of its rooms, so walking the rooms in turn
// produces both directions of every tunnel exactly once.
func (n *network) addTunnelsFrom(room int, neighbours []string) {
	for _, neighbour := range neighbours {
		if other, known := n.roomIndex[neighbour]; known {
			n.addEdge(exitNode(room), entryNode(other))
		}
	}
}

// addEdge opens a passage, and behind it a closed undo passage. Sending an ant
// forwards is what opens the undo, which is how a later search can walk back
// through a booking, take it back, and reroute the ant that made it.
func (n *network) addEdge(from, to int) {
	// Ids are handed out in the order passages are appended, so the next two
	// land on the current length and the one after it.
	onward := len(n.edges)
	undo := onward + 1

	n.edges = append(n.edges,
		edge{from: from, to: to, open: true, reverse: undo},
		edge{from: to, to: from, isUndo: true, reverse: onward},
	)

	n.leaving[from] = append(n.leaving[from], onward)
	n.leaving[to] = append(n.leaving[to], undo)
}
