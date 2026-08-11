package graph

// augment finds one more route from start to end that still has room, books it,
// and reports whether it found one.
//
// Searching and booking are kept apart on purpose: bfs only reads, so a search
// that finds nothing leaves the network untouched and there is nothing to undo.
func (n *network) augment() bool {
	arrivedBy, reached := n.bfs()
	if !reached {
		return false
	}
	n.pushFlow(arrivedBy)
	return true
}

// bfs searches breadth-first for a route from source to sink along passages that
// still have room. It returns the passage each node was reached by, and whether
// the sink was reached at all.
//
// Breadth-first rather than depth-first on purpose: it always returns the
// shortest route left, and path length is exactly what decides the turn count.
// Depth-first would find just as many paths, but slower ones.
func (n *network) bfs() ([]int, bool) {
	nodes := n.nodeCount()
	visited := make([]bool, nodes)
	arrivedBy := make([]int, nodes)

	queue := []int{n.source}
	visited[n.source] = true

	// Stop as soon as the sink is reached: breadth-first guarantees the route
	// found at that moment is the shortest one, so looking further cannot help.
	for len(queue) > 0 && !visited[n.sink] {
		current := queue[0]
		queue = queue[1:]

		for _, id := range n.leaving[current] {
			next := n.edges[id].to
			if !n.edges[id].open || visited[next] {
				continue // full, or somewhere we have already been
			}
			visited[next] = true
			arrivedBy[next] = id
			queue = append(queue, next)
		}
	}
	return arrivedBy, visited[n.sink]
}

// pushFlow sends one ant down the route bfs just found.
//
// arrivedBy records only how each node was reached, so the route can be read
// from the end backwards and no other way. Every step closes the passage it
// walks and opens the undo passage behind it, and that second line is what lets
// a later search take this booking back and reroute the ant.
func (n *network) pushFlow(arrivedBy []int) {
	current := n.sink

	for current != n.source {
		id := arrivedBy[current]    // the passage that reached this node
		undo := n.edges[id].reverse // the one that cancels it

		n.edges[id].open = false
		n.edges[undo].open = true

		current = n.edges[id].from
	}
}
