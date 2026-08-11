package graph

// decompose reads the booked passages back out as whole paths, written in room
// names so nothing outside this package has to know rooms were ever split.
func (n *network) decompose(startName string) []Path {
	walked := make([]bool, len(n.edges))
	var paths []Path

	for {
		path, ok := n.walkOneRoute(startName, walked)
		if !ok {
			return paths
		}
		paths = append(paths, path)
	}
}

// walkOneRoute follows booked passages from the start room until it reaches the
// end room, ticking off every passage it uses so the next walk takes a different
// route. It reports false once no complete route is left.
//
// It can never get stuck halfway. A room's middle passage fits one ant, so a
// booked one means exactly one ant went in and exactly one came out: wherever
// the walk lands, there is always a booked passage leading onwards.
func (n *network) walkOneRoute(startName string, walked []bool) (Path, bool) {
	// The walk begins at the start room's exit, so the start room is never
	// entered and has to be written down by hand. The end room needs no such
	// help: the walk finishes on its entry, which the check below picks up.
	path := Path{startName}
	current := n.source

	for current != n.sink {
		id := n.nextFlowEdge(current, walked)
		if id < 0 {
			return nil, false
		}
		walked[id] = true
		current = n.edges[id].to

		// Nodes alternate entry, exit, entry, exit. Landing on an entry means
		// stepping into a new room; landing on an exit is the same room again,
		// just past its middle passage.
		if isEntry(current) {
			path = append(path, n.roomNames[roomOf(current)])
		}
	}
	return path, true
}

// nextFlowEdge returns the next passage out of a node that is carrying an ant
// and has not been walked yet, or -1 when there is none left.
func (n *network) nextFlowEdge(from int, walked []bool) int {
	for _, id := range n.leaving[from] {
		if n.edges[id].inUse() && !walked[id] {
			return id
		}
	}
	return -1
}
